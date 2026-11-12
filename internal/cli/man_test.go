package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestManPage(t *testing.T) {
	// Locate man/expose.1 relative to internal/cli directory
	manPath := filepath.Join("..", "..", "man", "expose.1")
	contentBytes, err := os.ReadFile(manPath)
	if err != nil {
		t.Fatalf("Failed to read man page at %s: %v", manPath, err)
	}
	content := string(contentBytes)

	// Verify essential man sections
	requiredSections := []string{
		".TH EXPOSE 1",
		".SH NAME",
		".SH SYNOPSIS",
		".SH DESCRIPTION",
		".SH COMMANDS",
		".SH OPTIONS",
		".SH CONFIGURATION",
		".SH FILES",
		".SH ENVIRONMENT",
		".SH EXIT STATUS",
		".SH EXAMPLES",
		".SH INSTALLATION",
		".SH SEE ALSO",
	}

	for _, sec := range requiredSections {
		if !strings.Contains(content, sec) {
			t.Errorf("Man page is missing required section/macro: %s", sec)
		}
	}

	// Normalize roff dashes (\-) to standard dashes (-) for content checking
	plainContent := strings.ReplaceAll(content, `\-`, "-")

	// Verify all Cobra commands are documented
	root := NewRootCmd()
	for _, cmd := range root.Commands() {
		name := cmd.Name()
		if !strings.Contains(plainContent, name) {
			t.Errorf("Subcommand %q is not mentioned in man page", name)
		}

		// Verify flags for each command
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			longFlag := "--" + f.Name
			if !strings.Contains(plainContent, longFlag) {
				t.Errorf("Flag %q for command %q is not documented in man page", longFlag, name)
			}
		})
	}

	// If mandoc is available on the test runner, ensure 0 lint errors
	mandocPath, err := exec.LookPath("mandoc")
	if err == nil {
		out, err := exec.Command(mandocPath, "-Tlint", manPath).CombinedOutput()
		if err != nil {
			t.Errorf("mandoc -Tlint failed: %v\nOutput: %s", err, string(out))
		}
	}
}
