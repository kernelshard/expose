package server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
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

func GenerateReadableSubdomain() (string, error) {
	adjective, err := randomWord(adjectives)
	if err != nil {
		return "", fmt.Errorf("failed to pick random adjective: %w", err)
	}

	noun, err := randomWord(nouns)
	if err != nil {
		return "", fmt.Errorf("failed to pick random noun: %w", err)
	}

	suffix, err := cryptoRandInt(9000)
	if err != nil {
		return "", fmt.Errorf("failed to pick random number: %w", err)
	}

	return fmt.Sprintf("%s-%s-%d", adjective, noun, suffix+1000), nil
}

func randomWord(words []string) (string, error) {
	index, err := cryptoRandInt(int64(len(words)))
	if err != nil {
		return "", err
	}
	return words[index], nil
}

func cryptoRandInt(max int64) (int64, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return nBig.Int64(), nil
}
