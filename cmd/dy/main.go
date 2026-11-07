package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/apiclient"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "login":
		runLogin(os.Args[2:])
	case "whoami":
		runWhoami()
	case "services":
		runServices(os.Args[2:])
	case "deploy":
		runDeploy(os.Args[2:])
	case "logs":
		runLogs(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `dy — dockyrd CLI

Usage:
  dy login <base-url> <token>
  dy whoami
  dy services <project-id>
  dy deploy <service-id>
  dy logs <service-id> [runtime|access] [tail]

Env: dockyrd_URL overrides saved base URL.
`)
}

func loadClient() (*apiclient.Client, error) {
	base, err := apiclient.LoadBaseURL()
	if err != nil {
		return nil, fmt.Errorf("base URL not configured (dy login or dockyrd_URL)")
	}
	tok, err := apiclient.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("not logged in (dy login)")
	}
	return apiclient.New(base, tok), nil
}

func runLogin(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dy login <base-url> <token>")
		os.Exit(1)
	}
	base := strings.TrimRight(strings.TrimSpace(args[0]), "/")
	token := strings.TrimSpace(args[1])
	if err := apiclient.SaveBaseURL(base); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := apiclient.SaveToken(token); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Saved credentials to ~/.config/dy")
}

func runWhoami() {
	c, err := loadClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out, err := c.Whoami()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func runServices(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: dy services <project-id>")
		os.Exit(1)
	}
	c, err := loadClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out, err := c.ListServices(strings.TrimSpace(args[0]))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func runDeploy(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: dy deploy <service-id>")
		os.Exit(1)
	}
	c, err := loadClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out, err := c.CreateDeployment(strings.TrimSpace(args[0]))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func runLogs(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: dy logs <service-id> [stream] [tail]")
		os.Exit(1)
	}
	stream := "runtime"
	tail := "200"
	if len(args) > 1 && args[1] != "" {
		stream = args[1]
	}
	if len(args) > 2 && args[2] != "" {
		tail = args[2]
	}
	c, err := loadClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	text, err := c.ServiceLogs(strings.TrimSpace(args[0]), stream, tail)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(text)
}
