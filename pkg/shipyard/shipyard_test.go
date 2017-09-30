package shipyard_test

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/microstorage/storagetest"
	"github.com/giantswarm/shipyard/pkg/shipyard"
	"github.com/giantswarm/shipyard/pkg/shipyard/files"
	"github.com/giantswarm/tprstorage"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestShipyard(t *testing.T) {
	var err error

	sy, err := shipyard.New()
	if err != nil {
		t.Fatalf("Could not start cluster: %v", err)
	}

	if err = sy.Start(); err != nil {
		t.Fatalf("could not start framework, %v", err)
	}
	defer sy.Stop()

	t.Run("API up", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:8080")
		if err != nil {
			t.Fatalf("could not access api, %v", err)
		}
		resp.Body.Close()
	})

	t.Run("tpr storage example", func(t *testing.T) {
		k8sClient, err := getK8sClient()
		if err != nil {
			t.Fatalf("error creating K8s client: %#v", err)
		}

		var storage *tprstorage.Storage

		config := tprstorage.DefaultConfig()
		config.K8sClient = k8sClient
		config.Logger = microloggertest.New()

		config.TPO.Name = "integration-test"

		storage, err = tprstorage.New(config)
		if err != nil {
			t.Fatalf("error creating storage: %#v", err)
		}

		err = storage.Boot(context.TODO())
		if err != nil {
			t.Fatalf("error booting storage: %#v", err)
		}

		storagetest.Test(t, storage)
	})
}

func getK8sClient() (*kubernetes.Clientset, error) {
	baseDir, err := files.Init()
	if err != nil {
		return nil, err
	}

	kubeconfig := filepath.Join(baseDir, "kubernetes/config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	return kubernetes.NewForConfig(config)
}

func TestLeak(t *testing.T) {
	defer leaktest.CheckTimeout(t, 15*time.Second)()

	sy, err := shipyard.New()
	if err != nil {
		t.Fatalf("Could not start cluster: %v", err)
	}

	if err = sy.Start(); err != nil {
		t.Fatalf("could not start framework, %v", err)
	}
	sy.Stop()
}
