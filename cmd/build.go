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
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	buildDirectoryDefaultConst   = "."
	templateFileNameDefaultConst = "docker-compose-dcm.yml"
	servicesDirectoryConst       = "services"
	composeFileNameConst         = "docker-compose.yml"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("build called")

		//Read the flags
		templateDirectory, _ := cmd.Flags().GetString("directory")
		templateFileName, _ := cmd.Flags().GetString("template")
		composeFileName, _ := cmd.Flags().GetString("compose")
		forceOverwrite, _ := cmd.Flags().GetBool("force")

		//Show the parameters
		fmt.Printf("Buuild directory: %v\n", templateDirectory)
		fmt.Printf("Template file: %v\n", templateFileName)
		fmt.Printf("Services directory: %v\n", servicesDirectoryConst)
		fmt.Printf("Compose file: %v\n", composeFileName)
		fmt.Printf("Force overwrite: %v\n", cmd.Flags().Lookup("force").Value.String())
		templateFilePath := filepath.Join(templateDirectory, templateFileName)
		serviceDirectoryPath := filepath.Join(templateDirectory, servicesDirectoryConst)
		composeFilePath := filepath.Join(templateDirectory, composeFileName)
		// Check if the directory exists
		_, err := os.Stat(templateDirectory)
		if err != nil {
			fmt.Printf("Build directory '%v' not exists\n", templateDirectory)
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
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("Compose file '%v' already exists. Overwrite [y/N]?", composeFileNameConst)
			answer, _ := reader.ReadString('\n')
			answer = strings.ToLower(strings.TrimSpace(answer))
			if answer != "y" {
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	buildCmd.Flags().StringP("directory", "d", buildDirectoryDefaultConst, "Specify the directory to build")
	buildCmd.Flags().StringP("template", "t", templateFileNameDefaultConst, "Specify the template file to build")
	buildCmd.Flags().StringP("compose", "c", composeFileNameConst, "Specify the compose file to build")
	buildCmd.Flags().BoolP("force", "f", false, "Force overwrite the compose file")
}
