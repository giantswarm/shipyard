package awsminikube

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/shipyard/pkg/engine"
	"golang.org/x/crypto/ssh"
)

const (
	templateBody = `
AWSTemplateFormatVersion: "2010-09-09"
Description: shipyard minikube template
Resources:
  SMInstance:
    Type: "AWS::EC2::Instance"
    Properties:
      ImageId:
        Ref: ImageIDParameter
      InstanceType: t2.medium
      KeyName:
        Ref: KeyPairNameParameter
      SecurityGroupIds:
      - Ref: SMSecurityGroup
      Tags:
      - Key: Name
        Value:
          Ref: InstanceNameParameter

  SMSecurityGroup:
    Type: "AWS::EC2::SecurityGroup"
    Properties:
      GroupDescription: Allow ssh to client host
      SecurityGroupIngress:
      - IpProtocol: tcp
        FromPort: '22'
        ToPort: '22'
        CidrIp: 0.0.0.0/0

Parameters:
  InstanceNameParameter:
    Type: String
    Description: Name of the instance
  ImageIDParameter:
    Type: String
    Description: Image ID to be used for the shipyard minikube instance
  KeyPairNameParameter:
    Type: String
    Description: Name of the keypair

Outputs:
  InstanceIP:
    Description: The Instance IP
    Value: !GetAtt SMInstance.PublicIp
`
)

type clients struct {
	CloudFormation *cloudformation.CloudFormation
	EC2            *ec2.EC2
}

type Engine struct {
	name    string
	clients *clients
	logger  micrologger.Logger
	result  *result
}

type result struct {
	imageID           string
	stackID           string
	instanceIP        string
	privateKeyContent string
	caCrtContent      string
	clientCrtContent  string
	clientKeyContent  string
	kubeconfigContent string
}

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	SessionToken    string
	Region          string
}

type stage func(*result) (*result, error)

func newClients(config *Config) *clients {
	awsCfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.AccessKeySecret, config.SessionToken),
		Region:      aws.String(config.Region),
	}
	s := session.New(awsCfg)

	return &clients{
		CloudFormation: cloudformation.New(s),
		EC2:            ec2.New(s),
	}
}

func DefaultConfig() *Config {
	return &Config{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
		Region:          "eu-central-1",
	}
}

func New(name string, config *Config, logger micrologger.Logger) *Engine {
	return &Engine{
		name:    name,
		clients: newClients(config),
		logger:  logger,
	}
}

func (e *Engine) SetUp() (*engine.Result, error) {
	stages := []stage{
		e.findImageID,
		e.createKeyPair,
		e.createStack,
		e.waitForSSH,
		e.getAssets,
		e.forwardPort,
		e.waitForAPIUp,
	}
	return e.executeStages(&result{}, stages)
}

func (e *Engine) TearDown(res *engine.Result) (*engine.Result, error) {
	e.name = res.ClusterID
	e.result = &result{
		instanceIP: res.InstanceIP,
	}
	stages := []stage{
		e.deleteKeyPair,
		e.deleteStack,
	}
	return e.executeStages(e.result, stages)
}

func (e *Engine) executeStages(res *result, stages []stage) (*engine.Result, error) {
	var err error
	for _, s := range stages {
		res, err = s(res)
		if err != nil {
			return nil, err
		}
	}
	e.result = res
	return &engine.Result{
		ClusterID:         e.name,
		CaCrtContent:      res.caCrtContent,
		ClientCrtContent:  res.clientCrtContent,
		ClientKeyContent:  res.clientKeyContent,
		KubeconfigContent: res.kubeconfigContent,
		InstanceIP:        res.instanceIP,
	}, nil
}

func (e *Engine) findImageID(res *result) (*result, error) {
	e.logger.Log("info", "looking for image id")

	filters := []*ec2.Filter{}
	filters = append(filters, &ec2.Filter{
		Name: aws.String("name"),
		Values: []*string{
			aws.String("shipyard minikube*"),
		},
	})

	i, err := e.clients.EC2.DescribeImages(&ec2.DescribeImagesInput{
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	if len(i.Images) < 1 {
		return res, fmt.Errorf("not found any shipyard image, refer to the project docs for info about how to create them")
	}

	res.imageID = *i.Images[0].ImageId
	e.logger.Log("info", fmt.Sprintf("found image with id %v", res.imageID))

	return res, nil
}

func (e *Engine) createKeyPair(res *result) (*result, error) {
	e.logger.Log("info", "creating key pair")

	input := &ec2.CreateKeyPairInput{
		KeyName: aws.String(e.name),
	}
	output, err := e.clients.EC2.CreateKeyPair(input)
	if err != nil {
		return nil, err
	}
	res.privateKeyContent = *output.KeyMaterial

	return res, nil
}

func (e *Engine) createStack(res *result) (*result, error) {
	e.logger.Log("info", "creating CloudFormation stack")

	stackInput := &cloudformation.CreateStackInput{
		StackName:        aws.String(e.name),
		TemplateBody:     aws.String(templateBody),
		TimeoutInMinutes: aws.Int64(2),
		Parameters: []*cloudformation.Parameter{
			{
				ParameterKey:   aws.String("ImageIDParameter"),
				ParameterValue: aws.String(res.imageID),
			},
			{
				ParameterKey:   aws.String("InstanceNameParameter"),
				ParameterValue: aws.String(e.name),
			},
			{
				ParameterKey:   aws.String("KeyPairNameParameter"),
				ParameterValue: aws.String(e.name),
			},
		},
	}

	output, err := e.clients.CloudFormation.CreateStack(stackInput)
	if err != nil {
		return nil, err
	}
	e.logger.Log("info", "CloudFormation stack created, waiting for completion")

	if err := e.clients.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(e.name),
	}); err != nil {
		return nil, err
	}

	res.stackID = *output.StackId

	// get instance ip
	describeInput := &cloudformation.DescribeStacksInput{
		StackName: aws.String(e.name),
	}
	describeOutput, err := e.clients.CloudFormation.DescribeStacks(describeInput)
	if err != nil {
		return nil, err
	}
	if len(describeOutput.Stacks) < 1 {
		return nil, fmt.Errorf("more than one stack with name %s", e.name)
	}
	for _, o := range describeOutput.Stacks[0].Outputs {
		if *o.OutputKey == "InstanceIP" {
			res.instanceIP = *o.OutputValue
		}
	}
	if res.instanceIP == "" {
		return nil, fmt.Errorf("instance IP not found")
	}
	e.logger.Log("info", fmt.Sprintf("CloudFormation stack completed, stack ID: %s, instance IP: %s", res.stackID, res.instanceIP))

	return res, nil
}

