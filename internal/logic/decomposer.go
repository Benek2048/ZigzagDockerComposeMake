package logic

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

// ServiceDecomposer handles the decomposition of docker-compose services into separate files
type ServiceDecomposer struct {
	fileSrc      string
	fileTemplate string
	servicesDir  string
}

// NewServiceDecomposer creates a new instance of ServiceDecomposer
func NewServiceDecomposer(fileSrc, fileTemplate, servicesDir string) *ServiceDecomposer {
	return &ServiceDecomposer{
		fileSrc:      fileSrc,
		fileTemplate: fileTemplate,
		servicesDir:  servicesDir,
	}
}

// Decompose performs the main decomposition logic
func (d *ServiceDecomposer) Decompose() error {
	// Create services directory if it doesn't exist
	if err := os.MkdirAll(d.servicesDir, 0755); err != nil {
		return fmt.Errorf("failed to create services directory: %w", err)
	}

	// Read the source file
	node, err := d.parseSourceFile()
	if err != nil {
		return fmt.Errorf("failed to parse source file: %w", err)
	}

	// Extract and write services
	if err := d.extractServices(node); err != nil {
		return fmt.Errorf("failed to extract services: %w", err)
	}

	// Create template file
	if err := d.createTemplateFile(node); err != nil {
		return fmt.Errorf("failed to create template file: %w", err)
	}

	return nil
}

// parseSourceFile reads and parses the source docker-compose file preserving comments
func (d *ServiceDecomposer) parseSourceFile() (*yaml.Node, error) {
	file, err := os.ReadFile(d.fileSrc)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}

	var node yaml.Node
	if err := yaml.Unmarshal(file, &node); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &node, nil
}

// findServicesNode locates the services section in the YAML tree
func findServicesNode(node *yaml.Node) *yaml.Node {
	if node.Kind != yaml.DocumentNode {
		return nil
	}

	rootMap := node.Content[0]
	for i := 0; i < len(rootMap.Content); i += 2 {
		if rootMap.Content[i].Value == "services" {
			return rootMap.Content[i+1]
		}
	}
	return nil
}

// extractServices extracts individual services and writes them to separate files
func (d *ServiceDecomposer) extractServices(node *yaml.Node) error {
	servicesNode := findServicesNode(node)
	if servicesNode == nil {
		return fmt.Errorf("services section not found in source file")
	}

	// Process each service
	for i := 0; i < len(servicesNode.Content); i += 2 {
		serviceName := servicesNode.Content[i].Value
		serviceNode := servicesNode.Content[i+1]

		// Create service YAML document
		serviceDoc := &yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{
					Kind:  yaml.MappingNode,
					Style: yaml.LiteralStyle,
					Content: []*yaml.Node{
						{
							Kind:        yaml.ScalarNode,
							Value:       serviceName,
							Style:       servicesNode.Content[i].Style,
							HeadComment: servicesNode.Content[i].HeadComment,
							LineComment: servicesNode.Content[i].LineComment,
							FootComment: servicesNode.Content[i].FootComment,
						},
						serviceNode,
					},
				},
			},
		}

		// Marshal service to YAML
		var buf strings.Builder
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err := encoder.Encode(serviceDoc); err != nil {
			return fmt.Errorf("failed to marshal service %s: %w", serviceName, err)
		}

		// Write service file
		filename := filepath.Join(d.servicesDir, fmt.Sprintf("%s.yml", serviceName))
		if err := os.WriteFile(filename, []byte(buf.String()), 0644); err != nil {
			return fmt.Errorf("failed to write service file %s: %w", filename, err)
		}
	}

	return nil
}

// createTemplateFile creates the template file with service inclusion directive
func (d *ServiceDecomposer) createTemplateFile(node *yaml.Node) error {
	rootMap := node.Content[0]

	// Find services section
	for i := 0; i < len(rootMap.Content); i += 2 {
		// Handle services section
		if rootMap.Content[i].Value == "services" {
			headComment := rootMap.Content[i].HeadComment
			lineComment := rootMap.Content[i].LineComment

			servicesKeyNode := &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       "services",
				HeadComment: headComment,
				LineComment: lineComment,
			}

			includeNode := &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       "<dcm: include services\\>",
				Style:       yaml.LiteralStyle,
				HeadComment: "\n",
				FootComment: "\n",
			}

			rootMap.Content[i] = servicesKeyNode
			rootMap.Content[i+1] = includeNode
		}

		// Add extra newline before each main section (volumes, networks)
		if i > 0 && (rootMap.Content[i].Value == "volumes" || rootMap.Content[i].Value == "networks") {
			currentComment := rootMap.Content[i].HeadComment
			if !strings.HasPrefix(currentComment, "\n") {
				rootMap.Content[i].HeadComment = "\n" + currentComment
			}
		}
	}

	// Marshal to YAML
	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(node); err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	// Post-process the content
	content := buf.String()

	// Fix the services section formatting
	content = strings.ReplaceAll(content, "services: |-\n  <dcm:", "services: \n<dcm:")
	content = strings.ReplaceAll(content, "services: |2-\n  <dcm:", "services: \n<dcm:")
	content = strings.ReplaceAll(content, "services: |\n  <dcm:", "services: \n<dcm:")
	content = strings.ReplaceAll(content, "services: >-\n  <dcm:", "services: \n<dcm:")
	content = strings.ReplaceAll(content, "services:\n  <dcm:", "services: \n<dcm:")

	// Clean up duplicate newlines while preserving intended spacing
	lines := strings.Split(content, "\n")
	var processedLines []string
	var previousEmpty, previousWasComment bool

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		isEmpty := trimmedLine == ""
		isComment := strings.HasPrefix(trimmedLine, "#")

		// Skip if we have three consecutive empty lines
		if isEmpty && previousEmpty && len(processedLines) > 0 && strings.TrimSpace(processedLines[len(processedLines)-1]) == "" {
			continue
		}

		// Ensure empty line before comments that start main sections (except the first one)
		if isComment && !previousWasComment && len(processedLines) > 0 && !strings.HasPrefix(trimmedLine, "# Main") {
			if !previousEmpty {
				processedLines = append(processedLines, "")
			}
		}

		processedLines = append(processedLines, line)
		previousEmpty = isEmpty
		previousWasComment = isComment
	}

	content = strings.Join(processedLines, "\n") + "\n"

	// Write the file
	if err := os.WriteFile(d.fileTemplate, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}
