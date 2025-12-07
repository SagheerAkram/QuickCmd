# QuickCMD Approval Workflow

## Overview

The approval workflow ensures that potentially dangerous or high-risk commands are reviewed and explicitly approved by authorized personnel before execution.

## How It Works

### 1. Job Submission

When a command is flagged as requiring approval (by policy engine or plugins):

```
User submits command
    ↓
Translation Engine
    ↓
Policy/Plugin checks → RequiresApproval = true
    ↓
Controller creates pending approval
    ↓
Job waits in approval queue
```

### 2. Approval Process

**Approvers** review pending jobs and can:
- **Approve** - Allow job to proceed to execution
- **Reject** - Deny job with reason

**Typed Confirmation Required:**

To approve a job, the approver must type the exact confirmation string:

```
APPROVE <approval_id>
```

Example: `APPROVE 123`

This prevents accidental approvals.

### 3. Post-Approval

After approval:

```
Approver grants approval
    ↓
Controller dispatches job to agent
    ↓
Agent executes in sandbox
    ↓
Results logged to audit database
```

## Role Assignments

### Recommended Roles

| Role | Who | Responsibilities |
|------|-----|------------------|
| **Viewer** | All team members | View execution history |
| **Operator** | Developers, DevOps | Execute safe commands |
| **Approver** | Senior engineers, Team leads | Approve high-risk operations |
| **Admin** | Platform team | Manage users, configure policies |

### Role Hierarchy

```
Admin (all permissions)
  ↓
Approver (can approve + execute + view)
  ↓
Operator (can execute + view)
  ↓
Viewer (can view only)
```

## Approval Criteria

### Commands Requiring Approval

1. **High-Risk Operations**
   - Destructive file operations (`rm -rf`, `delete`)
   - Database modifications
   - Production deployments

2. **Plugin-Flagged Operations**
   - Git: Force push, branch deletion
   - Kubernetes: Apply, delete, scale
   - AWS: Resource creation, cost > threshold

3. **Policy-Based**
   - Commands matching approval patterns in policy
   - Operations affecting sensitive paths
   - Network-accessible operations

### Auto-Approved Operations

- Read-only commands (`ls`, `cat`, `get`)
- Safe file operations (`mkdir`, `touch`)
- Information queries (`status`, `describe`)

## Approval Metadata

Each approval records:

```json
{
  "id": 123,
  "run_id": 456,
  "prompt": "delete old logs",
  "command": "rm -rf /var/log/old/*",
  "risk_level": "high",
  "requested_by": "alice",
  "requested_at": "2025-01-07T10:00:00Z",
  "approved_by": "bob",
  "approved_at": "2025-01-07T10:05:00Z",
  "confirmation": "APPROVE 123",
  "approval_note": "Approved for cleanup task"
}
```

## Audit Trail

All approvals are logged to the audit database:

- **Who** requested the command
- **Who** approved/rejected it
- **When** the approval was granted
- **Why** (approval note or rejection reason)
- **What** was executed and the result

### Audit Retention

**Recommended:**
- Keep approval records for **90 days minimum**
- Archive to long-term storage after 90 days
- Retain indefinitely for compliance-critical operations

**Query Approvals:**

```sql
SELECT * FROM approvals 
WHERE approved_by = 'bob' 
  AND approved_at > datetime('now', '-30 days');
```

## Best Practices

### For Approvers

1. **Review Carefully**
   - Read the full command
   - Check affected paths/resources
   - Verify risk level is appropriate

2. **Ask Questions**
   - If unsure, ask the requester for context
   - Check recent history for similar operations
   - Verify this aligns with planned work

3. **Document Decisions**
   - Add approval notes explaining reasoning
   - For rejections, provide clear reasons
   - Reference tickets/issues when applicable

4. **Time-Sensitive**
   - Review approvals promptly
   - Set up notifications for pending approvals
   - Escalate if blocked

### For Requesters

1. **Provide Context**
   - Use descriptive prompts
   - Mention related work (ticket #, PR #)
   - Explain urgency if time-sensitive

2. **Test First**
   - Run dry-run when possible
   - Test in staging before production
   - Verify command syntax

3. **Be Available**
   - Be ready to answer approver questions
   - Monitor approval status
   - Have rollback plan ready

## Security Considerations

### Approval Bypass Prevention

- Approvals **cannot** override core denylist
- Agent still validates policy before execution
- Approval only signals "proceed to agent"
- Agent performs final safety checks

### Separation of Duties

- Requesters should not approve their own jobs
- Implement in OIDC role mapping or custom logic
- Audit logs track self-approvals

### Approval Expiration

**Recommended:**
- Approvals expire after 1 hour
- Requester must re-submit if expired
- Prevents stale approvals

## Notifications (Future)

Planned notification channels:

- **Slack**: Post to #approvals channel
- **Email**: Send to approver group
- **Webhook**: Custom integrations
- **In-App**: Web UI notifications

## Example Workflows

### Workflow 1: Production Deployment

```
1. Developer: "deploy api to production"
2. System: Creates approval #123 (high risk)
3. Slack: Posts to #approvals channel
4. Team Lead: Reviews in Web UI
5. Team Lead: Types "APPROVE 123" with note "Approved for v2.1 release"
6. System: Dispatches to production agent
7. Agent: Executes deployment in sandbox
8. System: Logs result to audit
9. Slack: Posts completion status
```

### Workflow 2: Emergency Rollback

```
1. On-Call: "rollback api to previous version"
2. System: Creates approval #124 (high risk)
3. On-Call: Pings approver on Slack
4. Approver: Reviews and approves within 2 minutes
5. System: Executes rollback
6. On-Call: Verifies service restored
```

### Workflow 3: Rejected Request

```
1. Junior Dev: "delete all test databases"
2. System: Creates approval #125 (high risk)
3. Senior Dev: Reviews request
4. Senior Dev: Rejects with reason "Too broad - specify database names"
5. Junior Dev: Resubmits "delete test-db-feature-x"
6. Senior Dev: Approves specific request
```

## Monitoring

### Metrics to Track

- **Approval Rate**: % of requests approved vs rejected
- **Time to Approval**: Average time from request to decision
- **Approval Volume**: Requests per day/week
- **Self-Approvals**: Track for audit purposes

### Alerts

Set up alerts for:
- Pending approvals > 1 hour old
- High volume of rejections (may indicate policy issues)
- Self-approvals (if not allowed)
- Approvals outside business hours

## Compliance

### SOC 2 / ISO 27001

Approval workflow supports compliance by:
- ✅ Documented approval process
- ✅ Separation of duties
- ✅ Complete audit trail
- ✅ Role-based access control
- ✅ Immutable audit logs

### GDPR / Data Privacy

- Approval records may contain personal data
- Apply data retention policies
- Support data deletion requests
- Redact sensitive information in logs

---

**For Web UI usage, see [WEB_UI.md](./WEB_UI.md)**
