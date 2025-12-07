# Contributing to QuickCMD

Thank you for your interest in contributing to QuickCMD! 🎉

## How to Contribute

### 1. Fork the Repository

Click the "Fork" button at the top right of the [QuickCMD repository](https://github.com/SagheerAkram/QuickCmd).

### 2. Clone Your Fork

```bash
git clone https://github.com/YOUR-USERNAME/QuickCmd.git
cd QuickCmd
```

### 3. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/add-docker-plugin` - for new features
- `fix/sandbox-memory-leak` - for bug fixes
- `docs/improve-readme` - for documentation
- `test/add-policy-tests` - for tests

### 4. Make Your Changes

- Write clean, readable code
- Follow existing code style
- Add tests for new features
- Update documentation if needed

### 5. Test Your Changes

```bash
# Run all tests
make test

# Run specific tests
go test ./core/translator/...

# Check code quality
make lint
make fmt
```

### 6. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "Add Docker plugin with container management"
```

**Good commit messages:**
- ✅ "Add Docker plugin with container management"
- ✅ "Fix memory leak in sandbox cleanup"
- ✅ "Update README with installation instructions"

**Bad commit messages:**
- ❌ "Update"
- ❌ "Fix bug"
- ❌ "Changes"

### 7. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 8. Create a Pull Request

1. Go to the [QuickCMD repository](https://github.com/SagheerAkram/QuickCmd)
2. Click "New Pull Request"
3. Select your fork and branch
4. Fill in the PR template with:
   - What changes you made
   - Why you made them
   - How to test them

## What to Contribute

### 🔌 New Plugins

We'd love more plugins! Popular requests:
- **Docker** - Container management
- **Terraform** - Infrastructure as code
- **Ansible** - Configuration management
- **npm/yarn** - Package management
- **systemd** - Service management

**Plugin Template:**
```go
package myplugin

import "github.com/SagheerAkram/QuickCmd/core/plugins"

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "myplugin"
}

func (p *MyPlugin) Translate(ctx plugins.Context, prompt string) ([]*plugins.Candidate, error) {
    // Your logic here
    return nil, nil
}

// Implement other required methods...
```

### 🐛 Bug Fixes

Found a bug? Please:
1. Check if it's already reported in [Issues](https://github.com/SagheerAkram/QuickCmd/issues)
2. If not, create a new issue with:
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Your environment (OS, Go version, Docker version)
3. Submit a PR with the fix

### 📖 Documentation

Help improve our docs:
- Fix typos
- Add examples
- Clarify confusing sections
- Translate to other languages

### 🧪 Tests

More tests = more confidence:
- Add unit tests
- Add integration tests
- Improve test coverage
- Add edge case tests

### 🎨 UI Improvements

Make the Web UI better:
- Improve design
- Add new features
- Fix responsive issues
- Improve accessibility

## Code Style

### Go Code

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `golint` for linting
- Write descriptive variable names
- Add comments for exported functions

**Example:**
```go
// TranslatePrompt converts a natural language prompt into shell commands.
// It returns a list of candidate commands ranked by confidence.
func TranslatePrompt(prompt string) ([]*Candidate, error) {
    // Implementation...
}
```

### JavaScript/TypeScript

- Use 2 spaces for indentation
- Use semicolons
- Use `const` and `let`, not `var`
- Write descriptive variable names

### Documentation

- Use clear, simple language
- Include code examples
- Add screenshots for UI changes
- Keep line length under 100 characters

## Testing Guidelines

### Unit Tests

Test individual functions:

```go
func TestTranslatePrompt(t *testing.T) {
    translator := NewTranslator()
    candidates, err := translator.Translate("find large files")
    
    assert.Nil(t, err)
    assert.NotEmpty(t, candidates)
    assert.Contains(t, candidates[0].Command, "find")
}
```

### Integration Tests

Test complete workflows:

```go
func TestSandboxExecution(t *testing.T) {
    runner := NewDockerRunner()
    result, err := runner.Execute("ls -la", &Config{})
    
    assert.Nil(t, err)
    assert.Equal(t, 0, result.ExitCode)
}
```

## Pull Request Guidelines

### PR Checklist

Before submitting, ensure:

- [ ] Code follows project style
- [ ] All tests pass
- [ ] New tests added for new features
- [ ] Documentation updated
- [ ] Commit messages are clear
- [ ] No merge conflicts
- [ ] PR description is complete

### PR Template

```markdown
## Description
Brief description of what this PR does.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring

## Testing
How did you test this?

## Screenshots (if applicable)
Add screenshots for UI changes.

## Checklist
- [ ] Tests pass
- [ ] Documentation updated
- [ ] Code follows style guidelines
```

## Review Process

1. **Automated Checks** - CI runs tests and linters
2. **Code Review** - Maintainers review your code
3. **Feedback** - Address any requested changes
4. **Approval** - Once approved, we'll merge!

## Getting Help

Need help? Here's how:

- 💬 **Discord**: [Join our community](https://discord.gg/Bg3gDAqDwz) - Chat with other contributors!
- 💬 **Discussions**: [GitHub Discussions](https://github.com/SagheerAkram/QuickCmd/discussions)
- 🐛 **Issues**: [GitHub Issues](https://github.com/SagheerAkram/QuickCmd/issues)
- 📧 **Contact**: Reach out to [@SagheerAkram](https://github.com/SagheerAkram)

## Code of Conduct

### Our Standards

- ✅ Be respectful and inclusive
- ✅ Welcome newcomers
- ✅ Accept constructive criticism
- ✅ Focus on what's best for the project

### Unacceptable Behavior

- ❌ Harassment or discrimination
- ❌ Trolling or insulting comments
- ❌ Personal attacks
- ❌ Publishing others' private information

## Recognition

Contributors will be:
- Listed in the project README
- Mentioned in release notes
- Given credit in commit history

## Questions?

Don't hesitate to ask! We're here to help:
- Open an issue with the `question` label
- Start a discussion
- Reach out to the maintainers

---

**Thank you for contributing to QuickCMD! 🚀**

*Built with ❤️ by the QuickCMD community*
