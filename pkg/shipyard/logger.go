package shipyard

import "github.com/giantswarm/micrologger"

var Log micrologger.Logger

func init() {
	Log, _ = micrologger.New(micrologger.DefaultConfig())
}
