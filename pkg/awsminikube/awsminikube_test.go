package awsminikube_test

import (
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/shipyard/pkg/awsminikube"
	"github.com/giantswarm/shipyard/pkg/names"
)

func TestLeak(t *testing.T) {
	defer leaktest.CheckTimeout(t, 80*time.Second)()

	logger := microloggertest.New()
	engine := awsminikube.New("shipyard-test-"+names.Rand(7), awsminikube.DefaultConfig(), logger)

	res, err := engine.SetUp()
	if err != nil {
		t.Fatalf("Could not start shipyard: %v", err)
	}
	engine.TearDown(res)
}
