package util

import (
	"fmt"
	"os"
	"os/signal"
)

func WaitForInterrupt() {
	// Create a channel to receive signals.
	sigs := make(chan os.Signal, 1)

	// Register a handler for SIGINT.
	signal.Notify(sigs, os.Interrupt, os.Kill)

	// Start a goroutine to listen for signals.
	go func() {
		<-sigs
		fmt.Println("Received SIGINT signal. Exiting...")
		os.Exit(1)
	}()

	<-sigs
	fmt.Println("Unexpectedly exited.")
}
