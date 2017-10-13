package k8s

import (
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetClient(ci bool) (*kubernetes.Clientset, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	var configDir, server string
	if ci {
		configDir = ".shipyard"
		server = "127.0.0.1"
	} else {
		configDir = ".minikube"
		out, err := exec.Command("minikube", "ip").Output()
		if err != nil {
			return nil, err
		}
		minikubeIP := strings.TrimSpace(string(out))
		server = string(minikubeIP)
	}

	crtFile := filepath.Join(user.HomeDir, configDir, "client.crt")
	keyFile := filepath.Join(user.HomeDir, configDir, "client.key")
	caFile := filepath.Join(user.HomeDir, configDir, "ca.crt")

	config := &rest.Config{
		Host: "https://" + server + ":8443",
		TLSClientConfig: rest.TLSClientConfig{
			CertFile: crtFile,
			KeyFile:  keyFile,
			CAFile:   caFile,
		},
	}

	// create the clientset
	return kubernetes.NewForConfig(config)
}
