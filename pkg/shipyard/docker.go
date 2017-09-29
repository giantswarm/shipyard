package shipyard

import (
	"os/exec"
	"strings"

	"github.com/giantswarm/micrologger"
)

// Docker is a simple shim to a Docker instance
type Docker interface {
	// Pull images into docker.
	Pull(images ...string) error
	// Run calls "docker run" args, returning the UUID of the container.
	Run(args ...string) (string, error)
	// Remove the container named by tag.
	Remove(tag string) error
	// Kill the container named by tag.
	Kill(tag string) error
	// List tags of containers that match filter. If filter is "", then all running containers
	// will be listed.
	List(filter string) ([]string, error)
}

// NewDocker returns a Docker for the default instance running on the host.
func NewDocker(logger micrologger.Logger) Docker {
	return &dockerWrapper{
		logger:     logger,
		dockerExec: "docker",
		socket:     "unix:///var/run/docker.sock",
	}
}

type dockerWrapper struct {
	logger micrologger.Logger

	dockerExec string

	socket string
}

var _ Docker = (*dockerWrapper)(nil)

func (d *dockerWrapper) Pull(images ...string) error {
	for _, image := range images {
		if err := d.runCommand([]string{"-H", d.socket, "pull", image}); err != nil {
			return err
		}
	}
	return nil
}

func (d *dockerWrapper) Run(args ...string) (string, error) {
	args = append(
		[]string{"-H", d.socket, "run"},
		args...)
	d.logger.Log("debug", "docker run %v", args)

	cmd := exec.Command(d.dockerExec, args...)
	output, err := cmd.CombinedOutput()
	d.logger.Log("debug", "docker output: ", string(output))

	if err != nil {
		return "", err
	}

	// This will be the UUID of the running container.
	return strings.TrimSpace(string(output)), nil
}

func (d *dockerWrapper) Remove(tag string) error {
	return d.runCommand([]string{"-H", d.socket, "rm", "-f", tag})
}

func (d *dockerWrapper) Kill(tag string) error {
	return d.runCommand([]string{"-H", d.socket, "kill", tag})
}

func (d *dockerWrapper) List(filter string) ([]string, error) {
	args := []string{"-H", d.socket, "ps", "-q"}
	if filter != "" {
		args = append(args, "--filter", filter)
	}
	d.logger.Log("debug", "docker %v", args)
	out, err := exec.Command(d.dockerExec, args...).Output()

	if err != nil {
		return []string{}, err
	}

	var ret []string
	for _, tag := range strings.Split(string(out), "\n") {
		if tag := strings.TrimSpace(tag); tag != "" {
			ret = append(ret, tag)
		}
	}

	return ret, nil
}

func (d *dockerWrapper) runCommand(args []string) error {
	d.logger.Log("debug", "docker %v", args)

	cmd := exec.Command(d.dockerExec, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		d.logger.Log("debug", "docker output: ", string(output))
		return err
	}
	return nil
}
