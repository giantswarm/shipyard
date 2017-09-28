package shipyard

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Docker is a simple shim to a Docker instance
type Docker interface {
	// Start the daemon (if needed)
	Start() error
	// Stop the daemon
	Stop() error
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
func NewDocker() Docker {
	return &dockerWrapper{
		dockerExec:   "docker",
		manageDaemon: false,
		baseDir:      "/",
		cidr:         "10.123.0.0/24",
		bridge:       "docker0",
		socket:       "unix:///var/run/docker.sock",
	}
}

type dockerWrapper struct {
	dockerExec string

	manageDaemon bool
	baseDir      string
	cidr         string
	bridge       string

	socket string
	cmd    *exec.Cmd
}

var _ Docker = (*dockerWrapper)(nil)

func (d *dockerWrapper) Start() error {
	if !d.manageDaemon {
		Log.Log("debug", "not set to manage daemon, exiting")
		return nil
	}

	execDir := d.baseDir + "/var/lib/docker"
	graphDir := d.baseDir + "/var/run/docker"

	if err := os.MkdirAll(execDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(graphDir, 0755); err != nil {
		return err
	}

	pidfile := d.baseDir + "/pid"
	d.socket = "unix://" + d.baseDir + "/var/run/docker.sock"

	if err := d.ensureBridge(); err != nil {
		return nil
	}

	args := []string{
		d.dockerExec, "daemon",
		"--bridge=" + d.bridge,
		"--exec-root=" + execDir,
		"--graph=" + graphDir,
		"--host=" + d.socket,
		"--pidfile=" + pidfile,
	}

	d.cmd = exec.Command("sudo", args...)

	Log.Log("debug", "Starting Docker %v", args)
	if err := d.cmd.Start(); err != nil {
		return err
	}

	return d.waitForStart()
}

func (d *dockerWrapper) Stop() error {
	if !d.manageDaemon {
		Log.Log("debug", "not set to manage daemon, exiting")
		return nil
	}

	// Need to use sudo kill as the docker daemon is running as `root`.
	if err := exec.Command(
		"sudo", "kill", fmt.Sprintf("%v", d.cmd.Process.Pid)).Run(); err != nil {
		return err
	}
	state, err := d.cmd.Process.Wait()
	if err != nil {
		Log.Log("debug", "Wait for docker failed")
		return err
	}
	Log.Log("debug", "Docker exited with %v", state)
	return nil
}

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
	Log.Log("debug", "docker run %v", args)

	cmd := exec.Command(d.dockerExec, args...)
	output, err := cmd.CombinedOutput()
	Log.Log("debug", "docker output: ", string(output))

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
	Log.Log("debug", "docker %v", args)
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
	Log.Log("debug", "docker %v", args)

	cmd := exec.Command(d.dockerExec, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		Log.Log("debug", "docker output: ", string(output))
		return err
	}
	return nil
}

func (d *dockerWrapper) ensureBridge() error {
	if exec.Command("ip", "link", "show", d.bridge).Run() == nil {
		Log.Log("debug", "Bridge device %v exists", d.bridge)
		return nil
	}

	Log.Log("debug", "Creating bridge device %v (%v)", d.bridge, d.cidr)
	if err := exec.Command("sudo", "brctl", "addbr", d.bridge).Run(); err != nil {
		return err
	}
	if err := exec.Command("sudo", "ip", "addr", "add", d.cidr, "dev", d.bridge).Run(); err != nil {
		return err
	}
	if err := exec.Command("sudo", "ip", "link", "set", "dev", d.bridge, "up").Run(); err != nil {
		return err
	}
	return nil
}

func (d *dockerWrapper) waitForStart() error {
	for i := 0; i < 10; i++ {
		if err := exec.Command(d.dockerExec, "-H", d.socket, "info").Run(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("docker daemon didn't started")
}
