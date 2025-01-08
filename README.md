# Audiobook Organizer

CLI tool to organize audiobooks based on metadata.json files.

## Features

- Organizes audiobooks by author/series/title structure
- Handles multiple authors
- Preserves spaces by default
- Optional space replacement with custom character
- Dry-run mode to preview changes
- Interactive prompt mode for reviewing moves
- Undo functionality
- Colored output
- Operation logs for recovery

## Installation

### Go Install
```bash
go install github.com/yourusername/audiobook-organizer@latest
```

### Docker
```bash
docker pull jeffsui/audiobook-organizer:latest

docker pull jeffsui/audiobook-organizer:latest

# Volume Mounting
When running the Docker container, you need to mount your audiobook directory to make it accessible to the container. The container will process and organize books within this mounted directory.

Examples:

# Mount and process current directory
docker run -v $(pwd):/books jeffsui/audiobook-organizer --dir=/books

# Mount a specific audiobook directory
docker run -v /path/to/your/audiobooks:/books jeffsui/audiobook-organizer --dir=/books

# Mount with prompt mode (interactive)
docker run -it -v /path/to/your/audiobooks:/books jeffsui/audiobook-organizer --dir=/books --prompt

# Mount read-only source directory and separate output directory
docker run \
-v /path/to/source/audiobooks:/source:ro \
-v /path/to/output/audiobooks:/output \
jeffsui/audiobook-organizer --dir=/output
```

Notes:
- The container path (/books in examples) must match the --dir parameter
- Use `-it` flag when running with `--prompt` for interactive mode
- Add `:ro` to source volume mount for read-only access if desired
- Multiple directories can be mounted for source/destination separation




## Usage

Basic organization:
```bash
audiobook-organizer --dir=/path/to/audiobooks
```

Options:
- `--dir`: Base directory (required)
- `--replace_space`: Character to replace spaces (optional)
- `--dry-run`: Preview changes without moving files
- `--verbose`: Show detailed progress
- `--undo`: Restore files to original locations
- `--prompt`: Review and confirm each book move interactively

### Interactive Mode

Using the `--prompt` flag will show each book's details and proposed move location:

```
Book found:
  Title: The Book Title
  Authors: Author One, Author Two
  Series: Amazing Series #1

Proposed move:
  From: /original/path/book
  To: /audiobooks/Author One,Author Two/Amazing Series #1/The Book Title

Proceed with move? [y/N]
```

## Metadata Format

Expects metadata.json files with structure:
```json
{
  "authors": ["Author Name"],
  "title": "Book Title",
  "series": ["Series Name #1"]
}
```

## Directory Structure

Without series:
```
/audiobooks/Author Name/Book Title/
```

With series:
```
/audiobooks/Author Name/Series Name #1/Book Title/
```

With multiple authors:
```
/audiobooks/Author One,Author Two/Book Title/
```

With space replacement (--replace_space="."):
```
/audiobooks/Author.Name/Series.Name.#1/Book.Title/
```

## Recovery

Operations are logged to `.abook-org.log`. Use `--undo` to restore files to their original locations.