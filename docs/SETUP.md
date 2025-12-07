# QuickCMD Setup Guide

This guide will help you set up QuickCMD on your system.

## Prerequisites

### 1. Install Go (Required)

QuickCMD is written in Go and requires Go 1.21 or later.

#### Windows

**Option A: Using Chocolatey**
```powershell
choco install golang
```

**Option B: Manual Installation**
1. Download Go from https://go.dev/dl/
2. Download the Windows installer (e.g., `go1.21.5.windows-amd64.msi`)
3. Run the installer
4. Verify installation:
   ```powershell
   go version
   ```

**Option C: Using winget**
```powershell
winget install GoLang.Go
```

After installation, restart your terminal and verify:
```powershell
go version
# Should output: go version go1.21.x windows/amd64
```

#### macOS

**Option A: Using Homebrew**
```bash
brew install go
```

**Option B: Manual Installation**
1. Download from https://go.dev/dl/
2. Download the macOS package (e.g., `go1.21.5.darwin-amd64.pkg`)
3. Run the installer
4. Verify: `go version`

#### Linux

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install golang-go
```

**Fedora/RHEL:**
```bash
sudo dnf install golang
```

**Arch Linux:**
```bash
sudo pacman -S go
```

**Manual Installation (any Linux):**
```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

### 2. Install Docker (Optional, for Sandbox Mode)

Docker is required only if you want to use sandbox execution mode.

#### Windows
1. Download Docker Desktop from https://www.docker.com/products/docker-desktop
2. Install and start Docker Desktop
3. Verify: `docker --version`

#### macOS
```bash
brew install --cask docker
```

#### Linux
```bash
# Ubuntu/Debian
sudo apt install docker.io
sudo systemctl start docker
sudo usermod -aG docker $USER

# Fedora/RHEL
sudo dnf install docker
sudo systemctl start docker
sudo usermod -aG docker $USER
```

## Building QuickCMD

### 1. Clone the Repository

```bash
cd c:\Users\Sagheer\Desktop\project\QuickCmd
```

### 2. Download Dependencies

```bash
go mod download
```

This will download all required Go packages.

### 3. Build the Binary

**Windows:**
```powershell
# Using Make (if you have Make installed)
make build

# Or directly with Go
go build -o bin/quickcmd.exe ./cmd/quickcmd
```

**macOS/Linux:**
```bash
# Using Make
make build

# Or directly with Go
go build -o bin/quickcmd ./cmd/quickcmd
```

### 4. Verify the Build

```bash
./bin/quickcmd version
# or on Windows:
.\bin\quickcmd.exe version
```

## Installation

### Option 1: Install to GOPATH (Recommended)

```bash
make install
```

This installs `quickcmd` to `$GOPATH/bin` (usually `~/go/bin`).

Make sure `$GOPATH/bin` is in your PATH:

**Windows (PowerShell):**
```powershell
$env:Path += ";$env:USERPROFILE\go\bin"
# To make permanent, add to system environment variables
```

**macOS/Linux:**
```bash
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
```

### Option 2: Copy to System Path

**Windows:**
```powershell
copy bin\quickcmd.exe C:\Windows\System32\
```

**macOS/Linux:**
```bash
sudo cp bin/quickcmd /usr/local/bin/
```

### Option 3: Use from Build Directory

You can run QuickCMD directly from the build directory:
```bash
./bin/quickcmd "your prompt here"
```

## Configuration

### 1. Create Config Directory

```bash
# Windows
mkdir $env:USERPROFILE\.quickcmd

# macOS/Linux
mkdir -p ~/.quickcmd
```

### 2. Copy Default Policy

```bash
# Windows
copy examples\policies\default.yaml $env:USERPROFILE\.quickcmd\policy.yaml

# macOS/Linux
cp examples/policies/default.yaml ~/.quickcmd/policy.yaml
```

### 3. (Optional) Create Custom Config

Create `~/.quickcmd/config.yaml`:

```yaml
# QuickCMD Configuration

# Policy file location
policy_file: ~/.quickcmd/policy.yaml

# Audit log location
audit_db: ~/.quickcmd/audit.db

# Default execution mode: dry-run, sandbox, or direct
default_mode: dry-run

# Enable verbose output
verbose: false

# Sandbox settings
sandbox:
  enabled: true
  default_image: alpine:latest
```

## Running Tests

### Unit Tests

```bash
make test
```

### With Coverage

```bash
make coverage
```

This generates `coverage.html` that you can open in a browser.

### Integration Tests (requires Docker)

```bash
make test-integration
```

## Usage Examples

### Basic Usage

```bash
# Get command suggestions (dry-run mode)
quickcmd "find files larger than 100MB"

# Execute in sandbox
quickcmd "delete all .DS_Store files" --sandbox

# View history
quickcmd history
```

### Testing with Example Prompts

```bash
# Try various prompts from examples/prompts.txt
quickcmd "show git changes"
quickcmd "archive logs older than 7 days"
quickcmd "cleanup docker containers"
```

## Troubleshooting

### Go Not Found

**Error:** `go: command not found`

**Solution:**
- Ensure Go is installed: Download from https://go.dev/dl/
- Restart your terminal after installation
- Verify `go version` works

### Docker Not Running

**Error:** `Cannot connect to the Docker daemon`

**Solution:**
- Start Docker Desktop (Windows/macOS)
- Or start Docker service: `sudo systemctl start docker` (Linux)
- Verify: `docker ps`

### Permission Denied

**Error:** `permission denied` when running quickcmd

**Solution:**
```bash
# Make binary executable (macOS/Linux)
chmod +x bin/quickcmd

# Or run with sudo
sudo ./bin/quickcmd
```

### Module Download Fails

**Error:** `go: module ... not found`

**Solution:**
```bash
# Clear module cache and retry
go clean -modcache
go mod download
go mod verify
```

### Build Fails on Windows

**Error:** `gcc: command not found` (for SQLite)

**Solution:**
- Install TDM-GCC: https://jmeubank.github.io/tdm-gcc/
- Or use pre-built binaries (coming soon)

## Development Setup

### Install Development Tools

```bash
# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Security scanner
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Auto-reload for development
go install github.com/cosmtrek/air@latest
```

### Run in Development Mode

```bash
make dev
```

This uses `air` for auto-reload on file changes.

### Run Linters

```bash
make lint
```

### Run Security Scan

```bash
make test-security
```

## Next Steps

1. **Try Example Prompts**
   - See `examples/prompts.txt` for ideas
   - Start with safe operations

2. **Customize Your Policy**
   - Edit `~/.quickcmd/policy.yaml`
   - Add organization-specific rules

3. **Read Documentation**
   - [Architecture](docs/ARCHITECTURE.md)
   - [Security Guidelines](docs/SECURITY.md)

4. **Join the Community**
   - Report issues on GitHub
   - Contribute improvements
   - Share your use cases

## Uninstallation

```bash
# Remove binary
rm $(which quickcmd)

# Remove config
rm -rf ~/.quickcmd

# Remove Go modules cache (optional)
go clean -modcache
```

## Getting Help

- **Documentation:** See `docs/` directory
- **Issues:** https://github.com/yourusername/quickcmd/issues
- **Discussions:** https://github.com/yourusername/quickcmd/discussions

---

**Happy commanding! 🚀**
