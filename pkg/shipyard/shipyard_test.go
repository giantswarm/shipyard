package shipyard_test

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/giantswarm/shipyard/pkg/shipyard"
)

func TestAPIUp(t *testing.T) {
	workDir, err := ioutil.TempDir("", "gs-shipyard")
	if err != nil {
		log.Fatalf("Error creating working directory: %v", err)
	}
	defer os.RemoveAll(workDir)

	shipyard.Start(workDir)
	defer shipyard.Stop()

	_, err = http.Get("http://127.0.0.1:8080")
	if err != nil {
		t.Errorf("error accesing api, %v", err)
	}
}

/*
func TestTPRStorage(t *testing.T) {
	workDir, err := ioutil.TempDir("", "gs-shipyard")
	if err != nil {
		log.Fatalf("Error creating working directory: %v", err)
	}
	defer os.RemoveAll(workDir)

	shipyard.Start(workDir)
	defer shipyard.Stop()

	cfgDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("error getting config directory: %v", err)
	}

	kubeconfig, err := filepath.Abs(cfgDir + "../../test/e2e/cluster/config")
	if err != nil {
		t.Fatalf("Error getting base directory: %v", err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatalf("error creating k8s config %#v", err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("error creating K8s client: %#v", err)
	}

	var storage *tprstorage.Storage
	{
		config := tprstorage.DefaultConfig()
		config.K8sClient = k8sClient
		config.Logger = microloggertest.New()

		config.TPO.Name = "integration-test"

		storage, err = tprstorage.New(config)
		if err != nil {
			t.Fatalf("error creating storage: %#v", err)
		}

		defer func() {
			path := path.Join(storage.tpr.Endpoint(config.TPO.Namespace), config.TPO.Name)
			_, err := k8sClient.CoreV1().RESTClient().Delete().AbsPath(path).DoRaw()
			if err != nil {
				t.Logf("error cleaning up TPO %s/%s: %#v", config.TPO.Namespace, config.TPO.Name, err)
			}
		}()
	}

	err = storage.Boot(context.TODO())
	if err != nil {
		t.Fatalf("error booting storage: %#v", err)
	}

	storagetest.Test(t, storage)
}
*/
