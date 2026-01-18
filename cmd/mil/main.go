package main

import (
	"os"

	"github.com/suelio/millhouse/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
