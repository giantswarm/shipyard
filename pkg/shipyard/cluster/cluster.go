package cluster

import (
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/shipyard/pkg/shipyard/docker"
)

const (
	startupTimeout = 10 * time.Second
	etcdImage      = "quay.io/coreos/etcd:v3.2.7"
	hyperkubeImage = "gcr.io/google_containers/hyperkube:v1.7.6"
)

type taskFn func() error

type Config struct {
	logger micrologger.Logger

	Kubectl string

	BaseDir string

	EtcdImage      string
	HyperkubeImage string
}

// Cluster encapsulates a mock Kubernetes cluster.
type Cluster struct {
	Config
	Docker docker.Docker

	containers struct {
		etcd    string
		api     string
		kubelet string
	}

	manifestDir   string
	certDir       string
	varLibDocker  string
	varLibKubelet string
	varRun        string
}

// defaultConfig to use to run the e2e test.
func defaultConfig(baseDir string, logger micrologger.Logger) Config {
	return Config{
		logger: logger,

		Kubectl: "kubectl",

		BaseDir: baseDir,

		EtcdImage:      etcdImage,
		HyperkubeImage: hyperkubeImage,
	}
}

// New is the cluster initializer
func New(baseDir string, logger micrologger.Logger, docker docker.Docker) *Cluster {
	config := defaultConfig(baseDir, logger)

	return &Cluster{
		Config: config,
		Docker: docker,
	}
}

// SetUp the e2e cluster.
func (cl *Cluster) SetUp() error {
	cl.logger.Log("debug", "SetUp")

	tasks := []taskFn{
		cl.resolveDirs,
		cl.pullImages,
		cl.startEtcd,
		cl.startAPIServer,
		cl.startKubelet,
		cl.waitForAPIServer,
	}

	return runTasks(tasks)
}

// TearDown the e2e cluster.
func (cl *Cluster) TearDown() error {
	cl.logger.Log("debug", "Teardown")

	tasks := []taskFn{cl.stopKubelet, cl.stopAPIServer, cl.stopEtcd}

	return runTasks(tasks)
}

func runTasks(tasks []taskFn) error {
	for _, task := range tasks {
		if err := task(); err != nil {
			return err
		}
	}
	return nil
}

func (cl *Cluster) resolveDirs() error {
	// TODO: directories should be configurable, but there seem to be issues with the
	// the nsenter mounter that prevent us from moving the location of /var/lib/kubelet.
	cl.manifestDir = fmt.Sprintf("%v/kubernetes/manifests", cl.BaseDir)
	cl.certDir = fmt.Sprintf("%v/kubernetes/cert", cl.BaseDir)
	cl.varLibDocker = "/var/lib/docker"
	cl.varLibKubelet = "/var/lib/kubelet"
	cl.varRun = "/var/run"

	var err error

	cl.manifestDir, err = filepath.Abs(cl.manifestDir)
	if err != nil {
		return err
	}

	cl.varLibDocker, err = filepath.Abs(cl.varLibDocker)
	if err != nil {
		return err
	}

	cl.varRun, err = filepath.Abs(cl.varRun)
	if err != nil {
		return err
	}

	cl.varLibKubelet, err = filepath.Abs(cl.varLibKubelet)
	if err != nil {
		return err
	}
	return nil
}

func (cl *Cluster) pullImages() error {
	return cl.Docker.Pull(
		cl.EtcdImage,
		cl.HyperkubeImage)
}

func (cl *Cluster) startEtcd() error {
	cl.logger.Log("debug", "Starting etcd")

	var err error
	cl.containers.etcd, err = cl.Docker.Run("-d", "--net=host", cl.EtcdImage)
	return err
}

func (cl *Cluster) stopEtcd() error {
	if cl.containers.etcd == "" {
		return nil
	}

	cl.logger.Log("debug", "Stopping etcd")

	if err := cl.Docker.Kill(cl.containers.etcd); err != nil {
		return err
	}

	cl.containers.etcd = ""
	return nil
}

