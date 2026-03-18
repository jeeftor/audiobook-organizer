# Audiobook Organizer GUI

Desktop application for organizing audiobooks with a visual interface.

**Full documentation:** See [docs/GUI.md](../docs/GUI.md) in the main project.

## Quick Start

```bash
# Development mode with hot reload
wails dev

# Build production binary
wails build
```

## Project Structure

```
audiobook-organizer-gui/
├── app.go          # Go backend (Wails bindings)
├── main.go         # Application entry point
├── wails.json      # Wails configuration
├── frontend/       # React + TypeScript UI
│   ├── src/
│   │   ├── App.tsx
│   │   └── components/
│   └── wailsjs/    # Auto-generated TypeScript bindings
└── build/          # Build configurations and outputs
```

## Development

- **Frontend:** React + TypeScript + Tailwind CSS
- **Backend:** Go with Wails v2 bindings
- **Build:** `wails build` creates platform-specific binaries

For detailed development instructions, see [CLAUDE.md](../CLAUDE.md).
