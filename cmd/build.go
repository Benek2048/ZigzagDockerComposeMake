// Package cmd /*
/*
Copyright © 2024 Benek <benek2048@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"fmt"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/helper/input"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Building the docker-compose.yml file",
	Long: `Creating the docker-compose.yml file based on:
'docker-compose-dcm.yml' - the main template into which services from the 'services' folder will be inserted
'./services' - folder with service files. Each file is one service.
'docker-compose.yml' - resulting file`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("build called")

		//Read the flags
		buildDirectory, _ := cmd.Flags().GetString("directory")
		templateFileName, _ := cmd.Flags().GetString("template")
		composeFileName, _ := cmd.Flags().GetString("compose")
		forceOverwrite, _ := cmd.Flags().GetBool("force")

		//Show the parameters
		fmt.Printf("Buuild directory: %v\n", buildDirectory)
		fmt.Printf("Template file: %v\n", templateFileName)
		fmt.Printf("Services directory: %v\n", servicesDirectoryConst)
		fmt.Printf("Compose file: %v\n", composeFileName)
		fmt.Printf("Force overwrite: %v\n", cmd.Flags().Lookup("force").Value.String())
		templateFilePath := filepath.Join(buildDirectory, templateFileName)
		serviceDirectoryPath := filepath.Join(buildDirectory, servicesDirectoryConst)
		composeFilePath := filepath.Join(buildDirectory, composeFileName)
		// Check if the directory exists
		_, err := os.Stat(buildDirectory)
		if err != nil {
			fmt.Printf("Build directory '%v' not exists\n", buildDirectory)
			return
		}
		// Check if the template file exists
		_, err = os.Stat(templateFilePath)
		if err != nil {
			fmt.Printf("Template file '%v 'not found\n", templateFileName)
			return
		}
		// Check if the services directory exists
		_, err = os.Stat(serviceDirectoryPath)
		if err != nil {
			fmt.Printf("Services directory '%v' not exists\n", servicesDirectoryConst)
			return
		}
		// Check if the compose file exists
		_, err = os.Stat(composeFilePath)
		if !forceOverwrite && err == nil {
			fmt.Printf("Compose file '%v' already exists. Overwrite[y/N]?", composeFileNameConst)
			answer := input.AskForYesOrNot("y", "N")
			if !answer {
				fmt.Println("Operation canceled")
				return
			}
		}

		// Read the template file
		templateData, err := os.ReadFile(templateFilePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}
		templateContent := string(templateData)
		//fmt.Println(templateContent)

		services, err := os.ReadDir(serviceDirectoryPath)
		if err != nil {
			fmt.Printf("Error reading services directory: %v\n", err)
			return
		}
		var serviceFiles []string
		for _, entry := range services {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".yml" {
				serviceFiles = append(serviceFiles, filepath.Join(serviceDirectoryPath, entry.Name()))
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
				servicesContent.WriteString("  ") // indent
				servicesContent.WriteString(scanner.Text())
				servicesContent.WriteString("\n")
			}
			servicesContent.WriteString("\n")
			if err := scanner.Err(); err != nil {
				panic(err)
			}
		}

		finalContent := strings.Replace(templateContent, "<dcm: include services\\>", servicesContent.String(), 1)
		if err := os.WriteFile(composeFilePath, []byte(finalContent), 0644); err != nil {
			panic(err)
		}
		fmt.Printf("Compose file '%v' created\n", composeFileName)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	wd, _ := os.Getwd()
	buildCmd.Flags().StringP("directory", "d", wd, "Specify the directory to build")
	buildCmd.Flags().StringP("template", "t", templateFileNameDefaultConst, "Specify the template file to build")
	buildCmd.Flags().StringP("compose", "c", composeFileNameConst, "Specify the compose file to build")
	buildCmd.Flags().BoolP("force", "f", false, "Force overwrite")
}
