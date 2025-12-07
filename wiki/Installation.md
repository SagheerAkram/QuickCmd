# Installation Guide

Get QuickCMD up and running in minutes!

---

## 📋 Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Docker** (optional) - For sandbox execution
- **Git** - For cloning repository

---

## 🚀 Quick Install

### Linux/macOS (One-Line Install)

```bash
curl -sSL https://raw.githubusercontent.com/SagheerAkram/QuickCmd/main/install.sh | bash
```

This will:
1. Download the latest release
2. Install to `/usr/local/bin`
3. Set up configuration directory
4. Verify installation

---

## 🔧 Installation Methods

### Method 1: Build from Source (Recommended)

```bash
# 1. Clone repository
git clone https://github.com/SagheerAkram/QuickCmd.git
cd QuickCmd

# 2. Build CLI
make build

# 3. Build agent (optional)
make build-agent

# 4. Install globally
sudo make install

# 5. Verify
quickcmd --version
```

### Method 2: Pre-built Binaries

```bash
# Download CLI
curl -L https://github.com/SagheerAkram/QuickCmd/releases/latest/download/quickcmd-linux-amd64 -o quickcmd
chmod +x quickcmd
sudo mv quickcmd /usr/local/bin/

# Download agent (optional)
curl -L https://github.com/SagheerAkram/QuickCmd/releases/latest/download/quickcmd-agent-linux-amd64 -o quickcmd-agent
chmod +x quickcmd-agent
sudo mv quickcmd-agent /usr/local/bin/
```

### Method 3: Docker

```bash
# Pull image
docker pull sagheerakram/quickcmd:latest

# Create alias
echo "alias quickcmd='docker run -it --rm sagheerakram/quickcmd'" >> ~/.bashrc
source ~/.bashrc

# Use it
quickcmd "find large files"
```

### Method 4: Go Install

```bash
go install github.com/SagheerAkram/QuickCmd/cmd/quickcmd@latest
go install github.com/SagheerAkram/QuickCmd/cmd/quickcmd-agent@latest
```

---

## ✅ Verify Installation

```bash
# Check version
quickcmd --version
# QuickCMD v2.0.0

# Test basic command
quickcmd "list files"
# Should show: ls -la

# Check configuration
quickcmd config show
```

---

## 🎨 Web UI Installation (Optional)

```bash
# Navigate to web frontend
cd web/frontend

# Install dependencies
npm install

# Build production bundle
npm run build

# Start web server
cd ../..
./bin/quickcmd-web --port 3000
```

Access at: `http://localhost:3000`

---

## 🐳 Docker Setup (Optional)

For sandbox execution:

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh

# Add user to docker group
sudo usermod -aG docker $USER

# Logout and login again

# Verify Docker
docker --version
```

---

## 📁 Directory Structure

After installation, QuickCMD creates:

```
~/.quickcmd/
├── config.yaml          # User configuration
├── policy.yaml          # Security policies
├── audit.db             # Command history
└── plugins/             # Custom plugins
```

---

## ⚙️ Initial Configuration

Create default configuration:

```bash
# Generate config
quickcmd init

# Edit configuration
nano ~/.quickcmd/config.yaml
```

**Basic config.yaml:**
```yaml
default_mode: sandbox
auto_approve_safe: true
show_confidence: true
enable_learning_mode: true
```

---

## 🔒 Security Setup

Create security policy:

```bash
# Generate default policy
quickcmd policy init

# Edit policy
nano ~/.quickcmd/policy.yaml
```

**Basic policy.yaml:**
```yaml
denylist:
  - pattern: "rm -rf /"
    reason: "Prevents root deletion"

approval_required:
  - pattern: "kubectl.*production"
    reason: "Production needs approval"
```

---

## 🚦 Next Steps

1. **[Quick Start Tutorial](Quick-Start)** - Learn basic usage
2. **[Configuration Guide](Configuration)** - Customize QuickCMD
3. **[Security Setup](Security)** - Configure policies
4. **[Examples](Examples)** - See real-world use cases

---

## 🆘 Troubleshooting

### Command Not Found

```bash
# Check PATH
echo $PATH

# Add to PATH
export PATH=$PATH:/usr/local/bin

# Make permanent
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
```

### Permission Denied

```bash
# Fix permissions
sudo chown -R $USER:$USER ~/.quickcmd

# Or run with sudo
sudo quickcmd "command"
```

### Docker Issues

```bash
# Check Docker
docker ps

# Restart Docker
sudo systemctl restart docker

# Check user in docker group
groups $USER
```

---

## 📞 Need Help?

- **Discord**: [Join community](https://discord.gg/Bg3gDAqDwz)
- **Issues**: [Report problem](https://github.com/SagheerAkram/QuickCmd/issues)
- **Docs**: [Full documentation](https://github.com/SagheerAkram/QuickCmd/wiki)
