# Command Examples

Real-world examples for every use case!

---

## 📁 File Operations

### Finding Files

```bash
# By name
quickcmd "find all log files"
# → find . -name "*.log"

quickcmd "find python files in src directory"
# → find src -name "*.py"

# By size
quickcmd "find files larger than 1GB"
# → find . -type f -size +1G

quickcmd "find files smaller than 1MB"
# → find . -type f -size -1M

# By date
quickcmd "find files modified today"
# → find . -type f -mtime 0

quickcmd "find files modified in last 7 days"
# → find . -type f -mtime -7

# By type
quickcmd "find empty directories"
# → find . -type d -empty

quickcmd "find broken symlinks"
# → find . -type l ! -exec test -e {} \; -print
```

### Searching Content

```bash
# Basic search
quickcmd "search for error in all files"
# → grep -r "error" .

quickcmd "search for TODO in python files"
# → grep -r "TODO" --include="*.py" .

# Case insensitive
quickcmd "search for password ignoring case"
# → grep -ri "password" .

# With context
quickcmd "search for error with 3 lines context"
# → grep -r "error" -C 3 .
```

### Disk Usage

```bash
# Show usage
quickcmd "show disk usage"
# → df -h

quickcmd "show disk usage sorted by size"
# → du -sh * | sort -h

# Top directories
quickcmd "show top 10 largest directories"
# → du -sh * | sort -rh | head -10

quickcmd "show size of current directory"
# → du -sh .
```

---

## 🐳 Docker

### Container Management

```bash
# List containers
quickcmd "list all running containers"
# → docker ps

quickcmd "list all containers including stopped"
# → docker ps -a

# Start/Stop
quickcmd "stop all containers"
# → docker stop $(docker ps -q)

quickcmd "restart nginx container"
# → docker restart nginx

# Logs
quickcmd "show logs for api container"
# → docker logs api

quickcmd "follow logs for nginx"
# → docker logs -f nginx

# Execute commands
quickcmd "run bash in nginx container"
# → docker exec -it nginx bash

# Cleanup
quickcmd "remove stopped containers"
# → docker container prune -f

quickcmd "remove all containers"
# → docker rm -f $(docker ps -aq)
```

### Image Management

```bash
# List images
quickcmd "list all docker images"
# → docker images

quickcmd "list images sorted by size"
# → docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | sort -k3 -h

# Remove images
quickcmd "remove unused images"
# → docker image prune -a

quickcmd "remove image nginx"
# → docker rmi nginx

# Build
quickcmd "build image from current directory"
# → docker build -t myapp .
```

---

## ☸️ Kubernetes

### Pod Management

```bash
# List pods
quickcmd "get all pods"
# → kubectl get pods

quickcmd "get pods in production namespace"
# → kubectl get pods -n production

quickcmd "get pods with labels"
# → kubectl get pods -l app=api

# Describe
quickcmd "describe pod api-5f7b8"
# → kubectl describe pod api-5f7b8

# Logs
quickcmd "get logs for api pod"
# → kubectl logs api-5f7b8

quickcmd "follow logs for api deployment"
# → kubectl logs -f deployment/api

# Execute
quickcmd "run bash in api pod"
# → kubectl exec -it api-5f7b8 -- bash
```

### Deployment Management

```bash
# Scale
quickcmd "scale api deployment to 5 replicas"
# → kubectl scale deployment api --replicas=5

# Restart
quickcmd "restart api deployment"
# → kubectl rollout restart deployment api

# Status
quickcmd "check rollout status for api"
# → kubectl rollout status deployment api

# History
quickcmd "show rollout history"
# → kubectl rollout history deployment api
```

### Service Management

```bash
# List services
quickcmd "list all services"
# → kubectl get services

# Expose
quickcmd "expose deployment api on port 8080"
# → kubectl expose deployment api --port=8080

# Port forward
quickcmd "forward port 8080 to api service"
# → kubectl port-forward service/api 8080:8080
```

---

## 🔧 Git

### Commits

```bash
# Undo
quickcmd "undo last commit but keep changes"
# → git reset --soft HEAD~1

quickcmd "undo last 3 commits"
# → git reset --soft HEAD~3

# History
quickcmd "show commits from last week"
# → git log --since="1 week ago"

quickcmd "show commits by John"
# → git log --author="John"

quickcmd "show one line commit history"
# → git log --oneline

# Amend
quickcmd "amend last commit"
# → git commit --amend
```

