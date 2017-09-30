package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/giantswarm/shipyard/pkg/shipyard"
)

func main() {
	sigs := make(chan os.Signal)
	done := make(chan struct{})

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		close(done)
	}()

	sy, err := shipyard.New()
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
