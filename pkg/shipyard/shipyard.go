package shipyard

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/giantswarm/micrologger"
)

// Shipyard is a framework for e2e testing.
type Shipyard struct {
	cluster Cluster
}

var shipyard *Shipyard

// New initializes the global framework.
func New(workDir string) (*Shipyard, error) {
	var err error

	baseDir, err := PrepareBaseDir()
	if err != nil {
		return nil, err
	}

	logger, _ := micrologger.New(micrologger.DefaultConfig())

	logger.Log("debug", fmt.Sprintf("Creating framework (baseDir=%v, workDir=%v)", baseDir, workDir))
	logger.Log("debug", fmt.Sprintf("It can be accessed with 'kubectl --kubeconfig %s/kubernetes/config ...'", baseDir))

	if !canSudo() {
		return nil, fmt.Errorf("e2e test requires `sudo` to be active. Run `sudo -v` before running the e2e test.")
	}
	keepSudoActive(logger)

	config := DefaultConfig(baseDir, workDir, logger)
	docker := NewDocker(logger)

	shipyard = &Shipyard{
		cluster: Cluster{
			Config: config,
			Docker: docker,
		},
	}

	if err := os.MkdirAll(workDir+"/logs", 0755); err != nil {
		return nil, fmt.Errorf("Could not mkdir %v: %v", workDir, err)
	}

	return shipyard, nil
}

// Start spins up a minimal k8s cluster in 3 base docker containers based on the
// hyperkube image, kube-apiserver, etcd and kubelet and 2 additional static
// pods running in the kubelet, controller-manager and scheduler.
func (sy *Shipyard) Start() error {
	return sy.cluster.SetUp()
}

// Stop finalizes the cluster and removes the working dir
func (sy *Shipyard) Stop() error {
	return sy.cluster.TearDown()
}

// canSudo returns true if the sudo command is allowed without a password.
func canSudo() bool {
	cmd := exec.Command("sudo", "-nv")
	return cmd.Run() == nil
}

// keepSudoActive periodically updates the sudo timestamp so we can keep
// running sudo.
func keepSudoActive(logger micrologger.Logger) {
	go func() {
		if err := exec.Command("sudo", "-nv").Run(); err != nil {
			logger.Log("debug", "Unable to keep sudo active: %v", err)
		}
		time.Sleep(10 * time.Second)
	}()
}
