package shipyard

import (
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	startupTimeout = 10 * time.Second
)

type taskFn func() error

// Cluster encapsulates a mock Kubernetes cluster.
type Cluster struct {
	Options
	Docker Docker

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

// SetUp the e2e cluster.
func (cl *Cluster) SetUp() error {
	Log.Log("debug", "SetUp")

	tasks := []taskFn{
		cl.resolveDirs,
		cl.pullImages,
		cl.startEtcd,
		cl.startApiServer,
		cl.startKubelet,
		cl.waitForApiServer,
	}

	return runTasks(tasks)
}

// TearDown the e2e cluster.
func (cl *Cluster) TearDown() error {
	Log.Log("debug", "Teardown")

	tasks := []taskFn{cl.stopKubelet, cl.stopApiServer, cl.stopEtcd}

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
	cl.manifestDir = fmt.Sprintf("%v/test/e2e/cluster/manifests", cl.BaseDir)
	cl.certDir = fmt.Sprintf("%v/test/e2e/cluster/cert", cl.BaseDir)
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
	Log.Log("debug", "Starting etcd")

	var err error
	cl.containers.etcd, err = cl.Docker.Run("-d", "--net=host", cl.EtcdImage)
	return err
}

func (cl *Cluster) stopEtcd() error {
	if cl.containers.etcd == "" {
		return nil
	}

	Log.Log("debug", "Stopping etcd")

	if err := cl.Docker.Kill(cl.containers.etcd); err != nil {
		return err
	}

	cl.containers.etcd = ""
	return nil
}

func (cl *Cluster) startApiServer() error {
	Log.Log("debug", "Starting API server")

	var err error

	cl.containers.api, err = cl.Docker.Run(
		"-d",
		fmt.Sprintf("--volume=%v:/src:ro", cl.BaseDir),
		fmt.Sprintf("--volume=%v:/data:rw", cl.WorkDir),
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

func (cl *Cluster) stopApiServer() error {
	if cl.containers.api == "" {
		return nil
	}

	Log.Log("debug", "Stopping API server")
	if err := cl.Docker.Kill(cl.containers.api); err != nil {
		return err
	}
	cl.containers.api = ""
	return nil
}

func (cl *Cluster) waitForApiServer() error {
	deadline := time.Now().Add(startupTimeout)

	for time.Now().Before(deadline) {
		if _, err := http.Get("http://localhost:8080"); err == nil {
			Log.Log("debug", "API server started")
			return nil
		}
		Log.Log("debug", "Waiting for API server to start")
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("API server failed to start")
}

func (cl *Cluster) startKubelet() error {
	Log.Log("debug", "Starting Kubelet")

	var err error

	err = exec.Command("sudo", "mkdir", "-p", cl.varLibKubelet).Run()
	if err != nil {
		Log.Log("error", "Could not create %v: %v", cl.varLibKubelet, err)
		return err
	}

	err = makeSharedMount(cl.varLibKubelet)
	if err != nil {
		return err
	}

	args := []string{
		"-d",
		"--volume=/:/rootfs:ro", // This is used by the nsenter mounter.
		"--volume=/sys:/sys:ro",
		"--volume=/dev:/dev",
		fmt.Sprintf("--volume=%v:/src:ro", cl.BaseDir),
		fmt.Sprintf("--volume=%v:/data:rw", cl.WorkDir),
		fmt.Sprintf("--volume=%v:/etc/kubernetes/manifests-e2e:ro", cl.manifestDir),
		fmt.Sprintf("--volume=%v:/srv/kubernetes:ro", cl.certDir),
		fmt.Sprintf("--volume=%v:/var/lib/docker:rw", cl.varLibDocker),
		fmt.Sprintf("--volume=%v:/var/run:rw", cl.varRun),
		fmt.Sprintf("--volume=%v:/var/lib/kubelet:shared", cl.varLibKubelet),
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
		"--kubeconfig=/src/test/e2e/cluster/config",
		"--pod-manifest-path=/etc/kubernetes/manifests-e2e",
	}

	cl.containers.kubelet, err = cl.Docker.Run(args...)

	return err
}

func (cl *Cluster) stopKubelet() error {
	if cl.containers.kubelet == "" {
		return nil
	}

	Log.Log("debug", "Stopping Kubelet")

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

	if err = umount(cl.varLibKubelet); err != nil {
		return err
	}

	if err = exec.Command("sudo", "rm", "-rf", cl.varLibKubelet).Run(); err != nil {
		return err
	}
	return nil
}

func umount(path string) error {
	return exec.Command("sudo", "umount", path).Run()
}

func makeSharedMount(path string) error {
	if err := exec.Command("sudo", "mount", "--bind", path, path).Run(); err != nil {
		return err
	}
	if err := exec.Command("sudo", "mount", "--make-rshared", path).Run(); err != nil {
		return err
	}
	return nil
}
