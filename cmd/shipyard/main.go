package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/giantswarm/shipyard/pkg/shipyard"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	workDir, err := ioutil.TempDir("", "gs-shipyard")
	if err != nil {
		fmt.Printf("Could not create working directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(workDir)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	sy, err := shipyard.New(workDir)
	if err != nil {
		fmt.Printf("Could not create shipyard: %v\n", err)
		os.Exit(1)
	}

	if err := sy.Start(); err != nil {
		fmt.Printf("Could not start shipyard: %v\n", err)
		os.Exit(1)
	}

	<-done
	fmt.Println("exiting")
	if err := sy.Stop(); err != nil {
		fmt.Printf("Could not stop shipyard: %v\n", err)
		os.Exit(1)
	}
}
