package server

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

// RFC 1123: alphanumeric, hyphens allowed in middle, 1 to 63 chars
var (
	subdomainRegex     = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	reservedSubdomains = map[string]struct{}{
		"admin":     {},
		"api":       {},
		"dashboard": {},
		"internal":  {},
		"localhost": {},
		"root":      {},
		"www":       {},
	}
	defaultSubdomainLength = 8
)

// ValidateSubdomain checkes if a subdomain compiles with RFC 1123 and is not reserved
func ValidateSubdomain(sub string) error {
	if len(sub) == 0 {
		return errors.New("subdomain cannot be empty")
	}
	if len(sub) > 63 {
		return errors.New("subdomain cannot exceed 63 characters")
	}
	if !subdomainRegex.MatchString(sub) {
		return errors.New("subdomain must contain only lowercase letteers, numbers, and hyphens (cannot start or end with a hyphen)")
	}
	if _, ok := reservedSubdomains[sub]; ok {
		return fmt.Errorf("subdomain %q is reserved", sub)
	}
	return nil
}

// GenerateRandomSubdomain generates a random alphanumeric subdomain
func GenerateRandomSubdomain(length int) (string, error) {
	if length <= 0 || length > 63 {
		return "", fmt.Errorf("length must be between 1 and 63, got %d", length)
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
