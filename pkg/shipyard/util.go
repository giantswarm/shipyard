package shipyard

import (
	"os/exec"
	"time"
)

const (
	StandardTimeout = 10 * time.Second
)

// keepSudoActive periodically updates the sudo timestamp so we can keep
// running sudo.
func KeepSudoActive() {
	go func() {
		if err := exec.Command("sudo", "-nv").Run(); err != nil {
			Log.Fatalf("Unable to keep sudo active: %v", err)
		}
		time.Sleep(10 * time.Second)
	}()
}

// CanSudo returns true if the sudo command is allowed without a password.
func CanSudo() bool {
	cmd := exec.Command("sudo", "-nv")
	return cmd.Run() == nil
}

func makeSharedMount(path string) {
	if err := exec.Command("sudo", "mount", "--bind", path, path).Run(); err != nil {
		Log.Fatalf("Error bind mounting %v: %v", path, err)
	}
	if err := exec.Command("sudo", "mount", "--make-rshared", path).Run(); err != nil {
		Log.Fatalf("Error mount --make-rshared %v: %v", path, err)
	}
}

func umount(path string) {
	if err := exec.Command("sudo", "umount", path).Run(); err != nil {
		Log.Fatalf("Error umount %v: %v", path, err)
	}
}
