package shipyard

import (
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/shipyard/pkg/engine"
	"github.com/giantswarm/shipyard/pkg/files"
)

// Shipyard is a framework for e2e testing.
type Shipyard struct {
	engine       engine.Engine
	logger       micrologger.Logger
	filesHandler *files.Handler
}

// New initializes the global framework.
func New(logger micrologger.Logger, engine engine.Engine, filesHandler *files.Handler) *Shipyard {
	return &Shipyard{
		logger:       logger,
		engine:       engine,
		filesHandler: filesHandler,
	}
}

// Start spins up the related cluster, waiting for it to report ready
func (sy *Shipyard) Start() error {
	sy.logger.Log("info", "Starting shipyard...")
	result, err := sy.engine.SetUp()
	if err != nil {
		return err
	}
	return sy.filesHandler.Write(result)
}

// Stop finalizes the cluster
func (sy *Shipyard) Stop() error {
	sy.logger.Log("info", "Stopping shipyard...")
	initialResult, err := sy.filesHandler.ReadShipyardCfg()
	if err != nil {
		return err
	}
	_, err = sy.engine.TearDown(initialResult)
	return err
}
