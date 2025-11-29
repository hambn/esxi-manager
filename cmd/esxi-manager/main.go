package main

import (
	"os"

	"github.com/esxi-manager/esxi-manager/internal/common"
	"github.com/esxi-manager/esxi-manager/internal/interfaces/cli"
)

func main() {
	// Parse CLI flags
	params, commandName, err := cli.ParseFlags()
	if err != nil {
		cli.PrintUsage()
		common.Error("flag parsing failed", "error", err.Error())
		os.Exit(1)
	}

	// Create and execute CLI
	executor := cli.NewExecutor(params, commandName)
	if err := executor.Execute(); err != nil {
		common.Error("execution failed", "error", err.Error(), "command", commandName)
		os.Exit(1)
	}
}
