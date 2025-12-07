# Confidence Scoring Visualization

## Overview

The confidence scoring system provides transparent insight into why QuickCMD chose a specific command translation. It breaks down the overall confidence into multiple components and explains the reasoning behind each score.

## How It Works

### Confidence Components

Each command translation is scored across four dimensions:

1. **Pattern Match** (0-100%)
   - How well the prompt matches known command templates
   - 95-100%: Exact template match
   - 80-94%: Strong pattern similarity
   - Below 80%: Fuzzy match

2. **Context Awareness** (0-100%)
   - Whether the command uses current directory, environment variables
   - Detects if command is context-specific or generic

3. **Risk Assessment** (0-100%)
   - Safety evaluation of the command
   - 100%: Safe (read-only operations)
   - 75%: Medium risk (modifies state)
   - 50%: High risk (destructive)

4. **Plugin Analysis** (0-100%)
   - Plugin-specific safety checks
   - Domain expertise applied to command

### Overall Score

The overall confidence is the average of all component scores:

```
Overall = (Pattern + Context + Risk + Plugin) / 4
```

## Visual Display

### CLI Output

```
✨ Confidence: 93% ████████████████████

Components:
  Pattern Match:       95% ███████████████████░
  Context:             90% ██████████████████░░
  Risk:               100% ████████████████████
  Plugin:              87% █████████████████░░░

Why this command?
  ✓ Exact match to template pattern
  ✓ Uses current directory context
  ✓ Safe operation (read-only)

Tips:
  💡 Could be slow in large directories
```

### Web UI

The Web UI displays confidence as:
- **Progress bars** for each component
- **Expandable sections** for detailed explanations
- **Comparison view** when multiple candidates exist
- **Historical trends** showing confidence over time

## API Endpoints

### Get Confidence Breakdown

```http
GET /api/v1/run/{id}/confidence
```

**Response:**
```json
{
  "overall": 93,
  "components": {
    "pattern": 95,
    "context": 90,
    "risk": 100,
    "plugin": 87
  },
  "reasons": [
    "Exact match to template pattern",
    "Uses current directory context",
    "Safe operation (read-only)"
  ],
  "warnings": [],
  "tips": [
    "Could be slow in large directories"
  ]
}
```

### Explain Command

```http
POST /api/v1/explain
Content-Type: application/json

{
  "command": "find . -type f -size +100M",
  "prompt": "find large files"
}
```

**Response:**
```json
{
  "explanation": "Finds all files in current directory larger than 100MB",
  "breakdown": [
    "Search current directory (.)",
    "Filter for files only (-type f)",
    "Match size greater than 100MB (-size +100M)"
  ],
  "risk_level": "safe",
  "affected_paths": ["."],
  "confidence": 95
}
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "github.com/SagheerAkram/QuickCmd/core/translator"
)

func main() {
    // Create candidate
    candidate := &translator.Candidate{
        Command:    "find . -type f -size +100M",
        Confidence: 95,
        RiskLevel:  translator.RiskSafe,
    }
    
    // Calculate breakdown
    breakdown := candidate.CalculateConfidenceBreakdown("find large files")
    
    // Display visualization
    fmt.Print(breakdown.Visualize())
}
```

### Custom Components

```go
breakdown := translator.NewConfidenceBreakdown()

// Add custom components
breakdown.AddComponent(translator.ComponentPattern, 98, "Exact template match")
breakdown.AddComponent(translator.ComponentContext, 92, "Current directory detected")
breakdown.AddComponent(translator.ComponentRisk, 100, "Safe read-only operation")

// Add warnings and tips
breakdown.AddWarning("Large directory - may be slow")
breakdown.AddTip("Use -maxdepth to limit search depth")

// Calculate and display
breakdown.Calculate()
fmt.Print(breakdown.Visualize())
```

## Interpreting Scores

### High Confidence (90-100%)
- ✅ Safe to execute
- ✅ Well-understood command
- ✅ Minimal risk

### Medium Confidence (70-89%)
- ⚠️ Review before executing
- ⚠️ May need adjustments
- ⚠️ Verify affected resources

### Low Confidence (<70%)
- 🔴 Verify carefully
- 🔴 High uncertainty
- 🔴 Consider alternative approaches

## Best Practices

### For Users

1. **Always check the breakdown** - Don't just look at the overall score
2. **Read warnings** - They highlight potential issues
3. **Follow tips** - They suggest optimizations
4. **Use sandbox mode** for commands with confidence <90%

### For Developers

1. **Add detailed reasons** - Explain why each score was assigned
2. **Provide actionable tips** - Help users improve their prompts
3. **Be conservative with risk scores** - Better safe than sorry
4. **Update patterns regularly** - Improve confidence over time

## Future Enhancements

- **ML-based scoring** - Learn from user feedback
- **Historical comparison** - "This is 15% more confident than last time"
- **Personalized confidence** - Adapt to user expertise level
- **Confidence trends** - Track improvements over time

---

**See also:**
- [Command Explanation](./COMMAND_EXPLANATION.md)
- [Risk Assessment](./RISK_ASSESSMENT.md)
- [Pattern Matching](./PATTERN_MATCHING.md)
