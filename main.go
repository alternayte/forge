package main

import (
	"os"

	"github.com/forge-framework/forge/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
