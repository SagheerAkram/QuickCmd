# Security Guidelines

## Overview

QuickCMD is designed with security as a top priority. This document outlines the security model, best practices, and guidelines for safe usage.

## Security Model

### Threat Model

**What QuickCMD Protects Against:**
- ✅ Accidental destructive commands
- ✅ Common dangerous patterns (rm -rf /, fork bombs, etc.)
- ✅ Credential leakage in logs
- ✅ Runaway resource consumption
- ✅ Unauthorized command execution (via policies)

**What QuickCMD Does NOT Protect Against:**
- ❌ Intentional malicious use by authorized users
- ❌ Zero-day container escape vulnerabilities
- ❌ Social engineering attacks
- ❌ Physical access to the system
- ❌ Compromised dependencies

### Security Boundaries

1. **Input Validation** - First line of defense
2. **Policy Engine** - Enforces organizational rules
3. **Sandbox Isolation** - Limits blast radius
4. **Audit Logging** - Provides accountability

## Best Practices

### For Individual Users

1. **Always Review Commands**
   - Never blindly execute commands, even from QuickCMD
   - Understand what each command does before running it
   - Use `--dry-run` mode first

2. **Use Sandbox Mode**
   ```bash
   quickcmd "your prompt" --sandbox
   ```
   - Isolates execution in containers
   - Limits resource usage
   - Prevents host filesystem modifications

3. **Never Use `--yes` Flag**
   - Bypasses all confirmation prompts
   - Only use in trusted, automated environments
   - Prefer explicit confirmations

4. **Maintain Backups**
   - QuickCMD provides undo strategies, but they're not foolproof
   - Keep regular backups of important data
   - Test restore procedures

5. **Review Audit Logs**
   ```bash
   quickcmd history
   ```
   - Periodically review execution history
   - Look for unexpected commands
   - Investigate anomalies

### For Organizations

1. **Implement Strict Policies**
   - Define allowlist of permitted commands
   - Maintain comprehensive denylist
   - Require approvals for high-risk operations

2. **Use Custom Policy Files**
   ```yaml
   # /etc/quickcmd/policy.yaml
   allowlist:
     - pattern: "^git (status|log|diff)"
     - pattern: "^ls "
   
   denylist:
     - pattern: "rm -rf"
     - pattern: "shutdown"
   
   approval:
     high_risk: true
     require_multi_party: true
   ```

3. **Enable Audit Logging**
   - Centralize audit logs
   - Set up monitoring and alerting
   - Retain logs for compliance

4. **Restrict Direct Execution**
   - Disable `--yes` flag in production
   - Require sandbox mode by default
   - Implement approval workflows

5. **Regular Security Reviews**
   - Review and update policies quarterly
   - Audit command execution patterns
   - Update denylist with new threats

## Policy Configuration

### Default Denylist

The default policy blocks these dangerous patterns:

- `rm -rf /` - Root directory deletion
- `:(){ :|:& };:` - Fork bomb
- `shutdown` / `reboot` - System control
- `mkfs` - Filesystem formatting
- `dd if=... of=/dev/...` - Disk overwrite
- `chmod 777 /` - Dangerous permissions
- `curl ... | bash` - Piping to shell
- `wget ... | sh` - Piping to shell

### Custom Policies

Create organization-specific policies:

```yaml
# Strict policy for production servers
allowlist:
  - pattern: "^systemctl status"
  - pattern: "^journalctl"
  - pattern: "^docker ps"

denylist:
  - pattern: ".*"  # Block everything not in allowlist

approval:
  high_risk: true
  require_multi_party: true
  allowed_users:
    - admin@example.com
    - ops@example.com
```

## Secrets Management

### Automatic Redaction

QuickCMD automatically redacts:
- Environment variables (PASSWORD, TOKEN, API_KEY, etc.)
- Bearer tokens
- Basic auth credentials
- AWS credentials
- Private keys
- Database connection strings

### Best Practices

1. **Never Log Secrets**
   - Secrets are redacted from audit logs
   - But still avoid passing secrets via command line

2. **Use Vault Integration**
   ```yaml
   secrets:
     vault_integration: true
   ```
   - Retrieve credentials from Vault/1Password
   - Avoid hardcoded secrets

3. **Rotate Credentials**
   - If a secret appears in logs (before redaction), rotate it
   - Treat any logged secret as compromised

## Sandbox Security

### Container Isolation

QuickCMD uses Docker for sandboxing:

**Enabled by Default:**
- Process isolation
- Filesystem isolation
- Resource limits (CPU, memory, time)
- Network isolation (no network access)

**Configuration:**
```yaml
sandbox:
  enabled: true
  network_access: false  # Disable network
  max_cpu: "1.0"         # 1 CPU core
  max_memory: "512m"     # 512 MB RAM
  max_time_seconds: 300  # 5 minutes
```

### Limitations

- **Not a Complete Security Boundary**
  - Container escapes are possible (though rare)
  - Kernel vulnerabilities can be exploited
  - Don't rely solely on sandbox for security

- **Performance Overhead**
  - Container startup: ~1-2 seconds
  - Acceptable for interactive use
  - May be slow for batch operations

## Incident Response

### If a Dangerous Command is Executed

1. **Immediate Actions**
   - Stop the command if still running
   - Assess the damage
   - Restore from backups if needed

2. **Investigation**
   - Review audit logs: `quickcmd history`
   - Identify how the command was generated
   - Check if policy was bypassed

3. **Remediation**
   - Update policy to prevent recurrence
   - Rotate any exposed credentials
   - Document the incident

4. **Prevention**
   - Add pattern to denylist
   - Increase approval requirements
   - Provide team training

### Reporting Security Issues

If you discover a security vulnerability in QuickCMD:

1. **Do NOT** open a public GitHub issue
2. Email security@quickcmd.dev (or your org's security team)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

## Compliance Considerations

### Audit Requirements

QuickCMD provides:
- Append-only audit logs
- Tamper-evident logging
- Complete command history
- User attribution

### Data Retention

Configure retention policies:
```yaml
audit:
  retention_days: 90
  archive_location: "/var/log/quickcmd/archive"
```

### Access Control

- Use OS-level permissions for config files
- Restrict who can modify policies
- Implement multi-party approval for sensitive operations

## Security Checklist

Before deploying QuickCMD:

- [ ] Review and customize default policy
- [ ] Enable sandbox mode by default
- [ ] Disable `--yes` flag in production
- [ ] Configure audit log retention
- [ ] Set up log monitoring and alerting
- [ ] Train users on safe usage
- [ ] Document incident response procedures
- [ ] Test backup and restore procedures
- [ ] Review and update denylist patterns
- [ ] Enable secrets redaction
- [ ] Configure resource limits
- [ ] Implement approval workflows (if needed)

## Regular Maintenance

### Weekly
- Review audit logs for anomalies
- Check for failed policy validations

### Monthly
- Update denylist with new threats
- Review and optimize policies
- Audit user access and permissions

### Quarterly
- Security review of policies
- Update documentation
- Team training refresher

## Additional Resources

- [OWASP Command Injection](https://owasp.org/www-community/attacks/Command_Injection)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks/)

---

**Remember: QuickCMD is a tool to assist developers, not a replacement for security awareness and best practices.**
