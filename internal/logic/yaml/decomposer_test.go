package yaml

import (
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/logic"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServiceDecomposer(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Source docker-compose file content with comments
	sourceContent := `# Main docker-compose configuration
services:
  # Application service configuration
  app: # Main application
    build: # Build configuration
      context: ./app
      dockerfile: Dockerfile
    container_name: go-redis-app # Container name
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
        max-file: "10"

  # Redis service configuration
  redis: # Cache service
    image: redis:alpine # Using Alpine for smaller footprint
    container_name: go-redis
    networks:
      - innernet
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "10"

# Volume configuration
volumes:
  redis-data:
    name: go-redis

# Network configuration for services
networks:
  innernet:
    driver: bridge
    driver_opts:
      com.docker.network.bridge.enable_ip_masquerade: "true"
      com.docker.network.bridge.enable_icc: "true"
      com.docker.network.driver.mtu: "1500"
      com.docker.network.bridge.name: "${BRIDGE_NAME}"
    name: "${NET_NAME}"
    ipam:
      driver: default
      config:
        - subnet: 10.1.${NET_ID}.0/24`

	// Expected content for generated files
	expectedTemplate := `# Main docker-compose configuration
services:
<dcm: include services\>

# Volume configuration
volumes:
  redis-data:
    name: go-redis

# Network configuration for services
networks:
  innernet:
    driver: bridge
    driver_opts:
      com.docker.network.bridge.enable_ip_masquerade: "true"
      com.docker.network.bridge.enable_icc: "true"
      com.docker.network.driver.mtu: "1500"
      com.docker.network.bridge.name: "${BRIDGE_NAME}"
    name: "${NET_NAME}"
    ipam:
      driver: default
      config:
        - subnet: 10.1.${NET_ID}.0/24
`

	expectedAppService := `# Application service configuration
app: # Main application
  build: # Build configuration
    context: ./app
    dockerfile: Dockerfile
  container_name: go-redis-app # Container name
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
      max-file: "10"
`

	expectedRedisService := `# Redis service configuration
redis: # Cache service
  image: redis:alpine # Using Alpine for smaller footprint
  container_name: go-redis
  networks:
    - innernet
  ports:
    - "6379:6379"
  volumes:
    - redis-data:/data
  restart: unless-stopped
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "10"
`

	// Create test files and directories
	fileSrc := filepath.Join(tmpDir, logic.ComposeFileNameConst)
	fileTemplate := filepath.Join(tmpDir, logic.TemplateFileNameDefaultConst)
	servicesDir := filepath.Join(tmpDir, logic.ServicesDirectoryConst)

	// Write source file
	err := os.WriteFile(fileSrc, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create and run decomposer
	decomposer := NewServiceDecomposer(fileSrc, fileTemplate, servicesDir)
	err = decomposer.Decompose()
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Helper function to compare file contents ignoring whitespace differences
	compareContents := func(expected, actual string) bool {
		// Normalize line endings and trim spaces
		expected = strings.TrimSpace(strings.ReplaceAll(expected, "\r\n", "\n"))
		actual = strings.TrimSpace(strings.ReplaceAll(actual, "\r\n", "\n"))
		return expected == actual
	}

	// Test template file
	templateContent, err := os.ReadFile(fileTemplate)
	if err != nil {
		t.Fatalf("Failed to read template file: %v", err)
	}
	if !compareContents(expectedTemplate, string(templateContent)) {
		t.Errorf("Template file content mismatch.\nExpected:\n%s\nGot:\n%s", expectedTemplate, string(templateContent))
	}

	// Test app service file
	appContent, err := os.ReadFile(filepath.Join(servicesDir, "app.yml"))
	if err != nil {
		t.Fatalf("Failed to read app service file: %v", err)
	}
	if !compareContents(expectedAppService, string(appContent)) {
		t.Errorf("App service file content mismatch.\nExpected:\n%s\nGot:\n%s", expectedAppService, string(appContent))
	}

	// Test redis service file
	redisContent, err := os.ReadFile(filepath.Join(servicesDir, "redis.yml"))
	if err != nil {
		t.Fatalf("Failed to read redis service file: %v", err)
	}
	if !compareContents(expectedRedisService, string(redisContent)) {
		t.Errorf("Redis service file content mismatch.\nExpected:\n%s\nGot:\n%s", expectedRedisService, string(redisContent))
	}

	// Additional checks
	t.Run("Check services directory created", func(t *testing.T) {
		if _, err := os.Stat(servicesDir); os.IsNotExist(err) {
			t.Error("Services directory was not created")
		}
	})

	t.Run("Check number of service files", func(t *testing.T) {
		files, err := os.ReadDir(servicesDir)
		if err != nil {
			t.Fatalf("Failed to read services directory: %v", err)
		}
		if len(files) != 2 {
			t.Errorf("Expected 2 service files, got %d", len(files))
		}
	})
}

func TestServiceDecomposerPreservesCommentedValues(t *testing.T) {
	tmpDir := t.TempDir()

	// Source docker-compose file content with commented values
	sourceContent := `services:
  app:
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
        max-file: "10"`

	expectedAppService := `app:
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
      max-file: "10"
`
	// Create test files and directories
	fileSrc := filepath.Join(tmpDir, logic.ComposeFileNameConst)
	fileTemplate := filepath.Join(tmpDir, logic.TemplateFileNameDefaultConst)
	servicesDir := filepath.Join(tmpDir, logic.ServicesDirectoryConst)

	// Write source file
	err := os.WriteFile(fileSrc, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create and run decomposer
	decomposer := NewServiceDecomposer(fileSrc, fileTemplate, servicesDir)
	err = decomposer.Decompose()
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Read generated app service file
	appContent, err := os.ReadFile(filepath.Join(servicesDir, "app.yml"))
	if err != nil {
		t.Fatalf("Failed to read app service file: %v", err)
	}

	// Helper function to compare file contents ignoring whitespace differences
	compareContents := func(expected, actual string) bool {
		// Normalize line endings and trim spaces
		expected = strings.TrimSpace(strings.ReplaceAll(expected, "\r\n", "\n"))
		actual = strings.TrimSpace(strings.ReplaceAll(actual, "\r\n", "\n"))

		// Print both versions for debugging if they don't match
		if expected != actual {
			t.Logf("Expected:\n%s\n", expected)
			t.Logf("Got:\n%s\n", actual)

			// Print each line for easier comparison
			expectedLines := strings.Split(expected, "\n")
			actualLines := strings.Split(actual, "\n")

			minLen := len(expectedLines)
			if len(actualLines) < minLen {
				minLen = len(actualLines)
			}

			for i := 0; i < minLen; i++ {
				if expectedLines[i] != actualLines[i] {
					t.Logf("Line %d differs:\nExpected: %q\nGot:      %q\n",
						i+1, expectedLines[i], actualLines[i])
				}
			}

			if len(expectedLines) != len(actualLines) {
				t.Logf("Line count differs: expected %d, got %d",
					len(expectedLines), len(actualLines))
			}
		}

		return expected == actual
	}

	// Compare content
	if !compareContents(expectedAppService, string(appContent)) {
		t.Errorf("App service file content mismatch. See above for detailed comparison.")
	}
}
