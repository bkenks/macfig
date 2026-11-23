package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dock",
		Short: "Configure the Dock (spacers, recents stacks)",
	}
	cmd.AddCommand(newSpacerCmd(), newRecentsCmd(), newDockListCmd(), newDockRestartCmd())
	return cmd
}

func newDockListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List every tile in the Dock, in order",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prefs, err := readDock()
			if err != nil {
				return err
			}
			for i, section := range []string{sectionApps, sectionOthers} {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("%s:\n", capitalize(sectionName(section)))
				tiles := prefs.tiles(section)
				if len(tiles) == 0 {
					fmt.Println("  (empty)")
				}
				for idx, t := range tiles {
					fmt.Printf("  %3d  %s\n", idx+1, tileLabel(t))
				}
			}
			return nil
		},
	}
}

func newDockRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart the Dock (killall Dock) to apply pending changes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.noRestart = false
			return restartDock()
		},
	}
}
