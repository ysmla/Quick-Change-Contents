# Quick-Change-Dict

A cross-platform command-line quick directory navigation tool. Save frequently used directories as shortcuts and jump to them instantly via clipboard.

[中文版](zhREADME.md)

## Install

```bash
# Clone and build
git clone https://github.com/ysmla/Quick-Change-Dict.git
cd Quick-Change-Dict
go build -o q .

# Add to PATH (optional, for global access)
# Add the project directory to your system PATH environment variable
```

## Commands

| Command | Usage | Description |
|---------|-------|-------------|
| `ls` | `q ls` | List files in the current directory with colored output |
| `cd` | `q cd <index>` | Copy a `cd <path>` command to clipboard by file index |
| `mi` | `q mi <name>` | Save the current directory as a shortcut (no duplicate names allowed) |
| `cg` | `q cg <name>` | Copy a saved shortcut's `cd` command to clipboard |
| `sh` | `q sh` | Show all saved shortcuts |
| `de` | `q de <name>` | Delete a saved shortcut |
| `fi` | `q fi <name>` | Find and display a saved shortcut |

## Examples

```bash
# List current directory
> q ls
  0 : Documents
  1 : Projects
  2 : Downloads

# Save a shortcut
> q mi projects        # saves current dir as "projects"

# Use a shortcut (copies cd command to clipboard)
> q cg projects
已将 cd /home/user/code/Projects 复制到剪贴板

# Show all shortcuts
> q sh
  0 : projects: /home/user/code/Projects
  1 : downloads: /home/user/Downloads

# Delete a shortcut
> q de downloads
已删除键 "downloads"

# Find a shortcut
> q fi projects
projects: /home/user/code/Projects
```

## Data Storage

All shortcuts are stored in `data.json` located next to the executable. The file is automatically created on first run.
