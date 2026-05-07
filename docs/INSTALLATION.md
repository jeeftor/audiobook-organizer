# Installation Guide

Complete installation instructions for all platforms and modes (GUI, TUI, CLI).

## Overview

Starting with v0.12.0, Audiobook Organizer uses a **unified binary** that includes all three modes:

| Command | Mode | Description |
|---------|------|-------------|
| `audiobook-organizer gui` | Desktop GUI | Modern desktop interface (requires GUI-enabled build) |
| `audiobook-organizer tui` | Interactive Terminal | Keyboard-driven terminal UI |
| `audiobook-organizer` | CLI | Command-line automation and scripting |

**Downloads:**
- **Unified builds** (`audiobook-organizer-*-unified.*`) include the GUI - recommended for most users
- **CLI-only builds** (`audiobook-organizer_*.*`) are smaller but lack GUI support
- **Standalone GUI** (`audiobook-organizer-gui*.*`) available for backward compatibility

Choose the installation method that fits your platform and preferred interface.

---

## Quick Install

### macOS

**Unified binary (GUI + CLI + TUI):**
```bash
# Download the unified DMG from releases (includes GUI)
# Or use Homebrew for CLI/TUI only
brew tap jeeftor/tap
brew install audiobook-organizer
```

**Launch modes:**
```bash
audiobook-organizer gui        # Desktop GUI
audiobook-organizer tui        # Interactive terminal
audiobook-organizer --dir=...  # CLI mode
```

### Linux (Debian/Ubuntu)

**Unified binary (recommended):**
```bash
# Download unified .deb from releases (includes GUI support)
wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer-unified_*_amd64.deb
sudo dpkg -i audiobook-organizer-unified_*.deb

# Launch GUI
audiobook-organizer gui
```

**CLI/TUI only (smaller, no GUI):**
```bash
sudo apt install audiobook-organizer
```

### Windows

**Unified binary (recommended):**
```powershell
# Download from releases:
# - audiobook-organizer-unified-windows-amd64.zip (GUI + CLI + TUI)
# Extract and run: audiobook-organizer.exe gui
```

**Standalone options:**
```powershell
# GUI only: audiobook-organizer-gui-setup.exe
# CLI/TUI only: audiobook-organizer-windows-amd64.zip
```

### Docker

```bash
docker pull jeffsui/audiobook-organizer:latest
```

---

## GUI Installation

The **Desktop GUI** provides a visual point-and-click interface. Available for macOS, Linux, and Windows.

### macOS

#### Option 1: Unified Binary DMG (Recommended)