func (cl *Cluster) startAPIServer() error {
	cl.logger.Log("debug", "Starting API server")

	var err error

	cl.containers.api, err = cl.Docker.Run(
		"-d",
		fmt.Sprintf("--volume=%v:/src:ro", cl.BaseDir),
		"--net=host",
		"--pid=host",
		cl.HyperkubeImage,
		"/hyperkube", "apiserver",
		"--insecure-bind-address=0.0.0.0",
		"--service-cluster-ip-range=10.0.0.1/24",
		"--etcd_servers=http://127.0.0.1:2379",
		"--v=2")

	return err
}

func (cl *Cluster) stopAPIServer() error {
	if cl.containers.api == "" {
		return nil
	}

	cl.logger.Log("debug", "Stopping API server")
	if err := cl.Docker.Kill(cl.containers.api); err != nil {
		return err
	}
	cl.containers.api = ""
	return nil
}

func (cl *Cluster) waitForAPIServer() error {
	deadline := time.Now().Add(startupTimeout)

	for time.Now().Before(deadline) {
		if resp, err := http.Get("http://localhost:8080"); err == nil {
			defer resp.Body.Close()
			cl.logger.Log("debug", "API server started")
			return nil
		}
		cl.logger.Log("debug", "Waiting for API server to start")
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("API server failed to start")
}

func (cl *Cluster) startKubelet() error {
	cl.logger.Log("debug", "Starting Kubelet")

	var err error

	err = exec.Command("sudo", "mkdir", "-p", cl.varLibKubelet).Run()
	if err != nil {
		cl.logger.Log("error", "Could not create %v: %v", cl.varLibKubelet, err)
		return err
	}

	args := []string{
		"-d",
		"--volume=/:/rootfs:ro", // This is used by the nsenter mounter.
		"--volume=/sys:/sys:ro",
		"--volume=/dev:/dev",
		fmt.Sprintf("--volume=%v:/src:ro", cl.BaseDir),
		fmt.Sprintf("--volume=%v:/etc/kubernetes/manifests-e2e:ro", cl.manifestDir),
		fmt.Sprintf("--volume=%v:/srv/kubernetes:ro", cl.certDir),
		fmt.Sprintf("--volume=%v:/var/lib/docker:rw", cl.varLibDocker),
		fmt.Sprintf("--volume=%v:/var/run:rw", cl.varRun),
		fmt.Sprintf("--volume=%v:/var/lib/kubelet:rw", cl.varLibKubelet),
		"--net=host",
		"--pid=host",
		"--privileged=true",
		cl.HyperkubeImage,
		"/hyperkube", "kubelet",
		"--v=4",
		"--containerized",
		"--hostname-override=0.0.0.0",
		"--address=0.0.0.0",
		"--cluster_dns=10.0.0.10",
		"--cluster_domain=cluster.local",
		"--require-kubeconfig",
		"--kubeconfig=/src/kubernetes/config",
		"--pod-manifest-path=/etc/kubernetes/manifests-e2e",
	}

	cl.containers.kubelet, err = cl.Docker.Run(args...)

	return err
}

func (cl *Cluster) stopKubelet() error {
	if cl.containers.kubelet == "" {
		return nil
	}

	cl.logger.Log("debug", "Stopping Kubelet")

	cl.Docker.Kill(cl.containers.kubelet)
	cl.containers.kubelet = ""

	// Remove all containers created by kubelet.
	containers, err := cl.Docker.List("name=k8s_*")
	if err != nil {
		return err
	}
	for _, tag := range containers {
		if err = cl.Docker.Kill(tag); err != nil {
			return err
		}
	}

	if err = exec.Command("sudo", "rm", "-rf", cl.varLibKubelet).Run(); err != nil {
		return err
	}
	return nil
}
