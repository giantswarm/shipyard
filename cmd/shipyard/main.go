package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/shipyard/pkg/awsminikube"
	"github.com/giantswarm/shipyard/pkg/files"
	"github.com/giantswarm/shipyard/pkg/names"
	"github.com/giantswarm/shipyard/pkg/shipyard"
	"github.com/spf13/afero"
)

var (
	action = flag.String("action", "start", "action to accomplish, valid values 'start' and 'stop'")
	name   = flag.String("name", "", "build name, if none is provided a random one will be generated")
)

func main() {
	flag.Parse()

	if *action == "start" && *name == "" {
		*name = "shipyard-" + names.Rand(7)
	}

	logger, err := micrologger.New(micrologger.DefaultConfig())
	if err != nil {
		fmt.Println("Could not create logger: ", err)
		os.Exit(1)
	}

	engine := awsminikube.New(*name, awsminikube.DefaultConfig(), logger)

	filesHandler := files.NewHandler(afero.NewOsFs())

	sy := shipyard.New(logger, engine, filesHandler)

	switch *action {
	case "start":
		if err := sy.Start(); err != nil {
			logger.Log("error", fmt.Sprintf("Could not start shipyard: %v", err))
			os.Exit(1)
		}
	case "stop":
		if err := sy.Stop(); err != nil {
			logger.Log("error", fmt.Sprintf("Could not stop shipyard: %v", err))
			os.Exit(1)
		}
	default:
		logger.Log("error", fmt.Sprintf("unknown action %v", *action))
		flag.Usage()
		os.Exit(1)
	}
}
