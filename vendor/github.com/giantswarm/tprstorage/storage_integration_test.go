// +build integration

package tprstorage

/*
	Usage:

		go test -tags=integration $(glide novendor) [FLAGS]

	Flags:

		-integration.ca string
			CA file path (default "$HOME/.minikube/ca.crt")
		-integration.crt string
			certificate file path (default "$HOME/.minikube/apiserver.crt")
		-integration.key string
			key file path (default "$HOME/.minikube/apiserver.key")
		-integration.server string
			Kubernetes API server address (default "https://$(minikube ip):8443")
*/

import (
	"context"
	"flag"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/microstorage/storagetest"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	server  string
	crtFile string
	keyFile string
	caFile  string
)

func init() {
	u, err := user.Current()
	homePath := func(relativePath string) string {
		if err != nil {
			return ""
		}
		return path.Join(u.HomeDir, relativePath)
	}

	var serverDefault string
	{
		out, err := exec.Command("minikube", "ip").Output()
		if err == nil {
			minikubeIP := strings.TrimSpace(string(out))
			serverDefault = "https://" + string(minikubeIP) + ":8443"
		}
	}

	flag.StringVar(&server, "integration.server", serverDefault, "Kubernetes API server address")
	flag.StringVar(&crtFile, "integration.crt", homePath(".minikube/apiserver.crt"), "certificate file path")
	flag.StringVar(&keyFile, "integration.key", homePath(".minikube/apiserver.key"), "key file path")
	flag.StringVar(&caFile, "integration.ca", homePath(".minikube/ca.crt"), "CA file path")
}

func TestIntegration(t *testing.T) {
	var err error

	var k8sClient *kubernetes.Clientset
	{
		config := &rest.Config{
			Host: server,
			TLSClientConfig: rest.TLSClientConfig{
				CertFile: crtFile,
				KeyFile:  keyFile,
				CAFile:   caFile,
			},
		}

		k8sClient, err = kubernetes.NewForConfig(config)
		if err != nil {
			t.Fatalf("error creating K8s client: %#v", err)
		}
	}

	var storage *Storage
	{
		config := DefaultConfig()
		config.K8sClient = k8sClient
		config.Logger = microloggertest.New()

		config.TPO.Name = "integration-test"

		storage, err = New(config)
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
