// Command specue is the CLI entry point: it wires the real process streams and
// arguments into the cli package and exits with the code the dispatch returns
// (0 clean, 1 a gate fired, 2 a usage/resolution error). All command logic lives
// in the cli package so it stays testable with injected streams; main only bridges
// to the OS.
package main

import (
	"os"

	"github.com/specue/specue/internal/cli"
)

func main() {
	os.Exit(cli.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
