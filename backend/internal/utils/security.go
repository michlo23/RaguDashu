package utils

import (
	"regexp"
	"strings"
)

// SanitizeErrorMessage removes sensitive information from error messages
// before storing them in database or sending to clients
func SanitizeErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Remove API keys (various patterns)
	patterns := []string{
		// OpenAI API keys (sk-proj-... or sk-...)
		`sk-proj-[A-Za-z0-9_-]{43,}`,
		`sk-[A-Za-z0-9]{48,}`,
		// Pinecone API keys (typically UUID format)
		`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
		// Generic API keys and tokens
		`[Aa][Pp][Ii][-_]?[Kk][Ee][Yy][-_:]?\s*[A-Za-z0-9_-]{20,}`,
		`[Tt][Oo][Kk][Ee][Nn][-_:]?\s*[A-Za-z0-9_-]{20,}`,
		// Bearer tokens
		`[Bb]earer\s+[A-Za-z0-9_-]{20,}`,
		// Base64 encoded secrets (40+ chars)
		`[A-Za-z0-9+/]{40,}={0,2}`,
	}

	sanitized := errMsg
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		sanitized = re.ReplaceAllString(sanitized, "[REDACTED]")
	}

	return sanitized
}

// SanitizeString removes sensitive data from any string
func SanitizeString(input string) string {
	// Similar to SanitizeErrorMessage but for plain strings
	patterns := []string{
		`sk-proj-[A-Za-z0-9_-]{43,}`,
		`sk-[A-Za-z0-9]{48,}`,
		`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
		`[Aa][Pp][Ii][-_]?[Kk][Ee][Yy][-_:]?\s*[A-Za-z0-9_-]{20,}`,
		`[Tt][Oo][Kk][Ee][Nn][-_:]?\s*[A-Za-z0-9_-]{20,}`,
		`[Bb]earer\s+[A-Za-z0-9_-]{20,}`,
		`[A-Za-z0-9+/]{40,}={0,2}`,
	}

	sanitized := input
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		sanitized = re.ReplaceAllString(sanitized, "[REDACTED]")
	}

	return sanitized
}

// GetGenericErrorMessage returns a safe error message without implementation details
func GetGenericErrorMessage(err error, category string) string {
	if err == nil {
		return ""
	}

	// Map specific errors to user-friendly messages
	errStr := strings.ToLower(err.Error())

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline") {
		return category + ": Request timeout. Please try again."
	}

	if strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "401") {
		return category + ": Invalid API credentials. Please check your API keys."
	}

	if strings.Contains(errStr, "forbidden") || strings.Contains(errStr, "403") {
		return category + ": Access forbidden. Please check your API key permissions."
	}

	if strings.Contains(errStr, "not found") || strings.Contains(errStr, "404") {
		return category + ": Resource not found."
	}

	if strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "429") {
		return category + ": Rate limit exceeded. Please try again later."
	}

	if strings.Contains(errStr, "network") || strings.Contains(errStr, "connection") {
		return category + ": Network error. Please check your connection."
	}

	// Generic safe message
	return category + ": Operation failed. Please try again or contact support."
}
