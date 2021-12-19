package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/nohns/bingo-box/server/config"
)

// Entrypoint for server binary.
func main() {

	// Create shared app context and listen for interrupts
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go cancelCtxOnInterrupt(cancel, c)

	// Read app configuration
	conf, err := config.Read()
	if err != nil {
		fmt.Printf("%v\n\nfailed to run server application due to bad configuration\n", err)
		os.Exit(1)
	}

	// Instantiate server application and inject dependencies
	app := NewApp()
	app.Conf = conf

	// Bootstrap server
	if err := app.Bootstrap(ctx); err != nil {
		fmt.Printf("%v\n\nencountered an uncoverable error while bootstrapping server application\n", err)
		os.Exit(1)
	}

	// Run server
	errChan := make(chan error)
	go app.Run(errChan)

	// Wait a runtime error or for context to be cancelled. E.g HTTP serve error or CTRL+C interrupt
	select {
	case err := <-errChan:
		if err != nil {
			fmt.Printf("%v\n\nencountered an uncoverable error while running server application\n", err)
		}
		cancel()
	case <-ctx.Done():
	}

	// Cleanup server application
	if err := app.Close(); err != nil {
		logFatal("failed to close server application properly due to error", err)
	}
}

// If any interrupts is received on channel c, then call the context cancel function
func cancelCtxOnInterrupt(cancel context.CancelFunc, c chan os.Signal) {
	<-c
	cancel()
}

// Log a fatal error to stderr
func logFatal(msg string, err error) {
	log.Fatalf("main: %s: \n%v\n\n", msg, err)
}
