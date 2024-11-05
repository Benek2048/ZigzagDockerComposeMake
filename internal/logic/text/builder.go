package text

import (
	"bufio"
	"fmt"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/helper/input"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/helper/path"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/logic"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"sort"
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

// NewBuilder creates a new instance of BuilderYaml with the specified paths and options
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
	// Check if the directory exists
	exists, err := path.IsExist(b.buildDir)
	if err != nil {
		cobra.CheckErr(err)
	}
	if !exists {
		return fmt.Errorf("Build directory '%v' not exists\n", b.outputPath)
	}

	exists, err = path.IsExist(b.templatePath)
	if err != nil {
		cobra.CheckErr(err)
	}
	if !exists {
		return fmt.Errorf("Template file '%v 'not found\n", b.templatePath)
	}

	// Check if the services directory exists
	exists, err = path.IsExist(b.servicesDir)
	if err != nil {
		cobra.CheckErr(err)
	}
	if !exists {
		return fmt.Errorf("Services directory '%v' not exists\n", logic.ServicesDirectoryConst)
	}

	// Check if the compose file exists
	composeFileExists, err := path.IsExist(b.outputPath)
	if composeFileExists && !b.forceOverwrite {
		fmt.Printf("Compose file '%v' already exists. Overwrite[y/N]?", logic.ComposeFileNameConst)
		answer := input.AskForYesOrNot("y", "N")
		if !answer {
			return fmt.Errorf("operation canceled")
		}
	}

	// Read the template file
	templateData, err := os.ReadFile(b.templatePath)
	if err != nil {
		return fmt.Errorf("Error reading file: %v\n", err)
	}
	templateContent := string(templateData)
	//fmt.Println(templateContent)

	services, err := os.ReadDir(b.servicesDir)
	if err != nil {
		return fmt.Errorf("Error reading services directory: %v\n", err)
	}
	var serviceFiles []string
	for _, entry := range services {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yml" {
			serviceFiles = append(serviceFiles, filepath.Join(b.servicesDir, entry.Name()))
		}
	}
	sort.Strings(serviceFiles)

	var servicesContent strings.Builder
	for _, file := range serviceFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			servicesContent.WriteString(scanner.Text())
			servicesContent.WriteString("\n")
		}
		servicesContent.WriteString("\n")
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}

	if composeFileExists {
		// Create backup of existing file before overwriting
		if err := path.BackupExistingFile(b.outputPath); err != nil {
			return fmt.Errorf("Error creating backup: %v\n", err)
		}
	}

	finalContent := strings.Replace(templateContent, "<dcm: include services\\>", servicesContent.String(), 1)
	if err := os.WriteFile(b.outputPath, []byte(finalContent), 0644); err != nil {
		panic(err)
	}
	fmt.Printf("Compose file '%v' created\n", b.outputPath)

	return nil
}
