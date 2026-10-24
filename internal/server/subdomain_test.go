package server

import (
	"strings"
	"testing"
)

func TestGenerateReadableSubdomain(t *testing.T) {
	sub1, err := GenerateReadableSubdomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sub2, err := GenerateReadableSubdomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub1 == sub2 {
		t.Fatalf("subdomains are not random")
	}

	// Must satisfy RFC 1123 validation
	if err := ValidateSubdomain(sub1); err != nil {
		t.Fatalf("sub1 %q failed validation: %v", sub1, err)
	}
	if err := ValidateSubdomain(sub2); err != nil {
		t.Fatalf("sub2 %q failed validation: %v", sub2, err)
	}

	// Must match expected adjective-noun-4digits structure
	parts := strings.Split(sub1, "-")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts (adjective-noun-number), got %q", sub1)
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 || len(parts[2]) != 4 {
		t.Fatalf("invalid format in %q", sub1)
	}
}

func TestValidateSubdomain(t *testing.T) {
	tests := []struct {
		name    string
		sub     string
		wantErr bool
	}{
		// Valid cases
		{"valid single char", "a", false},
		{"valid standard", "my-app", false},
		{"valid alphanumeric", "app-123", false},
		{"valid 63 chars", strings.Repeat("a", 63), false},

		// Invalid lengths
		{"empty string", "", true},
		{"exceeds 63 chars", strings.Repeat("a", 64), true},

		// Invalid characters & RFC rules
		{"leading hyphen", "-myapp", true},
		{"trailing hyphen", "myapp-", true},
		{"uppercase rejected", "MyApp", true},
		{"dots rejected", "my.app", true},
		{"underscore rejected", "my_app", true},
		{"space rejected", "my app", true},
		{"special char rejected", "my$app", true},

		// Reserved keywords
		{"reserved admin", "admin", true},
		{"reserved api", "api", true},
		{"reserved localhost", "localhost", true},
		{"reserved www", "www", true},
		{"reserved dashboard", "dashboard", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubdomain(tt.sub)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSubdomain(%q) error = %v, wantErr = %v", tt.sub, err, tt.wantErr)
			}
		})
	}
}
