package shipyard

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/shipyard/pkg/shipyard/cluster"
	"github.com/giantswarm/shipyard/pkg/shipyard/docker"
	"github.com/giantswarm/shipyard/pkg/shipyard/files"
)

// Shipyard is a framework for e2e testing.
type Shipyard struct {
	cluster *cluster.Cluster
	cancel  chan struct{}
}

// New initializes the global framework.
func New() (*Shipyard, error) {
	if !canSudo() {
		return nil, fmt.Errorf("e2e test requires `sudo` to be active. Run `sudo -v` before running the e2e test.")
	}

	var err error

	baseDir, err := files.Init()
	if err != nil {
		return nil, err
	}

	logger, _ := micrologger.New(micrologger.DefaultConfig())

	logger.Log("debug", fmt.Sprintf("Creating framework (baseDir=%v)", baseDir))
	logger.Log("debug", fmt.Sprintf("It can be accessed with 'kubectl --kubeconfig %s/kubernetes/config ...'", baseDir))

	ch := make(chan struct{})
	keepSudoActive(logger, ch)

	docker := docker.New(logger)

	cl := cluster.New(baseDir, logger, docker)

	shipyard := &Shipyard{
		cluster: cl,
		cancel:  ch,
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
	close(sy.cancel)
	return sy.cluster.TearDown()
}

// canSudo returns true if the sudo command is allowed without a password.
func canSudo() bool {
	cmd := exec.Command("sudo", "-nv")
	return cmd.Run() == nil
}

// keepSudoActive periodically updates the sudo timestamp so we can keep
// running sudo.
func keepSudoActive(logger micrologger.Logger, cancel <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for {
			select {
			case <-cancel:
				ticker.Stop()
				return
			case <-ticker.C:
				if err := exec.Command("sudo", "-nv").Run(); err != nil {
					logger.Log("debug", "Unable to keep sudo active: %v", err)
				}
			}
		}
	}()
}
