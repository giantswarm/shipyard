package shipyard

import (
	"log"
	"os"
	"path/filepath"

	"github.com/kubernetes/dns/pkg/e2e"
)

var fr *e2e.Framework

// Start spins up a minimal k8s cluster in 3 docker containers based on the
// hyperkube image, kube-apiserver, etcd and kubelet
func Start(workDir string) {
	pkgDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}

	baseDir, err := filepath.Abs(pkgDir + "../..")
	if err != nil {
		log.Fatalf("Error getting base directory: %v", err)
	}

	e2e.InitFramework(baseDir, workDir)
	fr = e2e.GetFramework()
	fr.SetUp()
}

// Stop finalizes the cluster and removes the working dir
func Stop() {
	fr.TearDown()
}