func (e *Engine) waitForSSH(res *result) (*result, error) {
	e.logger.Log("info", "waiting for ssh...")
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 1 * time.Minute
	ticker := backoff.NewTicker(b)

	var err error
	for _ = range ticker.C {
		if _, err = e.executeRemote("true", res.instanceIP, res.privateKeyContent); err != nil {
			e.logger.Log("info", fmt.Sprintf("ssh failed with %v, will retry...", err))
			continue
		}
		e.logger.Log("info", "ssh available")
		ticker.Stop()
		break
	}
	return res, nil
}

func (e *Engine) getAssets(res *result) (*result, error) {
	e.logger.Log("info", "getting assets")

	caCrtContent, err := e.executeRemote("cat /home/ubuntu/.minikube/ca.crt", res.instanceIP, res.privateKeyContent)
	if err != nil {
		return nil, err
	}
	res.caCrtContent = caCrtContent

	clientCrtContent, err := e.executeRemote("cat /home/ubuntu/.minikube/client.crt", res.instanceIP, res.privateKeyContent)
	if err != nil {
		return nil, err
	}
	res.clientCrtContent = clientCrtContent

	clientKeyContent, err := e.executeRemote("cat /home/ubuntu/.minikube/client.key", res.instanceIP, res.privateKeyContent)
	if err != nil {
		return nil, err
	}
	res.clientKeyContent = clientKeyContent

	kubeconfigContent, err := e.executeRemote("cat /home/ubuntu/.kube/config", res.instanceIP, res.privateKeyContent)
	if err != nil {
		return nil, err
	}
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Join(dir, ".shipyard")
	kubeconfigContent = strings.Replace(kubeconfigContent,
		"/home/ubuntu/.minikube",
		baseDir,
		-1)
	res.kubeconfigContent = kubeconfigContent

	return res, nil
}

func (e *Engine) forwardPort(res *result) (*result, error) {
	e.logger.Log("info", "forwarding port...")

	tmpfile, err := ioutil.TempFile("", "shipyard-pem")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpfile.Name())
	if err := ioutil.WriteFile(tmpfile.Name(), []byte(res.privateKeyContent), 0600); err != nil {
		return nil, err
	}

	cmd := exec.Command("ssh",
		"-i", tmpfile.Name(),
		"-o", "UserKnownHostsFile=/dev/null",
		"-o StrictHostKeyChecking=no",
		"ubuntu@"+res.instanceIP,
		"-NfL", "8443:127.0.0.1:8443")
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return res, nil
}

func (e *Engine) waitForAPIUp(res *result) (*result, error) {
	e.logger.Log("info", "waiting for API up...")

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 1 * time.Minute
	ticker := backoff.NewTicker(b)

	var err error
	for _ = range ticker.C {
		if _, err = e.executeRemote("kubectl get svc | grep kubernetes", res.instanceIP, res.privateKeyContent); err != nil {
			e.logger.Log("info", fmt.Sprintf("API not ready with error %v, will retry...", err))
			continue
		}
		e.logger.Log("info", "API up")
		ticker.Stop()
		break
	}
	return res, nil
}

func (e *Engine) deleteKeyPair(res *result) (*result, error) {
	e.logger.Log("info", "deleting private key")

	input := &ec2.DeleteKeyPairInput{
		KeyName: aws.String(e.name),
	}
	if _, err := e.clients.EC2.DeleteKeyPair(input); err != nil {
		return nil, err
	}
	return res, nil
}

func (e *Engine) deleteStack(res *result) (*result, error) {
	e.logger.Log("info", "deleting CloudFormation stack")

	if _, err := e.clients.CloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(e.name),
	}); err != nil {
		return nil, err
	}

	return res, nil
}

func (e *Engine) executeRemote(cmd string, addr string, privateKey string) (string, error) {
	key, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return "", err
	}
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	client, err := ssh.Dial("tcp", addr+":22", config)
	if err != nil {
		return "", err
	}

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var b bytes.Buffer
	session.Stdout = &b

	err = session.Run(cmd)
	return b.String(), err
}
