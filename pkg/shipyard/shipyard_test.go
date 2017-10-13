package shipyard_test

import (
	"context"
	"os"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/microstorage/storagetest"
	"github.com/giantswarm/shipyard/pkg/awsminikube"
	"github.com/giantswarm/shipyard/pkg/files"
	"github.com/giantswarm/shipyard/pkg/k8s"
	"github.com/giantswarm/shipyard/pkg/names"
	"github.com/giantswarm/shipyard/pkg/shipyard"
	"github.com/giantswarm/tprstorage"
	"github.com/spf13/afero"
)

func TestShipyard(t *testing.T) {
	var ci bool
	logger := microloggertest.New()
	if os.Getenv("CIRCLECI") == "true" {
		ci = true
		engine := awsminikube.New("shipyard-test-"+names.Rand(7), awsminikube.DefaultConfig(), logger)
		filesHandler := files.NewHandler(afero.NewOsFs())
		sy := shipyard.New(logger, engine, filesHandler)

		if err := sy.Start(); err != nil {
			t.Fatalf("Could not start shipyard: %v", err)
		}
		defer sy.Stop()
	}

	t.Run("tpr storage example", func(t *testing.T) {
		k8sClient, err := k8s.GetClient(ci)
		if err != nil {
			t.Fatalf("error creating K8s client: %#v", err)
		}

		var storage *tprstorage.Storage

		config := tprstorage.DefaultConfig()
		config.K8sClient = k8sClient
		config.Logger = logger

		config.TPO.Name = "integration-test"

		storage, err = tprstorage.New(config)
		if err != nil {
			t.Fatalf("error creating storage: %#v", err)
		}

		err = storage.Boot(context.TODO())
		if err != nil {
			t.Fatalf("error booting storage: %#v", err)
		}

		storagetest.Test(t, storage)
	})
}
