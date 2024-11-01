package text

import (
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/logic"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilderSimple_Build(t *testing.T) {
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
    networks: # Network configuration
      - innernet
    ports: # Port mapping
      - "0:7001"
    environment: # Environment variables
      - REDIS_URL=redis:6379
    depends_on: # Dependencies
      - redis
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "10"`,

		filepath.Join(logic.ServicesDirectoryConst, "redis.yml"): `  redis:
    image: redis:alpine # Redis image to use
    container_name: go-redis
    networks:
      - innernet
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data # Volume to persist Redis data
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "10"`,
	}

	tests := []struct {
		name           string
		setupFiles     bool             // Whether to set up the standard test files
		setupFunc      func(dir string) // Additional setup if needed
		forceOverwrite bool
		expectedError  bool
		validateFunc   func(t *testing.T, outputContent string)
	}{
		{
			name:           "successful_build",
			setupFiles:     true,
			forceOverwrite: true,
			validateFunc: func(t *testing.T, outputContent string) {
				// Check for comment preservation
				assert.Contains(t, outputContent, "# Docker Compose configuration for the Go-Redis application")
				assert.Contains(t, outputContent, "# Volume configuration 1")
				assert.Contains(t, outputContent, "# Volume configuration 2")
				assert.Contains(t, outputContent, "# Volume configuration 3")
				assert.Contains(t, outputContent, "# Network configuration for services")

				// Check for volume configuration
				assert.Contains(t, outputContent, "redis-data:")
				assert.Contains(t, outputContent, "name: go-redis")

				// Check for network configuration
				assert.Contains(t, outputContent, "innernet:")
				assert.Contains(t, outputContent, "driver: bridge")
				assert.Contains(t, outputContent, "com.docker.network.bridge.enable_ip_masquerade: \"true\"")

				// Check for service configuration
				assert.Contains(t, outputContent, "app:")
				assert.Contains(t, outputContent, "context: ./app")
				assert.Contains(t, outputContent, "container_name: go-redis-app")

				assert.Contains(t, outputContent, "redis:")
				assert.Contains(t, outputContent, "image: redis:alpine")
				assert.Contains(t, outputContent, "container_name: go-redis")

				// Verify services indentation
				lines := strings.Split(outputContent, "\n")
				for _, line := range lines {
					if strings.HasPrefix(strings.TrimSpace(line), "app:") ||
						strings.HasPrefix(strings.TrimSpace(line), "redis:") {
						assert.True(t, strings.HasPrefix(line, "  "),
							"Service definitions should have correct indentation")
					}
				}
			},
		},
		{
			name:           "existing_output_file_without_force",
			setupFiles:     true,
			forceOverwrite: false,
			setupFunc: func(dir string) {
				// Create existing output file
				err := os.WriteFile(
					filepath.Join(dir, logic.ComposeFileNameConst),
					[]byte("existing content"),
					0644,
				)
				assert.NoError(t, err)

				// In Builder, we need to simulate user input for 'N'
				// Since we can't easily mock the input in this test, we'll let the error
				// come from the existence check instead
			},
			expectedError: true, // Expected error from existence check
		},
		{
			name:           "missing_template_file",
			setupFiles:     false,
			forceOverwrite: true,
			expectedError:  true,
		},
		{
			name:           "missing_services_directory",
			setupFiles:     false,
			forceOverwrite: true,
			setupFunc: func(dir string) {
				// Create only the template file
				err := os.WriteFile(
					filepath.Join(dir, logic.TemplateFileNameDefaultConst),
					[]byte(testFiles[logic.TemplateFileNameDefaultConst]),
					0644,
				)
				assert.NoError(t, err)
			},
			expectedError: true,
		},
		{
			name:           "force_overwrite_existing_file",
			setupFiles:     true,
			forceOverwrite: true,
			setupFunc: func(dir string) {
				err := os.WriteFile(
					filepath.Join(dir, logic.ComposeFileNameConst),
					[]byte("old content"),
					0644,
				)
				assert.NoError(t, err)
			},
			validateFunc: func(t *testing.T, outputContent string) {
				assert.NotContains(t, outputContent, "old content")
				assert.Contains(t, outputContent, "app:")
				assert.Contains(t, outputContent, "redis:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a subdirectory for this test case
			testDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			assert.NoError(t, err)

			if tt.setupFiles {
				// Create necessary directories and files
				err = os.Mkdir(filepath.Join(testDir, logic.ServicesDirectoryConst), 0755)
				assert.NoError(t, err)

				// Create all test files
				for filename, content := range testFiles {
					filePath := filepath.Join(testDir, filename)
					err := os.MkdirAll(filepath.Dir(filePath), 0755)
					assert.NoError(t, err)
					err = os.WriteFile(filePath, []byte(content), 0644)
					assert.NoError(t, err)
				}
			}

			// Run additional setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc(testDir)
			}

			// Create builder instance with Builder
			builder := NewBuilder(
				testDir,
				filepath.Join(testDir, logic.TemplateFileNameDefaultConst),
				filepath.Join(testDir, logic.ServicesDirectoryConst),
				filepath.Join(testDir, logic.ComposeFileNameConst),
				tt.forceOverwrite,
			)

			// Execute build
			err = builder.Build()

			// Verify error expectation
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// If validation function is provided, read output and validate
			if tt.validateFunc != nil {
				content, err := os.ReadFile(filepath.Join(testDir, logic.ComposeFileNameConst))
				assert.NoError(t, err)
				tt.validateFunc(t, string(content))
			}
		})
	}
}

func TestBuilderSimple_Build_PreservesCommentedValues(t *testing.T) {
	tempDir := t.TempDir()

	// Source docker-compose file content with commented values
	testFiles := map[string]string{
		logic.TemplateFileNameDefaultConst: `services:
<dcm: include services\>`,

		filepath.Join(logic.ServicesDirectoryConst, "app.yml"): `  app:
    environment: # Environment variables
      #- REDIS_URL=redis:6001 # Old redis port
      - REDIS_URL=redis:6379 # Current redis port
      #- REDIS_URL=redis:7001 # Future redis port
    ports: # Port mappings
      #- "0:7000" # Old port
      - "0:7001" # Current port
      #- "0:7002" # Future port
    logging: # Logging configuration
      driver: "json-file" # JSON log driver
      #driver: "syslog" # Alternative driver
      options:
        #max-size: "5m" # Old size
        max-size: "10m" # Current size
        #max-size: "20m" # Future size
        max-file: "10"`,
	}

	// Setup test environment
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

	// Create and run builder
	builder := NewBuilder(
		tempDir,
		filepath.Join(tempDir, logic.TemplateFileNameDefaultConst),
		filepath.Join(tempDir, logic.ServicesDirectoryConst),
		filepath.Join(tempDir, logic.ComposeFileNameConst),
		true,
	)

	err = builder.Build()
	assert.NoError(t, err)

	// Read and verify output
	content, err := os.ReadFile(filepath.Join(tempDir, logic.ComposeFileNameConst))
	assert.NoError(t, err)

	outputContent := string(content)

	// Verify commented values are preserved
	assert.Contains(t, outputContent, "#- REDIS_URL=redis:6001 # Old redis port")
	assert.Contains(t, outputContent, "- REDIS_URL=redis:6379 # Current redis port")
	assert.Contains(t, outputContent, "#- REDIS_URL=redis:7001 # Future redis port")
	assert.Contains(t, outputContent, "#- \"0:7000\" # Old port")
	assert.Contains(t, outputContent, "- \"0:7001\" # Current port")
	assert.Contains(t, outputContent, "#- \"0:7002\" # Future port")
	assert.Contains(t, outputContent, "#driver: \"syslog\" # Alternative driver")
	assert.Contains(t, outputContent, "#max-size: \"5m\" # Old size")
	assert.Contains(t, outputContent, "max-size: \"10m\" # Current size")
	assert.Contains(t, outputContent, "#max-size: \"20m\" # Future size")
}
