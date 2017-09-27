package shipyard

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Framework for e2e testing.
type Framework struct {
	Options Options
	Docker  Docker
	Cluster Cluster

	Processes map[string]*exec.Cmd
}

var (
	framework *Framework

	// Failed is set to true if a test case has failed. We are forced to
	// communicate this to ginkgo via a global variable.
	Failed bool
)

// InitFramework initializes the global framework.
func InitFramework(baseDir string, workDir string) {
	Log.Logf("Creating framework (baseDir=%v, workDir=%v)", baseDir, workDir)

	if !CanSudo() {
		Log.Fatalf(
			"e2e test requires `sudo` to be active. Run `sudo -v` before running the e2e test.")
	}
	KeepSudoActive()

	options := DefaultOptions(baseDir, workDir)
	docker := NewDocker()

	framework = &Framework{
		Options: options,
		Docker:  docker,
		Cluster: Cluster{
			Options: options,
			Docker:  docker,
		},
		Processes: make(map[string]*exec.Cmd),
	}

	for _, dir := range []string{
		workDir + "/logs",
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			Log.Fatalf("Could not mkdir %v: %v", workDir, err)
		}
	}
}

// GetFramework returns the global framework.
func GetFramework() *Framework {
	if framework == nil {
		Log.Fatal("InitFramework must be called before use")
	}
	return framework
}

// SetUp the framework.
func (fr *Framework) SetUp() {
	fr.Cluster.SetUp()
}

// TearDown the framework.
func (fr *Framework) TearDown() {
	fr.Cluster.TearDown()

	if Failed {
		for name := range fr.Processes {
			Log.Logf("Failure detected, dumping logs for '%v'", name)
			Log.Logf("==== %v stdout ====", name)
			f, err := os.Open(fr.StdoutLogfile(name))
			if err != nil {
				Log.Fatalf("Could not open %v: %v", fr.StdoutLogfile(name), err)
			}
			io.Copy(os.Stderr, f)

			Log.Logf("==== %v stderr ====", name)
			f, err = os.Open(fr.StderrLogfile(name))
			if err != nil {
				Log.Fatalf("Could not open %v: %v", fr.StderrLogfile(name), err)
			}
			io.Copy(os.Stderr, f)
		}
	}
}

// Path returns an absolute path for a relative path in the repository.
func (fr *Framework) Path(relative string) string {
	ret, err := filepath.Abs(fr.Options.BaseDir + "/" + relative)
	if err != nil {
		Log.Fatal(err)
	}
	return ret
}

// StdoutLogfile is stdout log file for RunInBackground.
func (fr *Framework) StdoutLogfile(name string) string {
	return fmt.Sprintf("%v/logs/%v.out", fr.Options.WorkDir, name)
}

// StderrLogfile is the stderr log file for RunInBackground.
func (fr *Framework) StderrLogfile(name string) string {
	return fmt.Sprintf("%v/logs/%v.err", fr.Options.WorkDir, name)
}
