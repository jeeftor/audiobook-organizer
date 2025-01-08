# Audiobook Organizer

CLI tool to organize audiobooks based on metadata.json files.

## Features

- Organizes audiobooks by author/title structure
- Handles multiple authors
- Dry-run mode to preview changes
- Undo functionality
- Colored output
- Verbose logging
- Operation logs for recovery

## Installation

```bash
go install github.com/yourusername/audiobook-organizer@latest
```

## Usage

Basic organization:
```bash
audiobook-organizer --dir=/path/to/audiobooks
```

Options:
- `--dir`: Base directory (required)
- `--replace_space`: Character to replace spaces (default ".")
- `--dry-run`: Preview changes without moving files
- `--verbose`: Show detailed progress
- `--undo`: Restore files to original locations

## Metadata Format

Expects metadata.json files with structure:
```json
{
  "authors": ["Author Name"],
  "title": "Book Title"
}
```

## Example

```bash
# Preview changes
audiobook-organizer --dir=/audiobooks --dry-run --verbose

# Execute organization
audiobook-organizer --dir=/audiobooks

# Undo changes
audiobook-organizer --dir=/audiobooks --undo
```

## Recovery

Operations are logged to `.abs-org.log`. Use `--undo` to restore files to their original locations.