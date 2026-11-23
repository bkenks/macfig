package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type spacerKind struct {
	name     string // flag name, e.g. "small"
	short    string // one-letter flag
	tileType string
	desc     string
}

var spacerKinds = []spacerKind{
	{"small", "s", "small-spacer-tile", "half-width blank space"},
	{"large", "l", "spacer-tile", "full-width blank space"},
	{"flex", "f", "flex-spacer-tile", "flexible space that stretches to fill extra room"},
}

func spacerByTileType(t string) (spacerKind, bool) {
	for _, k := range spacerKinds {
		if k.tileType == t {
			return k, true
		}
	}
	return spacerKind{}, false
}

func (k spacerKind) tileValue() string {
	return fmt.Sprintf(`{"tile-type"="%s";}`, k.tileType)
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

func newSpacerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "spacer",
		Aliases: []string{"spacers"},
		Short:   "Add, remove and list Dock spacers",
		Long: `Add, remove and list Dock spacers.

Kinds:
  small  half-width blank space        (small-spacer-tile)
  large  full-width blank space        (spacer-tile)
  flex   stretches to fill extra room  (flex-spacer-tile)

Spacers go on the left side of the Dock divider (with the apps) unless
--right is given. The Dock is restarted once after all changes are made.`,
	}
	cmd.AddCommand(newSpacerAddCmd(), newSpacerRemoveCmd(), newSpacerListCmd())
	return cmd
}

func newSpacerAddCmd() *cobra.Command {
	counts := make([]*countValue, len(spacerKinds))
	var right bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Append spacers to the end of the Dock",
		Example: `  macfig dock spacer add --small 5
  macfig dock spacer add -s 2 -l 1 -f 1
  macfig dock spacer add --flex 1 --right`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			section := sectionApps
			if right {
				section = sectionOthers
			}
			var tiles, summary []string
			for i, k := range spacerKinds {
				n := counts[i].n
				for range n {
					tiles = append(tiles, k.tileValue())
				}
				if n > 0 {
					summary = append(summary, plural(n, k.name+" spacer"))
				}
			}
			if len(tiles) == 0 {
				return errors.New("nothing to add: pass at least one of --small, --large or --flex with a count")
			}
			if err := dockAppend(section, tiles); err != nil {
				return err
			}
			if !opts.dryRun {
				fmt.Printf("Added %s to the %s of the Dock.\n", strings.Join(summary, ", "), sectionName(section))
			}
			return restartDock()
		},
	}
	for i, k := range spacerKinds {
		counts[i] = &countValue{}
		cmd.Flags().VarP(counts[i], k.name, k.short, "number of "+k.name+" spacers to add ("+k.desc+")")
	}
	cmd.Flags().BoolVarP(&right, "right", "r", false, "add to the right side of the divider instead of the apps side")
	return cmd
}

func newSpacerRemoveCmd() *cobra.Command {
	counts := make([]*countValue, len(spacerKinds))
	var right, all bool

	cmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "Remove spacers, starting from the end of the Dock",
		Long: `Remove spacers, starting from the end of the Dock (right-most first).

The current Dock settings are backed up to
~/Library/Application Support/macfig/backups before anything is changed.
Restore one with: defaults import com.apple.dock <backup.plist> && killall Dock`,
		Example: `  macfig dock spacer remove --small 2
  macfig dock spacer remove --flex all
  macfig dock spacer remove --all`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			want := make([]int, len(spacerKinds))
			requested := false
			for i := range spacerKinds {
				want[i] = counts[i].n
				if all {
					want[i] = countAll
				}
				requested = requested || want[i] != 0
			}
			if !requested {
				return errors.New("nothing to remove: pass --all, or --small/--large/--flex with a count or \"all\"")
			}

			section := sectionApps
			if right {
				section = sectionOthers
			}
			prefs, err := readDock()
			if err != nil {
				return err
			}
			tiles := prefs.tiles(section)

			// Walk right-to-left so the most recently added spacers go first.
			drop := make([]bool, len(tiles))
			removed := make([]int, len(spacerKinds))
			for idx := len(tiles) - 1; idx >= 0; idx-- {
				for i, k := range spacerKinds {
					if tileType(tiles[idx]) != k.tileType {
						continue
					}
					if want[i] == countAll || removed[i] < want[i] {
						drop[idx] = true
						removed[i]++
					}
				}
			}

			var summary []string
			for i, k := range spacerKinds {
				if removed[i] > 0 {
					summary = append(summary, plural(removed[i], k.name+" spacer"))
				}
				if want[i] > removed[i] {
					fmt.Printf("Note: asked for %d %s spacer(s) but the %s only has %d.\n",
						want[i], k.name, sectionName(section), removed[i])
				}
			}
			if len(summary) == 0 {
				fmt.Printf("No matching spacers on the %s of the Dock; nothing changed.\n", sectionName(section))
				return nil
			}

			kept := make([]any, 0, len(tiles))
			for idx, t := range tiles {
				if !drop[idx] {
					kept = append(kept, t)
				}
			}
			prefs[section] = kept
			if err := writeDock(prefs); err != nil {
				return err
			}
			if !opts.dryRun {
				fmt.Printf("Removed %s from the %s of the Dock.\n", strings.Join(summary, ", "), sectionName(section))
			} else {
				fmt.Printf("Would remove %s from the %s of the Dock.\n", strings.Join(summary, ", "), sectionName(section))
			}
			return restartDock()
		},
	}
	for i, k := range spacerKinds {
		counts[i] = &countValue{allowAll: true}
		cmd.Flags().VarP(counts[i], k.name, k.short, "number of "+k.name+" spacers to remove, or \"all\"")
	}
	cmd.Flags().BoolVarP(&all, "all", "a", false, "remove every spacer of every kind")
	cmd.Flags().BoolVarP(&right, "right", "r", false, "remove from the right side of the divider instead of the apps side")
	return cmd
}

func newSpacerListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Show where the spacers are in the Dock",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prefs, err := readDock()
			if err != nil {
				return err
			}
			for _, section := range []string{sectionApps, sectionOthers} {
				tiles := prefs.tiles(section)
				totals := map[string]int{}
				var lines []string
				for idx, t := range tiles {
					k, ok := spacerByTileType(tileType(t))
					if !ok {
						continue
					}
					totals[k.name]++
					lines = append(lines, fmt.Sprintf("  %3d  %s spacer%s", idx+1, k.name, neighbours(tiles, idx)))
				}
				fmt.Printf("%s (%d tiles):\n", capitalize(sectionName(section)), len(tiles))
				if len(lines) == 0 {
					fmt.Println("  no spacers")
				} else {
					fmt.Println(strings.Join(lines, "\n"))
					var parts []string
					for _, k := range spacerKinds {
						if totals[k.name] > 0 {
							parts = append(parts, fmt.Sprintf("%s %d", k.name, totals[k.name]))
						}
					}
					fmt.Println("  total:", strings.Join(parts, ", "))
				}
			}
			return nil
		},
	}
}

// neighbours describes what a spacer sits between, to make positions readable.
func neighbours(tiles []any, idx int) string {
	before, after := "start", "end"
	if idx > 0 {
		before = tileLabel(tiles[idx-1])
	}
	if idx < len(tiles)-1 {
		after = tileLabel(tiles[idx+1])
	}
	return fmt.Sprintf("   (between %s and %s)", before, after)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
