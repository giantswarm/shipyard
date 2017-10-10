package files

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/giantswarm/shipyard/pkg/engine"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"
)

type Handler struct {
	fs afero.Fs
}

func NewHandler(fs afero.Fs) *Handler {
	return &Handler{fs: fs}
}

func (h *Handler) Write(res *engine.Result) error {
	// certs
	if err := h.writeCerts(res); err != nil {
		return err
	}

	// kubeconfig
	if err := h.writeKubeConfig(res); err != nil {
		return err
	}

	// shipyard config
	if err := h.writeShipyardCfg(res); err != nil {
		return err
	}

	return nil
}

func (h *Handler) writeCerts(res *engine.Result) error {
	user, err := user.Current()
	if err != nil {
		return err
	}

	baseDir := filepath.Join(user.HomeDir, ".shipyard")
	if err := h.fs.MkdirAll(baseDir, 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(
		filepath.Join(baseDir, "ca.crt"),
		[]byte(res.CaCrtContent),
		0644); err != nil {
		return err
	}

	if err := ioutil.WriteFile(
		filepath.Join(baseDir, "client.crt"),
		[]byte(res.ClientCrtContent),
		0644); err != nil {
		return err
	}

	if err := ioutil.WriteFile(
		filepath.Join(baseDir, "client.key"),
		[]byte(res.ClientKeyContent),
		0644); err != nil {
		return err
	}

	return nil
}

func (h *Handler) writeKubeConfig(res *engine.Result) error {
	user, err := user.Current()
	if err != nil {
		return err
	}

	baseDir := filepath.Join(user.HomeDir, ".shipyard")
	if err := h.fs.MkdirAll(baseDir, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(
		filepath.Join(baseDir, "config"),
		[]byte(res.KubeconfigContent),
		0644); err != nil {
		return err
	}

	return nil
}

func (h *Handler) writeShipyardCfg(res *engine.Result) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	content, err := yaml.Marshal(&res)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(
		filepath.Join(dir, ".shipyard.yaml"),
		[]byte(content),
		0644); err != nil {
		return err
	}

	return nil
}

func (h *Handler) ReadShipyardCfg() (*engine.Result, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(filepath.Join(dir, ".shipyard.yaml"))
	if err != nil {
		return nil, err
	}

	e := &engine.Result{}

	if err := yaml.Unmarshal(content, e); err != nil {
		return nil, err
	}
	return e, nil
}
