package main

import (
	"fmt"
	"os"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/server"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

var version = "0.0.1-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.Load()

	switch os.Args[1] {
	case "serve":
		if err := server.Serve(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "serve: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println(version)
	case "init":
		res, err := setup.Init(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "init: %v\n", err)
			os.Exit(1)
		}
		if res.AlreadyInitialized {
			fmt.Println("Dockyard already has a user account.")
			os.Exit(0)
		}
		setup.PrintInitBanner(res.Token, res.TokenExpiresAt, cfg.Listen)
	case "setup-token":
		token, exp, err := setup.IssueSetupToken(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "setup-token: %v\n", err)
			os.Exit(1)
		}
		setup.PrintInitBanner(token, exp, cfg.Listen)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: dockyard <serve|version|init|setup-token>\n")
}
