package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad tests the Load function of the config package
func TestLoad(t *testing.T) {

	t.Run("valid config file", func(t *testing.T) {
		// create temporary config file
		content := []byte("project: test-project\nport: 8080\n")

		tempFile, err := os.CreateTemp("", "test-config-*.yml")
		if err != nil {
			t.Fatal(err)
		}

		defer os.Remove(tempFile.Name()) // clean up

		if _, err := tempFile.Write(content); err != nil {
			t.Fatal(err)
		}
		// close the file so it can be read, it's opened in write mode
		_ = tempFile.Close()

		// Test Load
		cfg, err := Load(tempFile.Name())
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.Project != "test-project" {
			t.Errorf("Expected project 'test-project', got '%s'", cfg.Project)
		}
		if cfg.Port != 8080 {
			t.Errorf("Expected port 8080, got %d", cfg.Port)
		}

	})

	t.Run("load with blank path uses default", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := filepath.Join(tmpDir, DefaultConfigFile)

		content := []byte("project: default-project\nport: 3000\n")
		tempFile, err := os.Create(filePath)
		if err != nil {
			t.Fatal(err)
		}

		// wrapping in closure to handle error
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				t.Error(err)
			}
		}(tempFile.Name()) // clean up

		if _, err := tempFile.Write(content); err != nil {
			t.Fatal(err)
		}
		_ = tempFile.Close()
		// change the working dir to the temp dir so Load uses the default path
		err = os.Chdir(tmpDir)
		if err != nil {
			t.Error(err)
		}

		cfg, err := Load("")
		if err != nil {
			t.Errorf("Expected no error loading default config, got %v", err)
		}

		if cfg.Project != "default-project" {
			t.Errorf("Expected project 'default-project', got '%s'", cfg.Project)
		}

		if cfg.Port != 3000 {
			t.Errorf("Expected port 3000, got %d", cfg.Port)
		}

	})

	t.Run("non existent file returns error", func(t *testing.T) {
		_, err := Load("nonexistent-config.yml")
		if err == nil {
			t.Errorf("Expected error for missing config file, got nil")
		}
	})

}

// TestConfigInit tests the Init function of the config package
func TestConfigInit(t *testing.T) {
	t.Run("error returned  when config exists", func(t *testing.T) {
		// Create a temp dir and a config file in it
		tmpDir := t.TempDir()
		tmpFilePath := filepath.Join(tmpDir, DefaultConfigFile)

		tmpFile, err := os.Create(tmpFilePath)
		if err != nil {
			t.Fatal(err)
		}

		defer os.Remove(tmpFile.Name())

		_ = tmpFile.Close()
		_ = os.Chdir(tmpDir)

		_, err = Init()
		if err == nil {
			t.Errorf("expected error that config already exists, got %v", err)
		}

	})
	// TODO: test init might need to change code as it's not testable
	t.Run("error returned when file is not yml formatted", func(t *testing.T) {

	})

	t.Run("creates config with default values", func(t *testing.T) {
		tmpDir := t.TempDir()
		_ = os.Chdir(tmpDir)

		cfg, err := Init()
		if err != nil {
			t.Fatalf("Init() failed: %v", err)
		}

		if cfg.Port != 3000 {
			t.Errorf("expected port 3000, got %d", cfg.Port)
		}

		// Verify file created
		if _, err := os.Stat(DefaultConfigFile); os.IsNotExist(err) {
			t.Error("config file not created")
		}
	})
}

// TestConfig_List tests the List method of the Config struct
func TestConfig_List(t *testing.T) {
	cfg := &Config{
		Project: "testProject",
		Port:    3000,
	}
	values := cfg.List()

	if values["project"] != "testProject" {
		t.Errorf("expected project=testProject, got %v", values["project"])

	}

	if values["port"] != 3000 {
		t.Errorf("expected port=3000, got %v", values["port"])
	}

}

func TestGet(t *testing.T) {
	cfg := &Config{
		Project: "my_project",
		Port:    3000,
	}

	tests := []struct {
		key      string
		expected interface{}
		wantErr  bool
	}{
		{"project", "my_project", false},
		{"port", 3000, false},
		{"invalid", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := cfg.Get(tt.key)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
