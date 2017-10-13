package shipyard_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/microstorage/storagetest"
	"github.com/giantswarm/shipyard/pkg/awsminikube"
	"github.com/giantswarm/shipyard/pkg/files"
	"github.com/giantswarm/shipyard/pkg/names"
	"github.com/giantswarm/shipyard/pkg/shipyard"
	"github.com/giantswarm/tprstorage"
	"github.com/spf13/afero"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestShipyard(t *testing.T) {
	logger := microloggertest.New()
	engine := awsminikube.New("shipyard-test-"+names.Rand(7), awsminikube.DefaultConfig(), logger)
	filesHandler := files.NewHandler(afero.NewOsFs())
	sy := shipyard.New(logger, engine, filesHandler)

	if err := sy.Start(); err != nil {
		t.Fatalf("Could not start shipyard: %v", err)
	}
	defer sy.Stop()

	t.Run("API up", func(t *testing.T) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get("https://127.0.0.1:8443")
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
		config.Logger = logger

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
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(user.HomeDir, ".shipyard", "config"))
	if err != nil {
		return nil, err
	}

	// create the clientset
	return kubernetes.NewForConfig(config)
}
