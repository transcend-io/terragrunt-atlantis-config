package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/transcend-io/terragrunt-atlantis-config/cmd"
)

// This variable is set at build time using -ldflags parameters.
// But we still set a default here for those using plain `go get` downloads
// For more info, see: http://stackoverflow.com/a/11355611/483528
var VERSION string = "1.20.0"

func main() {
	// Create context that can be cancelled by signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start signal handler in goroutine
	go func() {
		_ = <-sigChan
		// Log which signal was received (you can uncomment if needed)
		// fmt.Printf("Received signal: %v, initiating graceful shutdown...\n", sig)

		// Cancel context to signal all operations to stop
		cancel()

		// If we receive another signal, force exit
		go func() {
			<-sigChan
			os.Exit(1)
		}()
	}()

	// Execute with context
	cmd.Execute(ctx, VERSION)
}
