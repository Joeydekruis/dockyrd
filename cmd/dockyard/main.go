package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Joeydekruis/dockyrd/internal/admin"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/hostops"
	"github.com/Joeydekruis/dockyrd/internal/doctor"
	"github.com/Joeydekruis/dockyrd/internal/server"
	"github.com/Joeydekruis/dockyrd/internal/setup"
	"github.com/Joeydekruis/dockyrd/internal/updater"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.Load()

	switch os.Args[1] {
	case "serve":
		server.SetBuildVersion(version)
		if err := server.Serve(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "serve: %v\n", err)
			os.Exit(1)
		}
	case "version", "--version", "-version":
		fmt.Println(version)
	case "init":
		res, err := setup.Init(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "init: %v\n", err)
			os.Exit(1)
		}
		if res.AlreadyInitialized {
			fmt.Println("dockyrd already has a user account.")
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
	case "doctor":
		rep := doctor.Run(context.Background(), cfg, version)
		doctor.Print(os.Stdout, rep)
		if rep.HasFailures() {
			os.Exit(1)
		}
	case "repair-edge":
		if err := server.RepairEdgeCLI(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "repair-edge: %v\n", err)
			os.Exit(1)
		}
	case "admin":
		if err := runAdmin(cfg, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "admin: %v\n", err)
			os.Exit(1)
		}
	case "host":
		if err := runHost(cfg, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "host: %v\n", err)
			os.Exit(1)
		}
	case "update":
		if err := runUpdate(cfg, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "update: %v\n", err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func runAdmin(cfg config.Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dockyrd admin <reset-password|disable-2fa> <email> [new-password]")
	}
	switch args[0] {
	case "reset-password":
		if len(args) < 2 {
			return fmt.Errorf("email required")
		}
		email := args[1]
		password := ""
		if len(args) >= 3 {
			password = args[2]
		} else {
			var err error
			password, err = admin.ReadPasswordLine("New password: ")
			if err != nil {
				return err
			}
		}
		if err := admin.ResetPassword(cfg, email, password); err != nil {
			return err
		}
		fmt.Println("Password updated.")
		return nil
	case "disable-2fa":
		if len(args) < 2 {
			return fmt.Errorf("email required")
		}
		if err := admin.Disable2FA(cfg, args[1]); err != nil {
			return err
		}
		fmt.Println("Two-factor authentication disabled.")
		return nil
	default:
		return fmt.Errorf("unknown admin command %q", args[0])
	}
}

func runHost(cfg config.Config, args []string) error {
	if len(args) < 2 || args[0] != "revert" {
		return fmt.Errorf("usage: dockyrd host revert <action-id>")
	}
	return hostops.RevertByID(context.Background(), cfg, args[1])
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: dockyrd <serve|version|init|setup-token|doctor|repair-edge|update|admin|host>\n")
	fmt.Fprintf(os.Stderr, "       dockyrd update [--check] [--rollback]\n")
	fmt.Fprintf(os.Stderr, "       dockyrd admin reset-password <email> [password]\n")
	fmt.Fprintf(os.Stderr, "       dockyrd admin disable-2fa <email>\n")
	fmt.Fprintf(os.Stderr, "       dockyrd host revert <action-id>\n")
}

func runUpdate(cfg config.Config, args []string) error {
	ctx := context.Background()
	opts := updater.Options{Repo: os.Getenv("dockyrd_GITHUB_REPO")}
	if len(args) == 0 {
		_, err := updater.Apply(ctx, cfg, version, opts)
		return err
	}
	for _, a := range args {
		switch a {
		case "--check":
			res, err := updater.Check(ctx, version, opts)
			if err != nil {
				return err
			}
			if res.Newer {
				fmt.Printf("Update available: %s → %s (%s)\n", res.Current, res.Latest.Tag, res.Latest.AssetName)
			} else {
				fmt.Printf("Up to date (%s)\n", res.Current)
			}
			return nil
		case "--rollback":
			dest := os.Getenv("dockyrd_BIN")
			if dest == "" {
				dest = "/usr/local/bin/dockyrd"
			}
			if err := updater.RollbackBinary(dest); err != nil {
				return err
			}
			fmt.Println("Restored previous binary. Restart dockyrd.service.")
			return nil
		case "--dry-run":
			opts.DryRun = true
		default:
			return fmt.Errorf("unknown flag %q", a)
		}
	}
	_, err := updater.Apply(ctx, cfg, version, opts)
	return err
}
