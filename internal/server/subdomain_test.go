package server

import (
	"strings"
	"testing"
)

func TestGenerateRandomSubdomain(t *testing.T) {
	// Generate two random subdomains and check if they are different
	sub1, err := GenerateRandomSubdomain(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sub2, err := GenerateRandomSubdomain(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub1 == sub2 {
		t.Fatalf("subdomains are not random")
	}
	if len(sub2) != 8 {
		t.Fatalf("expected 8 characters, got %d", len(sub2))
	}
	if len(sub1) != 7 {
		t.Fatalf("expected 7 characters, got %d", len(sub1))
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
