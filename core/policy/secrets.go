package policy

import (
	"regexp"
	"strings"
)

// SecretRedactor handles redaction of sensitive information
type SecretRedactor struct {
	patterns []*regexp.Regexp
}

// NewSecretRedactor creates a new secret redactor with default patterns
func NewSecretRedactor() *SecretRedactor {
	defaultPatterns := []string{
		// Environment variables
		`(?i)(password|passwd|pwd)\s*=\s*[^\s]+`,
		`(?i)(token|auth_token|access_token)\s*=\s*[^\s]+`,
		`(?i)(api[_-]?key|apikey)\s*=\s*[^\s]+`,
		`(?i)(secret|secret_key)\s*=\s*[^\s]+`,
		
		// Common credential formats
		`(?i)Bearer\s+[A-Za-z0-9\-._~+/]+=*`,
		`(?i)Basic\s+[A-Za-z0-9+/]+=*`,
		
		// AWS credentials
		`(?i)(aws[_-]?access[_-]?key[_-]?id)\s*=\s*[A-Z0-9]{20}`,
		`(?i)(aws[_-]?secret[_-]?access[_-]?key)\s*=\s*[A-Za-z0-9/+=]{40}`,
		
		// Private keys
		`-----BEGIN\s+(?:RSA\s+)?PRIVATE\s+KEY-----[\s\S]*?-----END\s+(?:RSA\s+)?PRIVATE\s+KEY-----`,
		
		// Database connection strings
		`(?i)(mysql|postgres|mongodb)://[^:]+:[^@]+@`,
	}
	
	var compiled []*regexp.Regexp
	for _, pattern := range defaultPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, re)
		}
	}
	
	return &SecretRedactor{
		patterns: compiled,
	}
}

// Redact redacts sensitive information from a string
func (sr *SecretRedactor) Redact(text string) string {
	result := text
	
	for _, pattern := range sr.patterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// Keep the key name but redact the value
			parts := strings.SplitN(match, "=", 2)
			if len(parts) == 2 {
				return parts[0] + "=***REDACTED***"
			}
			
			// For patterns without '=', redact the entire match
			return "***REDACTED***"
		})
	}
	
	return result
}

// RedactEnvVars redacts environment variables from a command
func (sr *SecretRedactor) RedactEnvVars(command string) string {
	// Pattern for environment variable assignments
	envPattern := regexp.MustCompile(`([A-Z_][A-Z0-9_]*)\s*=\s*([^\s]+)`)
	
	return envPattern.ReplaceAllStringFunc(command, func(match string) string {
		parts := strings.SplitN(match, "=", 2)
		if len(parts) != 2 {
			return match
		}
		
		varName := strings.TrimSpace(parts[0])
		
		// List of sensitive variable name patterns
		sensitiveVars := []string{
			"PASSWORD", "PASSWD", "PWD",
			"TOKEN", "AUTH_TOKEN", "ACCESS_TOKEN",
			"API_KEY", "APIKEY",
			"SECRET", "SECRET_KEY",
			"PRIVATE_KEY",
			"DATABASE_URL", "DB_PASSWORD",
			"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY",
		}
		
		for _, sensitive := range sensitiveVars {
			if strings.Contains(strings.ToUpper(varName), sensitive) {
				return varName + "=***REDACTED***"
			}
		}
		
		return match
	})
}

// AddPattern adds a custom redaction pattern
func (sr *SecretRedactor) AddPattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	
	sr.patterns = append(sr.patterns, re)
	return nil
}
