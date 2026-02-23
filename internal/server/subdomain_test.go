package server

import "testing"

func TestGenerateRandomSubdomain(t *testing.T) {
	// Generate two random subdomains and check if they are different
	sub1, err := GenerateRandomSubdomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sub2, err := GenerateRandomSubdomain()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub1 == sub2 {
		t.Fatalf("subdomains are not random")
	}
	if len(sub2) != 8 {
		t.Fatalf("expected 8 characters, got %d", len(sub2))
	}
}
