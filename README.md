# macfig

A helper for macOS configuration. It wraps the `defaults write` tweaks from
`docs/mac-os/configuration.md` so you don't have to remember them. Right now it only handles the Dock.

```
go install github.com/bkenks/macfig@latest   # or, from a checkout: go build -o macfig .
```

Every command that changes the Dock makes all its changes first, then restarts
the Dock once (`killall Dock`).

Global flags:
- `-n, --dry-run` prints the commands without running them.
- `--no-restart` skips the Dock restart.

## Spacers

| Kind    | Flag          | Tile                | Looks like                        |
|---------|---------------|---------------------|-----------------------------------|
| small   | `-s, --small` | `small-spacer-tile` | half-width blank space            |
| large   | `-l, --large` | `spacer-tile`       | full-width blank space            |
| flex    | `-f, --flex`  | `flex-spacer-tile`  | stretches to fill extra room      |

```sh
macfig dock spacer add --small 5            # 5 small spacers, one Dock restart
macfig dock spacer add -s 2 -l 1 -f 1       # appended in the order small, large, flex
macfig dock spacer add --flex 1 --right     # right side of the divider

macfig dock spacer list                     # where each spacer is and what it sits between
macfig dock spacer remove --small 2         # removes the right-most 2 small spacers
macfig dock spacer remove --flex all
macfig dock spacer remove --all
```

New spacers are added at the end of the Dock. Drag them where you want them.

## Recents stacks

These are auto-filling stacks on the right side of the Dock.

```sh
macfig dock recents add apps documents      # kinds: apps, documents, servers, volumes, items
macfig dock recents list
macfig dock recents remove documents
macfig dock recents remove --all
```

## Other

```sh
macfig dock list        # every tile in order
macfig dock restart     # killall Dock
```

## How it works

**Adding** uses a single `defaults write com.apple.dock persistent-apps -array-add …`
call that contains every tile.

**Removing** works in four steps:
1. Read the domain with `defaults export`.
2. Filter out the matching tiles.
3. Write the domain back with `defaults import`.
4. Restart the Dock.

Before each removal, the current settings are backed up to
`~/Library/Application Support/macfig/backups/`. To undo a removal:

```sh
defaults import com.apple.dock ~/Library/Application\ Support/macfig/backups/dock-<timestamp>.plist && killall Dock
```

## Layout

- `main.go` sets up the root command and global flags.
- `dockcmd.go` has `dock list` and `dock restart`.
- `spacer.go` and `recents.go` hold the spacer and recents commands.
- `dock.go` reads and writes the Dock prefs, handles backups and restarts the Dock.
- `run.go` runs commands and handles dry-run.
- `count.go` handles count flags (`N` or `all`).
