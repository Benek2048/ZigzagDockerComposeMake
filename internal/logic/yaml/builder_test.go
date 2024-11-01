// internal/logic/builder_test.go
package yaml

import (
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/logic"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBuilder_Build(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Create test files structure with real-world content and comments
	testFiles := map[string]string{
		logic.TemplateFileNameDefaultConst: `# Docker Compose configuration for the Go-Redis application
services:
<dcm: include services\>

# Volume configuration 1
# Volume configuration 2
# Volume configuration 3
volumes:
  redis-data:
    name: go-redis # Name of the volume

# Network configuration for services
networks:
  innernet:
    driver: bridge # Network driver to use
    driver_opts:
      com.docker.network.bridge.enable_ip_masquerade: "true" # Enable IP masquerading
      com.docker.network.bridge.enable_icc: "true" # Enable inter-container communication
      com.docker.network.driver.mtu: "1500" # Set the MTU for the network
      com.docker.network.bridge.name: "${BRIDGE_NAME}" # Name of the bridge
    name: "${NET_NAME}" # Name of the network
    ipam:
      driver: default # IPAM driver to use
      config:
        - subnet: 10.1.${NET_ID}.0/24 # Subnet configuration for the network`,

		filepath.Join(logic.ServicesDirectoryConst, "app.yml"): `  app:
    build:
      context: ./app # Path to the build context
      dockerfile: Dockerfile # Dockerfile to use for building the image
    container_name: go-redis-app # Name of the container
    networks:
      - innernet # Network to which the container will be connected
    ports:
      - "0:7001" # Port mapping for the container
    environment:
      #1- REDIS_URL=redis:6001  # Environment variable for Redis URL (commented out)
      #2- REDIS_URL=redis:6001  # Environment variable for Redis URL (commented out)
      #3- REDIS_URL=redis:6001  # Environment variable for Redis URL (commented out)

      - REDIS_URL=redis:6379 # Environment variable for Redis URL
    depends_on:
      - redis # Dependency on the Redis service
    restart: unless-stopped # Restart policy for the container
    logging:
      driver: "json-file" # Logging driver to use
      options:
        max-size: "10m" # Maximum size of the log file
        max-file: "10" # Maximum number of log files to retain`,

		filepath.Join(logic.ServicesDirectoryConst, "redis.yml"): `  redis:
    image: redis:alpine # Redis image to use
    container_name: go-redis # Name of the container
    networks:
      - innernet # Network to which the container will be connected
    ports:
      - "6379:6379" # Port mapping for the container
    volumes:
      - redis-data:/data # Volume to persist Redis data
    restart: unless-stopped # Restart policy for the container
    logging:
      driver: "json-file" # Logging driver to use
      options:
        max-size: "10m" # Maximum size of the log file
        max-file: "10" # Maximum number of log files to retain`,
	}

	// Create necessary directories and files
	err := os.Mkdir(filepath.Join(tempDir, logic.ServicesDirectoryConst), 0755)
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
		tempDir,
		filepath.Join(tempDir, logic.TemplateFileNameDefaultConst),
		filepath.Join(tempDir, logic.ServicesDirectoryConst),
		filepath.Join(tempDir, logic.ComposeFileNameConst),
		true,
	)

	err = builder.Build()
	assert.NoError(t, err)

	// Read the generated docker-compose.yml
	content, err := os.ReadFile(filepath.Join(tempDir, logic.ComposeFileNameConst))
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	generatedContent := string(content)

	// Test for preservation of specific comments and structure
	t.Run("Comments Preservation", func(t *testing.T) {
		// Check for template file comments
		assert.Contains(t, generatedContent, "# Docker Compose configuration for the Go-Redis application")
		assert.Contains(t, generatedContent, "# Volume configuration 1")
		assert.Contains(t, generatedContent, "# Volume configuration 2")
		assert.Contains(t, generatedContent, "# Volume configuration 3")
		assert.Contains(t, generatedContent, "# Network configuration for services")
		assert.Contains(t, generatedContent, "# Enable IP masquerading")
		assert.Contains(t, generatedContent, "# Enable inter-container communication")
		assert.Contains(t, generatedContent, "# Name of the bridge")
		assert.NotContains(t, generatedContent, "<dcm: include services\\>")

		// Check for app.yml comments
		assert.Contains(t, generatedContent, "# Path to the build context")
		assert.Contains(t, generatedContent, "# Dockerfile to use for building the image")
		assert.Contains(t, generatedContent, "# Name of the container")
		assert.Contains(t, generatedContent, "# Port mapping for the container")
		assert.Contains(t, generatedContent, "#1- REDIS_URL=redis:6001  # Environment variable for Redis URL (commented out)")
		assert.Contains(t, generatedContent, "#2- REDIS_URL=redis:6001  # Environment variable for Redis URL (commented out)")
		assert.Contains(t, generatedContent, "#3- REDIS_URL=redis:6001  # Environment variable for Redis URL (commented out)")
		assert.Contains(t, generatedContent, "# Environment variable for Redis URL")

		// Check for redis.yml comments
		assert.Contains(t, generatedContent, "# Redis image to use")
		assert.Contains(t, generatedContent, "# Volume to persist Redis data")
		assert.Contains(t, generatedContent, "# Maximum size of the log file")
		assert.Contains(t, generatedContent, "# Maximum number of log files to retain")
	})

	// Test for correct structure
	var result map[string]interface{}
	err = yaml.Unmarshal(content, &result)
	if err != nil {
		t.Fatalf("Failed to parse output YAML: %v", err)
	}

	// Verify services
	services, ok := result["services"].(map[string]interface{})
	assert.True(t, ok, "Services section not found or invalid")
	assert.Contains(t, services, "app")
	assert.Contains(t, services, "redis")

	// Verify configuration values
	t.Run("Configuration Values", func(t *testing.T) {
		// App service checks
		app := services["app"].(map[string]interface{})
		assert.Equal(t, "go-redis-app", app["container_name"])
		assert.Equal(t, "0:7001", app["ports"].([]interface{})[0])

		// Redis service checks
		redis := services["redis"].(map[string]interface{})
		assert.Equal(t, "redis:alpine", redis["image"])
		assert.Equal(t, "6379:6379", redis["ports"].([]interface{})[0])

		// Network config checks
		networks := result["networks"].(map[string]interface{})
		innernet := networks["innernet"].(map[string]interface{})
		assert.Equal(t, "bridge", innernet["driver"])
	})
}

