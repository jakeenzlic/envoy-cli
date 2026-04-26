// main is the entry point for the envoy-cli application.
// It wires together all CLI subcommands and executes the root command.
package main

import (
	"fmt"
	"os"

	"github.com/envoy-cli/envoy/internal/cli"
)

func main() {
	root := cli.NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
