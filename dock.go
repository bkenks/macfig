package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"howett.net/plist"
)

const dockDomain = "com.apple.dock"

// Dock sections. "apps" is the left/top side of the divider, "others" is the
// right/bottom side (Downloads, Trash neighbourhood).
const (
	sectionApps   = "persistent-apps"
	sectionOthers = "persistent-others"
)

func sectionName(section string) string {
	if section == sectionOthers {
		return "right side"
	}
	return "left side"
}

// dockAppend appends raw plist tile values to a Dock section in one
// `defaults write ... -array-add v1 v2 ...` call.
func dockAppend(section string, tiles []string) error {
	if len(tiles) == 0 {
		return nil
	}
	args := append([]string{"write", dockDomain, section, "-array-add"}, tiles...)
	return run(nil, "defaults", args...)
}

// dockPrefs is the whole com.apple.dock domain, decoded.
type dockPrefs map[string]any

func readDock() (dockPrefs, error) {
	raw, err := output("defaults", "export", dockDomain, "-")
	if err != nil {
		return nil, err
	}
	var prefs dockPrefs
	if _, err := plist.Unmarshal(raw, &prefs); err != nil {
		return nil, fmt.Errorf("parsing Dock preferences: %w", err)
	}
	return prefs, nil
}

// tiles returns the tiles in a section (nil if the section is missing).
func (p dockPrefs) tiles(section string) []any {
	arr, _ := p[section].([]any)
	return arr
}

// writeDock replaces the whole com.apple.dock domain. The current domain is
// backed up first so a bad write can be undone with `defaults import`.
func writeDock(p dockPrefs) error {
	data, err := plist.MarshalIndent(map[string]any(p), plist.XMLFormat, "\t")
	if err != nil {
		return fmt.Errorf("encoding Dock preferences: %w", err)
	}
	if !opts.dryRun {
		backup, err := backupDock()
		if err != nil {
			return fmt.Errorf("backing up Dock preferences (nothing changed): %w", err)
		}
		fmt.Println("Backed up current Dock settings to", backup)
	}
	return run(data, "defaults", "import", dockDomain, "-")
}

func backupDock() (string, error) {
	raw, err := output("defaults", "export", dockDomain, "-")
	if err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Library", "Application Support", "macfig", "backups")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "dock-"+time.Now().Format("20060102-150405")+".plist")
	return path, os.WriteFile(path, raw, 0o644)
}

// restartDock applies pending Dock changes, unless --no-restart was given.
func restartDock() error {
	if opts.noRestart {
		fmt.Println("Skipping Dock restart (--no-restart). Run `killall Dock` to apply.")
		return nil
	}
	if err := run(nil, "killall", "Dock"); err != nil {
		return fmt.Errorf("restarting Dock: %w", err)
	}
	if !opts.dryRun {
		fmt.Println("Restarted Dock.")
	}
	return nil
}

// Tile helpers ---------------------------------------------------------------

func tileType(t any) string {
	m, _ := t.(map[string]any)
	s, _ := m["tile-type"].(string)
	return s
}

func tileData(t any) map[string]any {
	m, _ := t.(map[string]any)
	d, _ := m["tile-data"].(map[string]any)
	return d
}

// tileLabel gives a human-readable name for any Dock tile.
func tileLabel(t any) string {
	typ := tileType(t)
	if k, ok := spacerByTileType(typ); ok {
		return k.name + " spacer"
	}
	if typ == recentsTileType {
		if k, ok := recentsByListType(listType(t)); ok {
			return "recents stack: " + k.label
		}
		return "recents stack"
	}
	if label, _ := tileData(t)["file-label"].(string); label != "" {
		return label
	}
	return typ
}

// listType reads tile-data.list-type, whatever integer type it decoded as.
func listType(t any) int {
	switch v := tileData(t)["list-type"].(type) {
	case uint64:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}