func TestBuilder_Build_NoTemplate(t *testing.T) {
	tempDir := t.TempDir()

	builder := NewBuilder(
		tempDir,
		filepath.Join(tempDir, logic.TemplateFileNameDefaultConst),
		filepath.Join(tempDir, logic.ServicesDirectoryConst),
		filepath.Join(tempDir, logic.ComposeFileNameConst),
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
		filepath.Join(tempDir, logic.TemplateFileNameDefaultConst),
		[]byte("services:\n  <dcm: include services>"),
		0644,
	)
	if err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	builder := NewBuilder(
		tempDir,
		filepath.Join(tempDir, logic.TemplateFileNameDefaultConst),
		filepath.Join(tempDir, logic.ServicesDirectoryConst),
		filepath.Join(tempDir, logic.ComposeFileNameConst),
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
		filepath.Join(tempDir, logic.ComposeFileNameConst),
		[]byte("existing content"),
		0644,
	)
	if err != nil {
		t.Fatalf("Failed to create existing output file: %v", err)
	}

	builder := NewBuilder(
		tempDir,
		filepath.Join(tempDir, logic.TemplateFileNameDefaultConst),
		filepath.Join(tempDir, logic.ServicesDirectoryConst),
		filepath.Join(tempDir, logic.ComposeFileNameConst),
		false, // forceOverwrite = false
	)

	err = builder.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}
