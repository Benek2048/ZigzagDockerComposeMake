package text

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

	// Compile regular expressions
	servicesRe := regexp.MustCompile(`^services:\s*$`)
	serviceDefRe := regexp.MustCompile(`^(\s{2})([^: ]+):\s*.*$`) // Matches any service definition with exactly 2 spaces
	topLevelRe := regexp.MustCompile(`^[^: ]+:\s*$`)              // Matches top-level sections

	// Open source file
	file, err := os.Open(d.fileSrc)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer file.Close()

	// Create template file
	fileTemplate, err := os.Create(d.fileTemplate)
	if err != nil {
		return fmt.Errorf("failed to create template file: %w", err)
	}
	defer fileTemplate.Close()

	scanner := bufio.NewScanner(file)
	var templateBuilder strings.Builder
	var serviceBuilder strings.Builder
	var currentServiceName string
	inServices := false
	lastLineWasEmpty := false
	lastLineWasComment := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		isEmptyLine := len(trimmedLine) == 0
		isCommentLine := strings.HasPrefix(trimmedLine, "#")

		// Handle empty lines
		if isEmptyLine {
			if currentServiceName != "" {
				serviceBuilder.WriteString(line + "\n")
			} else if !inServices || !lastLineWasEmpty {
				templateBuilder.WriteString(line + "\n")
			}
			lastLineWasEmpty = true
			lastLineWasComment = false
			continue
		}
		lastLineWasEmpty = false

		// Check for services section
		if servicesRe.MatchString(line) {
			inServices = true
			templateBuilder.WriteString("services:\n")
			templateBuilder.WriteString("<dcm: include services\\>\n")
			lastLineWasComment = false
			continue
		}

		// Check for new top-level section
		if topLevelRe.MatchString(line) && !strings.HasPrefix(line, " ") {
			if currentServiceName != "" {
				// Save current service
				err := os.WriteFile(
					filepath.Join(d.servicesDir, currentServiceName+".yml"),
					[]byte(serviceBuilder.String()),
					0644,
				)
				if err != nil {
					return fmt.Errorf("failed to write service file: %w", err)
				}
				serviceBuilder.Reset()
				currentServiceName = ""
			}
			inServices = false

			// Add newline before section only if previous line wasn't a comment
			if !lastLineWasComment {
				templateBuilder.WriteString("\n")
			}
			templateBuilder.WriteString(line + "\n")
			lastLineWasComment = false
			continue
		}

		// Process services section content
		if inServices {
			if matches := serviceDefRe.FindStringSubmatch(line); matches != nil {
				// Save previous service if exists
				if currentServiceName != "" {
					err := os.WriteFile(
						filepath.Join(d.servicesDir, currentServiceName+".yml"),
						[]byte(serviceBuilder.String()),
						0644,
					)
					if err != nil {
						return fmt.Errorf("failed to write service file: %w", err)
					}
					serviceBuilder.Reset()
				}

				// Start new service
				currentServiceName = matches[2]
				serviceBuilder.WriteString(line + "\n")
				lastLineWasComment = false
				continue
			}

			// Add line to current service
			if currentServiceName != "" {
				serviceBuilder.WriteString(line + "\n")
			}
		} else {
			// Outside services section - add to template
			templateBuilder.WriteString(line + "\n")
		}

		lastLineWasComment = isCommentLine
	}

	// Save last service if exists
	if currentServiceName != "" {
		err := os.WriteFile(
			filepath.Join(d.servicesDir, currentServiceName+".yml"),
			[]byte(serviceBuilder.String()),
			0644,
		)
		if err != nil {
			return fmt.Errorf("failed to write service file: %w", err)
		}
	}

	// Write template file
	if _, err := fileTemplate.WriteString(templateBuilder.String()); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}
