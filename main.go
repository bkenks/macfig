// macfig is a small helper for macOS configuration tweaks that otherwise live
// in a notes file full of `defaults write` one-liners.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Global options shared by every command that changes something.
var opts struct {
	dryRun    bool
	noRestart bool
}

func main() {
	root := &cobra.Command{
		Use:           "macfig",
		Short:         "Helper for macOS configuration",
		Long:          "macfig wraps the fiddly `defaults write` tweaks for macOS so you don't have to remember them.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().BoolVarP(&opts.dryRun, "dry-run", "n", false, "print the commands instead of running them")
	root.PersistentFlags().BoolVar(&opts.noRestart, "no-restart", false, "don't restart the affected app (e.g. Dock) after changing it")

	root.AddCommand(newDockCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "macfig:", err)
		os.Exit(1)
	}
}
