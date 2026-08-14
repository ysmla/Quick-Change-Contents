# Quick-Change-Contents

A cross-platform command-line quick directory navigation tool. Save frequently used directories as shortcuts and jump to them instantly via clipboard.

[中文版](zhREADME.md)

## Install

```bash
# Clone and build
git clone https://github.com/ysmla/Quick-Change-Contents.git
cd Quick-Change-Contents
go build -o q .

# Add to PATH (optional, for global access)
# Add the project directory to your system PATH environment variable
```

## Commands

| Command | Usage | Description |
|---------|-------|-------------|
| `ls` | `q ls` | List entries in the current directory (directories are marked with a trailing `/`) |
| `cd` | `q cd <index>` | Copy a `cd <path>` command to clipboard by entry index (paths with spaces are quoted) |
| `mi` | `q mi <name>` | Save the current directory as a shortcut (no duplicate names allowed) |
| `cg` | `q cg <name>` | Copy a saved shortcut's `cd` command to clipboard |
| `sh` | `q sh` | Show all saved shortcuts, sorted by name |
| `de` | `q de <name>` | Delete a saved shortcut |
| `fi` | `q fi <name>` | Find and display a saved shortcut |
| `help` | `q help` | Show usage help |

## Examples

```bash
# List current directory (directories end with /)
> q ls
  0 : Documents/
  1 : Projects/
  2 : Downloads/

# Save a shortcut
> q mi projects        # saves current dir as "projects"

# Use a shortcut (copies cd command to clipboard)
> q cg projects
已将 cd "/home/user/code/Projects" 复制到剪贴板

# Show all shortcuts (sorted by name)
> q sh
downloads   : /home/user/Downloads
projects    : /home/user/code/Projects

# Delete a shortcut
> q de downloads
已删除键 "downloads"

# Find a shortcut
> q fi projects
projects: /home/user/code/Projects
```

## Data Storage

Shortcuts are stored in `data.json` in the **user config directory**:

- Windows: `%AppData%\q\data.json`
- Linux/macOS: `~/.config/q/data.json` (or `$XDG_CONFIG_HOME/q/data.json`)

This works even when the executable is installed in a read-only location. The file is created automatically on first run. To override the location, set the `Q_DATA_DIR` environment variable to any writable directory.

> **Migration:** if you used an older version that stored `data.json` next to the executable, the file is automatically copied to the new location on the first run.

If `data.json` becomes corrupted (invalid JSON), it is automatically backed up as `broken_data_<timestamp>.json` and a fresh file is created — your data is never silently lost.
