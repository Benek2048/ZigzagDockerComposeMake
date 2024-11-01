package yaml

import (
	"fmt"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/helper/input"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/helper/path"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/logic"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/logic/yaml/helper"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

// Builder handles the process of combining separate service files into a complete docker-compose.yml
type Builder struct {
	buildDir       string
	templatePath   string
	servicesDir    string
	outputPath     string
	forceOverwrite bool
}

// NewBuilder creates a new instance of Builder with the specified paths and options
func NewBuilder(buildDir, templatePath, servicesDir, outputPath string, forceOverwrite bool) *Builder {
	return &Builder{
		buildDir:       buildDir,
		templatePath:   templatePath,
		servicesDir:    servicesDir,
		outputPath:     outputPath,
		forceOverwrite: forceOverwrite,
	}
}

// Build processes the template and service files to create a complete docker-compose.yml
func (b *Builder) Build() error {
	// Check if the compose file exists
	composeFileExists, err := path.IsExist(b.outputPath)
	if composeFileExists && !b.forceOverwrite {
		fmt.Printf("Compose file '%v' already exists. Overwrite[y/N]?", logic.ComposeFileNameConst)
		answer := input.AskForYesOrNot("y", "N")
		if !answer {
			return fmt.Errorf("operation canceled")
		}
	}

	// Read and parse the template file
	templateNode, err := b.readTemplate()
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	// Read all service definitions
	services, err := b.readServices()
	if err != nil {
		return fmt.Errorf("failed to read services: %w", err)
	}

	// Merge services into the template
	err = b.mergeServices(templateNode, services)
	if err != nil {
		return fmt.Errorf("failed to merge services: %w", err)
	}

	if composeFileExists {
		// Create backup of existing file before overwriting
		if err := path.BackupExistingFile(b.outputPath); err != nil {
			return fmt.Errorf("Error creating backup: %v\n", err)
		}
	}

	// Write the final docker-compose.yml
	err = b.writeOutput(templateNode)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// readTemplate reads and parses the template docker-compose file preserving comments
func (b *Builder) readTemplate() (*yaml.Node, error) {
	content, err := os.ReadFile(b.templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var node yaml.Node
	err = yaml.Unmarshal(content, &node)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template YAML: %w", err)
	}

	return &node, nil
}

// readServices reads all service definition files from the services directory preserving comments
func (b *Builder) readServices() ([]*yaml.Node, error) {
	var services []*yaml.Node

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

		var node yaml.Node
		err = yaml.Unmarshal(content, &node)
		if err != nil {
			return nil, fmt.Errorf("failed to parse service YAML %s: %w", file.Name(), err)
		}

		if len(node.Content) > 0 {
			services = append(services, node.Content[0])
		}
	}

	return services, nil
}

// mergeServices combines service definitions with the template preserving comments
func (b *Builder) mergeServices(templateNode *yaml.Node, services []*yaml.Node) error {
	if templateNode.Kind != yaml.DocumentNode || len(templateNode.Content) == 0 {
		return fmt.Errorf("invalid template structure")
	}

	rootMap := templateNode.Content[0]
	servicesNode := helper.FindServicesNode(templateNode)
	if servicesNode == nil {
		return fmt.Errorf("services section not found in template")
	}

	// Create new mapping node for services
	newServicesNode := &yaml.Node{
		Kind:        yaml.MappingNode,
		Style:       servicesNode.Style,
		HeadComment: servicesNode.HeadComment,
		LineComment: servicesNode.LineComment,
		FootComment: "\n", // Add a blank line after the services section
	}

	// Add all service definitions
	for _, serviceNode := range services {
		for i := 0; i < len(serviceNode.Content); i += 2 {
			key := serviceNode.Content[i]
			value := serviceNode.Content[i+1]

			// Add a blank line before each service (except the first)
			if len(newServicesNode.Content) > 0 && key.HeadComment == "" {
				key.HeadComment = "\n"
			}

			newServicesNode.Content = append(newServicesNode.Content,
				key,   // key
				value, // value
			)
		}
	}

	// Replace the services node
	for i := 0; i < len(rootMap.Content); i += 2 {
		if rootMap.Content[i].Value == "services" {
			// Keep comments from the original services node
			rootMap.Content[i+1] = newServicesNode
			break
		}
	}

	// Add blank lines between main sections
	for i := 2; i < len(rootMap.Content); i += 2 {
		if rootMap.Content[i].HeadComment == "" {
			rootMap.Content[i].HeadComment = "\n"
		}
	}

	return nil
}

// writeOutput writes the final docker-compose.yml file preserving comments
func (b *Builder) writeOutput(node *yaml.Node) error {
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(node); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	output := buf.String()

	// Remove lines containing placeholder
	lines := strings.Split(output, "\n")
	var filteredLines []string
	for _, line := range lines {
		if !strings.Contains(line, "<dcm: include services") {
			filteredLines = append(filteredLines, line)
		}
	}
	output = strings.Join(filteredLines, "\n")

	// Make sure the file ends with a single blank line
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	err := os.WriteFile(b.outputPath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
