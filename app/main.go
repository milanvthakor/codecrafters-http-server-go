package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codecrafters-io/http-server-starter-go/app/config"
	"github.com/codecrafters-io/http-server-starter-go/app/server"
)

func main() {
	// Parse command-line flags
	directory := flag.String("directory", "", "The directory from which files should be served")
	port := flag.String("port", "4221", "The port on which the server should listen")
	flag.Parse()

	// Create configuration
	cfg := config.NewConfig(*directory, *port)

	// Create and start server
	srv := server.NewServer(cfg)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