### Branches

```bash
# Create
quickcmd "create new branch from main"
# → git checkout -b new-branch main

# Delete
quickcmd "delete merged branches"
# → git branch --merged | grep -v "\\*" | xargs -n 1 git branch -d

quickcmd "delete branch feature-x"
# → git branch -D feature-x

# List
quickcmd "list all branches"
# → git branch -a

quickcmd "list remote branches"
# → git branch -r
```

### Stash

```bash
# Save
quickcmd "stash current changes"
# → git stash

quickcmd "stash with message"
# → git stash save "work in progress"

# Apply
quickcmd "apply last stash"
# → git stash pop

quickcmd "list all stashes"
# → git stash list
```

---

## ☁️ AWS

### S3

```bash
# List
quickcmd "list all s3 buckets"
# → aws s3 ls

quickcmd "list files in bucket"
# → aws s3 ls s3://my-bucket

# Copy
quickcmd "copy file to s3"
# → aws s3 cp file.txt s3://my-bucket/

quickcmd "sync folder to s3"
# → aws s3 sync ./folder s3://my-bucket/folder

# Delete
quickcmd "delete file from s3"
# → aws s3 rm s3://my-bucket/file.txt
```

### EC2

```bash
# List instances
quickcmd "list ec2 instances"
# → aws ec2 describe-instances

quickcmd "list running instances"
# → aws ec2 describe-instances --filters "Name=instance-state-name,Values=running"

# Start/Stop
quickcmd "stop instance i-1234567890"
# → aws ec2 stop-instances --instance-ids i-1234567890

quickcmd "start instance i-1234567890"
# → aws ec2 start-instances --instance-ids i-1234567890
```

---

## 💻 System Monitoring

### Processes

```bash
# List processes
quickcmd "show top 10 CPU consuming processes"
# → ps aux --sort=-%cpu | head -11

quickcmd "show top 10 memory consuming processes"
# → ps aux --sort=-%mem | head -11

# Kill
quickcmd "kill process on port 8080"
# → kill $(lsof -t -i:8080)

quickcmd "kill all node processes"
# → pkill node
```

### Network

```bash
# Ports
quickcmd "show listening ports"
# → netstat -tuln | grep LISTEN

quickcmd "show process using port 8080"
# → lsof -i :8080

# Connections
quickcmd "show active connections"
# → netstat -an | grep ESTABLISHED

# Test
quickcmd "test connection to google.com"
# → ping -c 4 google.com

quickcmd "check if port 80 is open"
# → nc -zv localhost 80
```

### Logs

```bash
# System logs
quickcmd "tail last 100 lines of syslog"
# → tail -n 100 /var/log/syslog

quickcmd "follow syslog"
# → tail -f /var/log/syslog

# Application logs
quickcmd "follow nginx error log"
# → tail -f /var/log/nginx/error.log

quickcmd "search for error in nginx logs"
# → grep "error" /var/log/nginx/error.log
```

---

## 🔐 Security & Permissions

```bash
# File permissions
quickcmd "make file executable"
# → chmod +x file.sh

quickcmd "change owner to user"
# → chown user:user file.txt

# SSH
quickcmd "generate ssh key"
# → ssh-keygen -t rsa -b 4096

quickcmd "copy ssh key to server"
# → ssh-copy-id user@server.com

# Firewall
quickcmd "allow port 80 in firewall"
# → sudo ufw allow 80

quickcmd "show firewall status"
# → sudo ufw status
```

---

## 📊 Data Processing

```bash
# CSV/Text
quickcmd "count lines in file"
# → wc -l file.txt

quickcmd "sort file and remove duplicates"
# → sort file.txt | uniq

quickcmd "get first 10 lines"
# → head -10 file.txt

# JSON
quickcmd "pretty print json file"
# → jq '.' file.json

quickcmd "extract field from json"
# → jq '.field' file.json

# Archives
quickcmd "compress folder to tar.gz"
# → tar -czf archive.tar.gz folder/

quickcmd "extract tar.gz file"
# → tar -xzf archive.tar.gz
```

---

**More examples in [Full Documentation](https://github.com/SagheerAkram/QuickCmd/wiki)**