1. Download `audiobook-organizer-universal.dmg` from [GitHub Releases](https://github.com/jeeftor/audiobook-organizer/releases)
2. Open the DMG file
3. Drag `Audiobook Organizer.app` to Applications folder
4. **First launch:** Right-click → Open (to bypass Gatekeeper)

**Launch the GUI:**
```bash
audiobook-organizer gui
```

#### Option 2: Homebrew Cask (Coming Soon)

```bash
brew install --cask audiobook-organizer
```

**Troubleshooting macOS:**
- If you see "unidentified developer" warning: Right-click → Open
- WebView2 is built into macOS, no additional dependencies needed

### Linux

#### Debian/Ubuntu - Unified Binary (Recommended)

```bash
# Download unified .deb file from releases (includes GUI, CLI, and TUI)
wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer-unified_*_amd64.deb

# Install
sudo dpkg -i audiobook-organizer-unified_*.deb

# Install dependencies if needed
sudo apt install -f

# Launch GUI
audiobook-organizer gui
```

#### RedHat/Fedora - Unified Binary

```bash
# Download unified .rpm from releases
wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer-unified_*.rpm

# Install
sudo rpm -i audiobook-organizer-unified_*.rpm

# Launch GUI
audiobook-organizer gui
```

#### Universal AppImage (All Distros)

```bash
# Download unified AppImage (includes all modes)
wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer-unified_*.AppImage

# Make executable
chmod +x audiobook-organizer-unified_*.AppImage

# Run GUI
./audiobook-organizer-unified_*.AppImage gui
```

**Dependencies:**

The GUI requires WebKit2GTK:

```bash
# Debian/Ubuntu
sudo apt install libwebkit2gtk-4.0-37

# Fedora
sudo yum install webkit2gtk3

# Arch
sudo pacman -S webkit2gtk
```

### Windows

#### Option 1: Unified Installer (Recommended)

1. Download `audiobook-organizer-unified-setup.exe` from [GitHub Releases](https://github.com/jeeftor/audiobook-organizer/releases)
2. Double-click the installer
3. Follow installation wizard
4. Launch GUI from Start Menu or run `audiobook-organizer.exe gui`

#### Option 2: Portable ZIP

1. Download `audiobook-organizer-unified-windows-amd64.zip` from [GitHub Releases](https://github.com/jeeftor/audiobook-organizer/releases)
2. Extract to desired location
3. Run `audiobook-organizer.exe gui`

**Dependencies:**

Windows 10/11 includes WebView2 by default. If you encounter issues on Windows 10:

1. Download [WebView2 Runtime](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)
2. Install and restart

### Verifying GUI Installation

```bash
# Check version
audiobook-organizer version

# Launch GUI
audiobook-organizer gui

# Test all modes
audiobook-organizer gui --help      # GUI options
audiobook-organizer tui --help      # TUI options
audiobook-organizer --help          # CLI options
```

> **Standalone GUI (Backward Compatibility):** The separate `audiobook-organizer-gui` binary is still available for users who prefer it, but the unified binary is recommended for new installations.

---

## CLI/TUI Installation

The **CLI/TUI binary** provides command-line and interactive terminal interfaces. Ideal for automation, scripting, and SSH sessions.

### macOS

#### Option 1: Homebrew (Recommended)

```bash
# Add tap
brew tap jeeftor/tap

# Install
brew install audiobook-organizer

# Update
brew upgrade audiobook-organizer
```

#### Option 2: Binary Download

```bash
# Download for your architecture
# Intel Macs (amd64)
wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer_Darwin_x86_64.tar.gz

# Apple Silicon (arm64)
wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer_Darwin_arm64.tar.gz

# Extract
tar -xzf audiobook-organizer_Darwin_*.tar.gz

# Move to PATH
sudo mv audiobook-organizer /usr/local/bin/

# Verify
audiobook-organizer version
```

### Linux

#### Option 1: Package Manager (Recommended)

**Debian/Ubuntu (APT):**

```bash
# Add repository (one-time setup)
echo "deb [trusted=yes] https://apt.fury.io/jeeftor/ /" | sudo tee /etc/apt/sources.list.d/audiobook-organizer.list

# Update and install
sudo apt update
sudo apt install audiobook-organizer

# Update
sudo apt upgrade audiobook-organizer
```

**RedHat/Fedora/CentOS (YUM/DNF):**

```bash
# Add repository (one-time setup)
sudo tee /etc/yum.repos.d/audiobook-organizer.repo <<EOF
[audiobook-organizer]
name=Audiobook Organizer Repository
baseurl=https://yum.fury.io/jeeftor/
enabled=1
gpgcheck=0
EOF

# Install
sudo yum install audiobook-organizer
# or
sudo dnf install audiobook-organizer

# Update
sudo yum upgrade audiobook-organizer
```

**Alpine (APK):**

```bash
# Add repository
echo "https://apk.fury.io/jeeftor/" | sudo tee -a /etc/apk/repositories

# Update and install
sudo apk update
sudo apk add audiobook-organizer

# Update
sudo apk upgrade audiobook-organizer
```

#### Option 2: Binary Download

```bash
# Download for your architecture
# x86_64 (amd64)
wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer_Linux_x86_64.tar.gz

# ARM64
wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer_Linux_arm64.tar.gz

# ARM (32-bit)
wget https://github.com/jeeftor/audiobook-organizer/releases/latest/download/audiobook-organizer_Linux_armv6.tar.gz

# Extract
tar -xzf audiobook-organizer_Linux_*.tar.gz

# Move to PATH
sudo mv audiobook-organizer /usr/local/bin/

# Verify
audiobook-organizer version
```

### Windows

#### Option 1: Binary Download

```powershell
# Download
# Visit: https://github.com/jeeftor/audiobook-organizer/releases/latest
# Download: audiobook-organizer_Windows_x86_64.zip

# Extract to desired location
Expand-Archive audiobook-organizer_Windows_x86_64.zip -DestinationPath C:\Program Files\AudiobookOrganizer

# Add to PATH (optional, for system-wide access)
# System Properties → Environment Variables → Path → Edit → New
# Add: C:\Program Files\AudiobookOrganizer
```

#### Option 2: Chocolatey (Coming Soon)

```powershell
choco install audiobook-organizer
```

### Verifying CLI/TUI Installation

```bash
# Check version
audiobook-organizer version

# Test CLI
audiobook-organizer --help

# Test TUI
audiobook-organizer tui
```

---

## Docker Installation

Run the organizer in an isolated container. Ideal for NAS devices, servers, and reproducible environments.

### Pulling the Image

```bash
# Latest stable release
docker pull jeffsui/audiobook-organizer:latest

# Specific version
docker pull jeffsui/audiobook-organizer:v1.2.3

# Beta/pre-release
docker pull jeffsui/audiobook-organizer:beta
```

### Basic Usage

```bash
# Organize audiobooks in place
docker run -v /path/to/audiobooks:/books jeffsui/audiobook-organizer --dir=/books

# Separate input and output directories
docker run \
  -v /source:/input:ro \
  -v /dest:/output \
  jeffsui/audiobook-organizer --dir=/input --out=/output

# Dry run (preview changes)
docker run -v /path:/books jeffsui/audiobook-organizer --dir=/books --dry-run
```

### Docker Compose

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  audiobook-organizer:
    image: jeffsui/audiobook-organizer:latest
    volumes:
      - /media/audiobooks:/input:ro  # Read-only source
      - /media/organized:/output     # Write destination
    environment:
      AO_LAYOUT: author-series-title
      AO_VERBOSE: "true"
      AO_REMOVE_EMPTY: "true"
    command: --dir=/input --out=/output
```

**Run:**
```bash
docker-compose up
```

### Environment Variables in Docker

```bash
docker run \
  -v /books:/books \
  -e AO_LAYOUT=author-series-title \
  -e AO_VERBOSE=true \
  -e AO_REMOVE_EMPTY=true \
  jeffsui/audiobook-organizer --dir=/books
```

### Docker on NAS (Synology, QNAP, etc.)

**Synology DSM:**
1. Open Docker package
2. Go to Registry → Search "audiobook-organizer"
3. Download image
4. Create container with volume mappings
5. Set environment variables in Advanced Settings

**QNAP Container Station:**
1. Search for "jeffsui/audiobook-organizer"
2. Create container
3. Map volumes (Shared Folders → Container paths)
4. Set environment variables

---

## Go Install (Development)

Install from source using Go. Requires Go 1.21 or later.

### CLI/TUI Binary

```bash
# Install latest version
go install github.com/jeeftor/audiobook-organizer@latest

# Install specific version
go install github.com/jeeftor/audiobook-organizer@v1.2.3

# Verify installation (ensure $GOPATH/bin is in PATH)
audiobook-organizer version
```

### Building from Source

```bash
# Clone repository
git clone https://github.com/jeeftor/audiobook-organizer.git
cd audiobook-organizer

# Build CLI/TUI
make dev

# Build GUI (requires Wails)
cd audiobook-organizer-gui
wails build

# Output locations:
# - CLI/TUI: bin/audiobook-organizer
# - GUI: audiobook-organizer-gui/build/bin/
```

**Dependencies for building GUI:**
- **macOS:** Xcode Command Line Tools
- **Linux:** `libgtk-3-dev libwebkit2gtk-4.0-dev`
- **Windows:** N/A (included in Wails)
- **All platforms:** Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

---

## Beta and Pre-Release Versions

### Finding Beta Releases

1. Go to [GitHub Releases](https://github.com/jeeftor/audiobook-organizer/releases)
2. Look for releases tagged with `-beta`, `-alpha`, or `-rc`
3. Download platform-specific binaries

**Example:**
```
audiobook-organizer-v1.3.0-beta.1
├── audiobook-organizer-gui_Linux_x86_64.AppImage
├── audiobook-organizer_Darwin_arm64.tar.gz
└── audiobook-organizer_Windows_x86_64.zip
```

### Installing Beta Releases

Follow the same installation steps as stable releases, but download from the beta release page.

**Warning:** Beta releases may have bugs or incomplete features. Use at your own risk.

### Docker Beta Images

```bash
# Pull beta image
docker pull jeffsui/audiobook-organizer:beta

# Or specific beta version
docker pull jeffsui/audiobook-organizer:v1.3.0-beta.1
```

---

## Updating

### Homebrew (macOS/Linux)

```bash
# Update CLI/TUI
brew upgrade audiobook-organizer

# Update GUI (when available)
brew upgrade --cask audiobook-organizer-gui
```

### APT (Debian/Ubuntu)

```bash
# Update package list
sudo apt update

# Update CLI/TUI
sudo apt upgrade audiobook-organizer
```

### YUM/DNF (RedHat/Fedora)

```bash
# Update CLI/TUI
sudo yum upgrade audiobook-organizer
# or
sudo dnf upgrade audiobook-organizer
```

### APK (Alpine)

```bash
# Update CLI/TUI
sudo apk upgrade audiobook-organizer
```

### Manual Binary Updates

1. Download new version from [Releases](https://github.com/jeeftor/audiobook-organizer/releases)
2. Replace existing binary
3. Verify version: `audiobook-organizer version`

### Docker Updates

```bash
# Pull latest image
docker pull jeffsui/audiobook-organizer:latest

# Restart container with new image
docker-compose down
docker-compose up
```

### Checking for Updates

```bash
# Built-in update checker (CLI/TUI)
audiobook-organizer update --check

# Auto-update (if supported by installation method)
audiobook-organizer update
```

---

## Uninstalling

### Homebrew (macOS/Linux)

```bash
# Uninstall CLI/TUI
brew uninstall audiobook-organizer

# Uninstall GUI
brew uninstall --cask audiobook-organizer-gui
```

### APT (Debian/Ubuntu)

```bash
# Uninstall CLI/TUI
sudo apt remove audiobook-organizer

# Remove config files too
sudo apt purge audiobook-organizer

# Uninstall GUI
sudo apt remove audiobook-organizer-gui
```

### YUM/DNF (RedHat/Fedora)

```bash
# Uninstall CLI/TUI
sudo yum remove audiobook-organizer

# Uninstall GUI
sudo yum remove audiobook-organizer-gui
```

### APK (Alpine)

```bash
# Uninstall CLI/TUI
sudo apk del audiobook-organizer
```

### Manual Binary Removal

```bash
# Remove CLI/TUI binary
sudo rm /usr/local/bin/audiobook-organizer

# Remove GUI (macOS)
rm -rf /Applications/Audiobook\ Organizer.app

# Remove config file (optional)
rm ~/.audiobook-organizer.yaml
```

### Docker Cleanup

```bash
# Remove image
docker rmi jeffsui/audiobook-organizer:latest

# Remove stopped containers
docker container prune
```

---

## Platform-Specific Notes

### macOS

**Apple Silicon (M1/M2/M3):**
- Use `_arm64` binaries for native performance
- Rosetta translation works but is slower

**Gatekeeper:**
- First launch of GUI: Right-click → Open
- Or: System Preferences → Security & Privacy → Allow

**Homebrew:**
- Installs to `/opt/homebrew` on Apple Silicon
- Ensure `/opt/homebrew/bin` is in PATH

### Linux

**Terminal Emulator:**
- TUI requires 256-color terminal
- Recommended: iTerm2, Alacritty, GNOME Terminal

**Permissions:**
- Ensure read access to source audiobooks
- Ensure write access to output directory
- Use `sudo` for system-wide installation only

**Headless Servers:**
- CLI mode works without display server
- TUI works over SSH
- GUI requires X11 or Wayland

### Windows

**PowerShell vs CMD:**
- Use PowerShell for better color support
- CMD works but has limited formatting

**Path Spaces:**
- Wrap paths with spaces in quotes: `--dir="C:\My Books"`

**Long Path Support:**
- Enable if organizing deeply nested folders
- Registry: `HKLM\SYSTEM\CurrentControlSet\Control\FileSystem` → `LongPathsEnabled=1`

### Docker

**Volume Permissions:**
- Ensure Docker has access to mount paths
- Use `:ro` for read-only source directories
- Check file ownership after organization

**Networking:**
- Not required for local file operations
- Use `--network none` for isolation

---

## Troubleshooting Installation

### "Command not found" after installation

**Solution:**
```bash
# Check if binary is in PATH
which audiobook-organizer

# Add to PATH if needed (add to ~/.bashrc or ~/.zshrc)
export PATH="$PATH:/usr/local/bin"
```

### Homebrew installation fails

**Solution:**
```bash
# Update Homebrew
brew update

# Retry tap and install
brew tap jeeftor/tap
brew install audiobook-organizer
```

### Linux package dependencies missing

**Solution:**
```bash
# Debian/Ubuntu
sudo apt install -f

# RedHat/Fedora
sudo yum install -y <missing-package>
```

### GUI won't launch (Linux)

**Solution:**
```bash
# Install WebKit2GTK
sudo apt install libwebkit2gtk-4.0-37

# Check for errors
audiobook-organizer-gui --verbose
```

### Docker volume permissions issues

**Solution:**
```bash
# Run container with your user ID
docker run \
  --user $(id -u):$(id -g) \
  -v /path:/books \
  jeffsui/audiobook-organizer --dir=/books
```

### Windows "unrecognized app" warning

**Solution:**
- Click "More info" → "Run anyway"
- Or: Right-click installer → Properties → Unblock

---

## See Also

- [CLI.md](CLI.md) - Command-line usage
- [GUI.md](GUI.md) - Desktop GUI guide
- [TUI.md](TUI.md) - Terminal UI guide
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration file setup
- [Main README](../README.md) - Project overview

---

## Getting Help

- **Installation issues:** [GitHub Issues](https://github.com/jeeftor/audiobook-organizer/issues)
- **Questions:** [GitHub Discussions Q&A](https://github.com/jeeftor/audiobook-organizer/discussions/categories/q-a)

When reporting installation problems, please include:
- Operating system and version
- Installation method used
- Error messages (full text)
- Output of `audiobook-organizer version` (if binary runs)
