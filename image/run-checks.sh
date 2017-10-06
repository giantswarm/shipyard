#!/bin/sh
set -ex

SHIPYARD_AMI=${1:-ami-7a46fb15}

echo "$SHIPYARD_PRIVATE_KEY" | base64 -d | tee shipyard.pem && chmod 600 ./shipyard.pem

instanceId=$(echo "$(aws ec2 run-instances --image-id $SHIPYARD_AMI --tag-specifications ResourceType=instance,Tags=[{Key=Name,Value=ci-awsop}] --instance-type t2.medium --key-name shipyard --security-groups shipyard --region eu-central-1)" | grep "InstanceId" | cut -d '"' -f4)
trap 'aws ec2 terminate-instances --instance-ids $instanceId --region eu-central-1' EXIT

publicIp=$(echo "$(aws ec2 describe-instances --instance-ids $instanceId --region eu-central-1)" | grep PublicIpAddress | cut -d '"' -f4)

retry=30
while ! ssh -i shipyard.pem -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ubuntu@${publicIp} true; do
    retry=$(( retry - 1 ))
    if [ $retry -le 0 ]; then
        echo "Timed out waiting for ssh. Aborting!"
        exit 1
    fi
    echo "Waiting for ssh..."
    sleep 1
done

mkdir -p ~/.minikube ~/.kube
scp -r -i shipyard.pem -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no  ubuntu@${publicIp}:/home/ubuntu/.minikube/{ca.crt,client.crt,client.key} ~/.minikube/
scp -r -i shipyard.pem -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ubuntu@${publicIp}:/home/ubuntu/.kube/config ~/.kube/

ssh -i shipyard.pem -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ubuntu@${publicIp} -NfL 8443:127.0.0.1:8443

retry=30
expected="(?s)scheduler *Healthy *ok.*controller-manager *Healthy *ok.*etcd-0 *Healthy"
while ! kubectl get cs | grep -Pz "$expected"; do
    retry=$(( retry - 1 ))
    if [ $retry -le 0 ]; then
        echo "Timed out waiting for cluster up. Aborting!"
        exit 1
    fi
    echo "Waiting for cluster up..."
    sleep 5
done

retry=30
expected="kubernetes *"
while ! kubectl get svc | grep -Pz "$expected"; do
    retry=$(( retry - 1 ))
    if [ $retry -le 0 ]; then
        echo "Timed out waiting for kubernetes svc. Aborting!"
        exit 1
    fi
    echo "Waiting for kubernetes svc..."
    sleep 5
done

helm init

retry=30
expected="kube-system *tiller.* *1/1 *Running *"
while ! kubectl get pod --all-namespaces | grep -Pz "$expected"; do
    retry=$(( retry - 1 ))
    if [ $retry -le 0 ]; then
        echo "Timed out waiting for tiller up. Aborting!"
        exit 1
    fi
    echo "Waiting for tiller up..."
    sleep 5
done

helm registry install quay.io/giantswarm/cert-operator-lab-chart -- \
       -n cert-operator-lab \
       --set imageTag=latest \
       --wait

retry=30
expected="certificate\.giantswarm\.io"
while ! kubectl get thirdpartyresources | grep -Pz "$expected"; do
    retry=$(( retry - 1 ))
    if [ $retry -le 0 ]; then
        echo "Timed out waiting for TPR. Aborting!"
        exit 1
    fi
    echo "Waiting for TPR..."
    sleep 1
done

sleep 5

helm registry install quay.io/giantswarm/cert-resource-lab-chart -- \
       -n cert-resource-lab
