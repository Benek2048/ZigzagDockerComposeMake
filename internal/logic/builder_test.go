// internal/logic/builder_test.go
package logic

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBuilder_Build(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Create test files structure
	testFiles := map[string]string{
		TemplateFileNameDefaultConst: `
services:
  <dcm: include services>
volumes:
  data:
    name: test-data
`,
		filepath.Join(ServicesDirectoryConst, "app.yml"): `
  app:
    image: test-app
    ports:
      - "8080:8080"
`,
		filepath.Join(ServicesDirectoryConst, "redis.yml"): `
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
`,
	}

	// Create necessary directories and files
	err := os.Mkdir(filepath.Join(tempDir, ServicesDirectoryConst), 0755)
	if err != nil {
		t.Fatalf("Failed to create services directory: %v", err)
	}

	for filename, content := range testFiles {
		err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Run the builder
	builder := NewBuilder(
		filepath.Join(tempDir, TemplateFileNameDefaultConst),
		filepath.Join(tempDir, ServicesDirectoryConst),
		filepath.Join(tempDir, ComposeFileNameConst),
		true,
	)

	err = builder.Build()
	assert.NoError(t, err)

	// Verify the output
	content, err := os.ReadFile(filepath.Join(tempDir, ComposeFileNameConst))
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var result map[string]interface{}
	err = yaml.Unmarshal(content, &result)
	if err != nil {
		t.Fatalf("Failed to parse output YAML: %v", err)
	}

	// Verify the structure
	services, ok := result["services"].(map[string]interface{})
	assert.True(t, ok, "Services section not found or invalid")
	assert.Contains(t, services, "app", "App service not found")
	assert.Contains(t, services, "redis", "Redis service not found")

	volumes, ok := result["volumes"].(map[string]interface{})
	assert.True(t, ok, "Volumes section not found or invalid")
	assert.Contains(t, volumes, "data", "Data volume not found")
}

func TestBuilder_Build_NoTemplate(t *testing.T) {
	tempDir := t.TempDir()

	builder := NewBuilder(
		filepath.Join(tempDir, TemplateFileNameDefaultConst),
		filepath.Join(tempDir, ServicesDirectoryConst),
		filepath.Join(tempDir, ComposeFileNameConst),
		true,
	)

	err := builder.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read template")
}

func TestBuilder_Build_NoServicesDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create template file
	err := os.WriteFile(
		filepath.Join(tempDir, TemplateFileNameDefaultConst),
		[]byte("services:\n  <dcm: include services>"),
		0644,
	)
	if err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	builder := NewBuilder(
		filepath.Join(tempDir, TemplateFileNameDefaultConst),
		filepath.Join(tempDir, ServicesDirectoryConst),
		filepath.Join(tempDir, ComposeFileNameConst),
		true,
	)

	err = builder.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read services")
}

func TestBuilder_Build_ExistingOutput(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	err := os.WriteFile(
		filepath.Join(tempDir, ComposeFileNameConst),
		[]byte("existing content"),
		0644,
	)
	if err != nil {
		t.Fatalf("Failed to create existing output file: %v", err)
	}

	builder := NewBuilder(
		filepath.Join(tempDir, TemplateFileNameDefaultConst),
		filepath.Join(tempDir, ServicesDirectoryConst),
		filepath.Join(tempDir, ComposeFileNameConst),
		false, // forceOverwrite = false
	)

	err = builder.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}
