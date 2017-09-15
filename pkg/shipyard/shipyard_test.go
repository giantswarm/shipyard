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
