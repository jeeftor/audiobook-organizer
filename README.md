# Audiobook Organizer Docker Image

Docker container for organizing audiobooks based on metadata.json files. This container provides a CLI tool that organizes your audiobooks into a structured directory format.

## Quick Start

```bash
docker run -v /path/to/your/audiobooks:/books jeffsui/audiobook-organizer --dir=/books
```

## Volume Mounting

The container requires access to your audiobook files through a mounted volume:

```bash
# Basic usage with current directory
docker run -v $(pwd):/books jeffsui/audiobook-organizer --dir=/books

# Mount specific directory
docker run -v /path/to/audiobooks:/books jeffsui/audiobook-organizer --dir=/books

# Interactive mode with prompt
docker run -it -v /path/to/audiobooks:/books jeffsui/audiobook-organizer --dir=/books --prompt
```

## Container Parameters

- `-v`: Mount your audiobook directory
- `-it`: Required for interactive mode (when using --prompt)

## CLI Options

- `--dir`: Source directory to scan (required, must match mounted volume path)
- `--out`: Output directory for organized files (optional, defaults to --dir)
- `--replace_space`: Character to replace spaces in filenames
- `--dry-run`: Preview changes without moving files
- `--verbose`: Show detailed progress
- `--prompt`: Interactive mode to confirm moves
- `--undo`: Restore previous organization

## Directory Structure

The organizer creates the following structure:

```
/books/
  ├── Author Name/
  │   ├── Series Name #1/
  │   │   └── Book Title/
  │   └── Standalone Book/
  └── Multiple Authors/
      └── Collaboration Book/
```

## Metadata Format

Requires metadata.json files in book directories:

```json
{
  "authors": ["Author Name"],
  "title": "Book Title",
  "series": ["Series Name #1"]
}
```

## Security Notes

- The container runs with minimal privileges
- Use `:ro` mount flag for read-only access to source directories
- Consider using separate source and destination mounts for added safety

## Examples

Read-only source with separate output:
```bash
docker run \
  -v /source/audiobooks:/source:ro \
  -v /output/audiobooks:/output \
  jeffsui/audiobook-organizer --dir=/output
```

Dry run to preview changes:
```bash
docker run -v /path/to/audiobooks:/books \
  jeffsui/audiobook-organizer --dir=/books --dry-run
```

## Support

For issues and feature requests, please visit the [GitHub repository](https://github.com/yourusername/audiobook-organizer).