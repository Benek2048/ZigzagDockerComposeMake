![ZigZag Everyday Tools](internal/assets/logo/ZigZagEverydayTools.png "ZigZag Everyday Tools")

# ZigzagDockerComposeMake

## Introduction

This program is designed to split a docker-compose.yml file into smaller files and reassemble them back together.

It aims to help with daily work with docker-compose.yml files by allowing easy monitoring of changes made within individual services and enabling simple reuse of service declarations across different projects.

Main functions:
- splitting (decomposition) of docker-compose.yml into a template and service files
- assembling a complete docker-compose.yml from the template and service files

## Usage Example

### Decomposition Example

Given a docker-compose.yml file:

```yaml
# Docker Compose configuration for the Go-Redis application
services:
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
```

After running the command:
```bash
dcm decompose
```

The program will create the following file structure:

1. Template file docker-compose-dcm.yml:
```
# Docker Compose configuration for the Go-Redis application
services:
<dcm: include services\>

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
```

2. File services/app.yml:
```yaml
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
```

3. File services/redis.yml:
```yaml
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
```

### Building Example

With the above file structure, after running the command:
```bash
dcm build
```

The program will reassemble them back into a complete docker-compose.yml file, preserving all comments and formatting.

## Operating Modes

The program can operate in two modes:
- text mode (default) - processing at text level
- yaml mode (--yaml-mode) - processing using yaml parser

## Program Parameters

### For decompose command:
```
  -d, --directory string    working directory (default: current)
  -t, --template string     template filename (default: docker-compose-dcm.yml)
  -c, --compose string      compose filename (default: docker-compose.yml)
  -f, --force               force overwrite existing files
      --yaml-mode           use yaml mode
```

### For build command:
```
  -d, --directory string    working directory (default: current)
  -t, --template string     template filename (default: docker-compose-dcm.yml)
  -c, --compose string      compose filename (default: docker-compose.yml)
  -f, --force               force overwrite existing files
      --yaml-mode           use yaml mode
```

## Project Structure

```
.
├── cmd/                 # CLI Commands
│   ├── build.go         # Build command implementation
│   ├── decompose.go     # Decompose command implementation
│   ├── root.go          # Main CLI configuration
│   └── version.go       # Version display command
│
├── internal/            # Internal application code
│   ├── assets/          # Static assets
│   ├── helper/          # Helper functions
│   │   ├── input/       # User input handling
│   │   └── path/        # Path operations
│   └── logic/           # Main business logic
│       ├── text/        # Text mode implementation
│       └── yaml/        # YAML mode implementation
│
├── bin/                 # Directory for executables
└── Makefile             # Compilation and testing scripts
```


## Project Compilation

The project uses Makefile to automate the build process. The same Makefile is used both for local development and GitHub Actions automated builds defined in `.github/workflows/release.yml`. This ensures consistency between local and CI/CD environments.

### Basic commands:

```bash
make all          # Executes all main tasks (test, build, build-linux, build-rpi, install)
make test         # Runs unit tests
make build        # Builds program for Windows (dcm.exe)
make install      # Installs program in GOBIN path
```

### Platform-specific commands:

```bash
make build        # Compiles Windows version (bin/dcm.exe)
make build-linux  # Compiles Linux version (bin/dcm)
make build-rpi    # Compiles Raspberry Pi version (bin/dcm-rpi)
make build-darwin # Compiles macOS version (bin/dcm-darwin-amd64, bin/dcm-darwin-arm64)
```

### CI/CD and maintenance commands:

```bash
make show_vars    # Displays current build variables (version, commit hash, etc.)
make update       # Updates version numbers across the project
make release      # Prepares a new release (updates version, builds all platforms)
make clean        # Cleans up build artifacts and temporary files
```

### Additional commands:

```bash
make refresh      # Refreshes project dependencies (go mod tidy)
```

### Compilation parameters:

- All binary versions are compiled with `-ldflags "-s -w"` flags to reduce file size
- Raspberry Pi version is compiled with additional parameters:
  - GOARCH='arm'
  - GOARM='7'
  - Using -trimpath flag

### Integration with GitHub Actions

The Makefile is designed to work seamlessly with GitHub Actions CI/CD pipeline defined in `.github/workflows/release.yml`. When making changes to the Makefile, always ensure:

1. Changes are compatible with both local development and GitHub Actions environment
2. New commands or modifications are reflected in the GitHub workflow file
3. Version handling remains consistent between local and CI builds
4. Build artifacts are generated in the expected locations for GitHub releases



### Prerequisites

The most convenient way to build the Windows resources is using a Linux environment, preferably:
- Native Linux system
- WSL (Windows Subsystem for Linux) with Ubuntu 24.04 or later

### Building the Resources

#### Windows Icon
When building the Windows version of the application, a special resource file `rsrc_windows_amd64.syso` is required to include the application icon in the executable. This file needs to be rebuilt if you modify the icon or other Windows-specific resources.


1. If using WSL or Linux, ensure you have the required tools:
   ```bash
   sudo apt-get update
   sudo apt-get install mingw-w64
   ```

2. Navigate to the resource's directory:
   ```bash
   cd internal/assets/windows
   ```

3. Run the build script:
   ```bash
   ./build-resources.sh
   ```

   The script will:
   - Check for required tools and install them if necessary
   - Create or update the `rsrc_windows_amd64.syso` file
   - Handle existing file conflicts

After running the script, the `rsrc_windows_amd64.syso` file will be updated with the new icon. You can now build the Windows version of the application.


## License

[Apache License 2.0](LICENSE)