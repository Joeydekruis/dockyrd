package main

import (
	"fmt"
	"os"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/server"
)

var version = "0.0.1-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		cfg := config.Load()
		if err := server.Serve(cfg.Listen); err != nil {
			fmt.Fprintf(os.Stderr, "serve: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println(version)
	case "init":
		fmt.Println("init: not implemented yet (M0 Sprint 3)")
		os.Exit(2)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: dockyard <serve|version|init>\n")
}
