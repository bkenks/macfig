package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const recentsTileType = "recents-tile"

type recentsKind struct {
	name     string
	aliases  []string
	listType int
	label    string
}

var recentsKinds = []recentsKind{
	{"apps", []string{"applications"}, 1, "Recent Applications"},
	{"documents", []string{"docs"}, 2, "Recent Documents"},
	{"servers", nil, 3, "Recent Servers"},
	{"volumes", nil, 4, "Favorite Volumes"},
	{"items", []string{"favorites"}, 5, "Favorite Items"},
}

func recentsByListType(n int) (recentsKind, bool) {
	for _, k := range recentsKinds {
		if k.listType == n {
			return k, true
		}
	}
	return recentsKind{}, false
}

func recentsByName(s string) (recentsKind, error) {
	s = strings.ToLower(s)
	for _, k := range recentsKinds {
		if k.name == s {
			return k, nil
		}
		for _, a := range k.aliases {
			if a == s {
				return k, nil
			}
		}
	}
	return recentsKind{}, fmt.Errorf("unknown recents kind %q (want one of: %s)", s, recentsKindNames())
}

func recentsKindNames() string {
	names := make([]string, len(recentsKinds))
	for i, k := range recentsKinds {
		names[i] = k.name
	}
	return strings.Join(names, ", ")
}

func (k recentsKind) tileValue() string {
	return fmt.Sprintf(`{"tile-data"={"list-type"=%d;}; "tile-type"="%s";}`, k.listType, recentsTileType)
}

func recentsKindsHelp() string {
	var b strings.Builder
	for _, k := range recentsKinds {
		fmt.Fprintf(&b, "  %-10s %s\n", k.name, k.label)
	}
	return b.String()
}

func newRecentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recents",
		Short: "Add, remove and list auto-filling recents stacks",
		Long: `Recents stacks are folder-like stacks on the right side of the Dock that fill
themselves automatically.

Kinds:
` + recentsKindsHelp(),
	}
	cmd.AddCommand(newRecentsAddCmd(), newRecentsRemoveCmd(), newRecentsListCmd())
	return cmd
}

func newRecentsAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add KIND...",
		Short: "Add one recents stack per KIND to the right side of the Dock",
		Example: `  macfig dock recents add apps
  macfig dock recents add apps documents`,
		Args:      cobra.MinimumNArgs(1),
		ValidArgs: []string{"apps", "documents", "servers", "volumes", "items"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var tiles, labels []string
			for _, a := range args {
				k, err := recentsByName(a)
				if err != nil {
					return err
				}
				tiles = append(tiles, k.tileValue())
				labels = append(labels, k.label)
			}
			if err := dockAppend(sectionOthers, tiles); err != nil {
				return err
			}
			if !opts.dryRun {
				fmt.Printf("Added recents stack(s): %s.\n", strings.Join(labels, ", "))
			}
			return restartDock()
		},
	}
}

func newRecentsRemoveCmd() *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:     "remove [KIND...]",
		Aliases: []string{"rm"},
		Short:   "Remove recents stacks of the given kinds (or --all)",
		Example: `  macfig dock recents remove documents
  macfig dock recents remove --all`,
		ValidArgs: []string{"apps", "documents", "servers", "volumes", "items"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if all == (len(args) > 0) {
				return errors.New("pass either one or more kinds or --all")
			}
			match := map[int]bool{}
			for _, a := range args {
				k, err := recentsByName(a)
				if err != nil {
					return err
				}
				match[k.listType] = true
			}

			prefs, err := readDock()
			if err != nil {
				return err
			}
			tiles := prefs.tiles(sectionOthers)
			kept := make([]any, 0, len(tiles))
			var removed []string
			for _, t := range tiles {
				if tileType(t) == recentsTileType && (all || match[listType(t)]) {
					removed = append(removed, tileLabel(t))
					continue
				}
				kept = append(kept, t)
			}
			if len(removed) == 0 {
				fmt.Println("No matching recents stacks in the Dock; nothing changed.")
				return nil
			}
			prefs[sectionOthers] = kept
			if err := writeDock(prefs); err != nil {
				return err
			}
			verb := "Removed"
			if opts.dryRun {
				verb = "Would remove"
			}
			fmt.Printf("%s: %s.\n", verb, strings.Join(removed, ", "))
			return restartDock()
		},
	}
	cmd.Flags().BoolVarP(&all, "all", "a", false, "remove every recents stack")
	return cmd
}

func newRecentsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List recents stacks in the Dock",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prefs, err := readDock()
			if err != nil {
				return err
			}
			found := false
			for idx, t := range prefs.tiles(sectionOthers) {
				if tileType(t) == recentsTileType {
					fmt.Printf("  %3d  %s\n", idx+1, tileLabel(t))
					found = true
				}
			}
			if !found {
				fmt.Println("No recents stacks in the Dock.")
			}
			return nil
		},
	}
}
