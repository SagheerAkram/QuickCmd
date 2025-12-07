# Quick Start Tutorial

Learn QuickCMD in 10 minutes!

---

## 🎯 Your First Command

```bash
quickcmd "find large files"
```

**Output:**
```
✨ Candidates for: find large files

1. ✓ safe
   find . -type f -size +100M
   Confidence: ████████████████████ 95%
   
   Finds all files in current directory larger than 100MB
   
   ✓ Exact match to template pattern
   ✓ Uses current directory context
   ✓ Safe operation (read-only)

ℹ️  Dry-run mode: commands will not be executed
Use --sandbox to run in isolated container, or --yes to execute directly
```

---

## 📚 Basic Concepts

### 1. Dry-Run Mode (Default)

By default, QuickCMD shows commands without executing:

```bash
quickcmd "delete old logs"
# Shows command, doesn't execute
```

### 2. Sandbox Mode (Safe Execution)

Execute in isolated Docker container:

```bash
quickcmd "delete old logs" --sandbox
# Executes safely in container
```

### 3. Direct Execution (⚠️ Use Carefully)

Execute directly on your system:

```bash
quickcmd "list files" --yes
# Executes immediately
```

---

## 💡 Common Use Cases

### File Operations

```bash
# Find files
quickcmd "find all log files"
quickcmd "find files larger than 1GB"
quickcmd "find empty directories"

# Search content
quickcmd "search for error in logs"
quickcmd "find files containing TODO"

# Disk usage
quickcmd "show disk usage"
quickcmd "show largest directories"
```

### Docker

```bash
# Containers
quickcmd "list running containers"
quickcmd "stop all containers"
quickcmd "show container logs"

# Images
quickcmd "list docker images"
quickcmd "remove unused images"
```

### Git

```bash
# Commits
quickcmd "undo last commit"
quickcmd "show recent commits"

# Branches
quickcmd "create new branch"
quickcmd "delete merged branches"
```

### System

```bash
# Processes
quickcmd "show memory usage"
quickcmd "kill process on port 8080"

# Network
quickcmd "show listening ports"
quickcmd "test connection to google"
```

---

## 🎓 Learning Mode

Get explanations for any command:

```bash
quickcmd explain "find . -name '*.log'"
```

**Output:**
```
📚 Command Breakdown:

find .                    Search from current directory
  -name '*.log'           Match files ending with .log

💡 What this does:
Finds all .log files in current directory and subdirectories

⚡ Optimization tip:
Use -maxdepth to limit search depth

🎓 Related commands:
- locate *.log
- grep -r pattern
```

---

## 📖 View History

```bash
# View all history
quickcmd history

# Search history
quickcmd history search "docker"

# Last 7 days
quickcmd history --last 7d
```

---

## ⚡ Create Aliases

Save frequently used commands:

```bash
# Create alias
quickcmd alias create deploy "kubectl rollout restart deployment"

# Use alias
quickcmd deploy api

# List aliases
quickcmd alias list
```

---

## 🔒 Security Features

### Policy Enforcement

QuickCMD blocks dangerous commands:

```bash
quickcmd "rm -rf /"
# ❌ Blocked by policy: Dangerous root deletion
```

### Approval Workflows

High-risk operations require approval:

```bash
quickcmd "scale production to 100" --request-approval
# 🔒 Approval requested
# Waiting for admin approval...
```

### Sandbox Execution

Always use sandbox for destructive operations:

```bash
quickcmd "delete all logs" --sandbox
# 🐳 Executing in isolated container
# 📸 Backup created
# ✓ Safe to execute
```

---

## 🚀 Advanced Features

### Scheduling

```bash
# Schedule daily backup
quickcmd schedule create "backup database" --cron "0 2 * * *"

# List scheduled jobs
quickcmd schedule list
```

### Cost Estimation

```bash
quickcmd "launch 10 ec2 instances" --estimate-cost
# 💰 Estimated: $0.42/hour ($306/month)
# 💡 Save 70% with spot instances
```

### Performance Optimization

```bash
quickcmd "find . | grep error" --optimize
# ⚡ Optimized: grep -r error .
# Speedup: 3x faster
```

---

## 🎯 Best Practices

### ✅ DO

- Use `--sandbox` for destructive operations
- Review commands in dry-run mode first
- Create aliases for frequent commands
- Use learning mode to understand commands
- Check history before repeating commands

### ❌ DON'T

- Use `--yes` for untrusted commands
- Disable security policies
- Skip approval workflows
- Ignore warnings
- Execute without understanding

---

## 📊 Example Workflow

### Daily DevOps Tasks

```bash
# Morning: Check system
quickcmd "show disk usage"
quickcmd "list running containers"
quickcmd "check kubernetes pods"

# Work: Deploy changes
quickcmd alias deploy "kubectl rollout restart"
quickcmd deploy api --sandbox

# Evening: Cleanup
quickcmd "remove old docker images"
quickcmd "compress old logs"
```

---

## 🆘 Common Issues

### Command Not Executing

```bash
# Use --yes or --sandbox
quickcmd "command" --sandbox
```

### Docker Not Available

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh

# Or skip sandbox
quickcmd "command" --yes
```

### Permission Denied

```bash
# Add to docker group
sudo usermod -aG docker $USER

# Or use sudo
sudo quickcmd "command"
```

---

## 🚦 Next Steps

1. **[Configuration](Configuration)** - Customize QuickCMD
2. **[Security Setup](Security)** - Configure policies
3. **[Command Examples](Examples)** - More use cases
4. **[Advanced Features](Advanced)** - Deep dive

---

## 💬 Get Help

- **Discord**: [Ask questions](https://discord.gg/Bg3gDAqDwz)
- **Docs**: [Full documentation](https://github.com/SagheerAkram/QuickCmd/wiki)
- **Issues**: [Report bugs](https://github.com/SagheerAkram/QuickCmd/issues)

---

**Ready to become a QuickCMD power user? 🚀**
