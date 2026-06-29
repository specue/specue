// decision: internal/tech/binary-entrypoint
package main

import (
	"os"

	"github.com/specue/specue/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
