package logic

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

// Builder handles the process of combining separate service files into a complete docker-compose.yml
type Builder struct {
	templatePath   string
	servicesDir    string
	outputPath     string
	forceOverwrite bool
}

// NewBuilder creates a new instance of Builder with the specified paths and options
func NewBuilder(templatePath, servicesDir, outputPath string, forceOverwrite bool) *Builder {
	return &Builder{
		templatePath:   templatePath,
		servicesDir:    servicesDir,
		outputPath:     outputPath,
		forceOverwrite: forceOverwrite,
	}
}

// Build processes the template and service files to create a complete docker-compose.yml
func (b *Builder) Build() error {
	// Check if output file exists and handle force overwrite
	if !b.forceOverwrite {
		if _, err := os.Stat(b.outputPath); err == nil {
			return fmt.Errorf("output file %s already exists. Use --force to overwrite", b.outputPath)
		}
	}

	// Read and parse the template file
	template, err := b.readTemplate()
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Read all service definitions
	services, err := b.readServices()
	if err != nil {
		return fmt.Errorf("failed to read services: %w", err)
	}

	// Merge services into the template
	err = b.mergeServices(template, services)
	if err != nil {
		return fmt.Errorf("failed to merge services: %w", err)
	}

	// Write the final docker-compose.yml
	err = b.writeOutput(template)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// readTemplate reads and parses the template docker-compose file
func (b *Builder) readTemplate() (map[string]interface{}, error) {
	content, err := os.ReadFile(b.templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var template map[string]interface{}
	err = yaml.Unmarshal(content, &template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	return template, nil
}

// readServices reads all service definition files from the services directory
func (b *Builder) readServices() (map[string]interface{}, error) {
	services := make(map[string]interface{})

	files, err := os.ReadDir(b.servicesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read services directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(b.servicesDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read service file %s: %w", file.Name(), err)
		}

		var service map[string]interface{}
		err = yaml.Unmarshal(content, &service)
		if err != nil {
			return nil, fmt.Errorf("failed to parse service YAML %s: %w", file.Name(), err)
		}

		for k, v := range service {
			services[k] = v
		}
	}

	return services, nil
}

// mergeServices combines service definitions with the template
func (b *Builder) mergeServices(template, services map[string]interface{}) error {
	if _, ok := template["services"]; !ok {
		template["services"] = make(map[string]interface{})
	}

	servicesMap, ok := template["services"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid services section in template")
	}

	// Replace <dcm: include services> with actual services
	for k, v := range services {
		servicesMap[k] = v
	}

	return nil
}

// writeOutput writes the final docker-compose.yml file
func (b *Builder) writeOutput(data map[string]interface{}) error {
	output, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal output YAML: %w", err)
	}

	err = os.WriteFile(b.outputPath, output, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
