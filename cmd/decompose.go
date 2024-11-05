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
	"fmt"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/helper/input"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/helper/path"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/logic"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/logic/text"
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/logic/yaml"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

//type Service map[string]interface{}
//
//type DockerCompose struct {
//	Services map[string]Service `yaml:"services"`
//}

// decomposeCmd represents the decompose command
var decomposeCmd = &cobra.Command{
	Use:   "decompose",
	Short: "Breaking down the 'docker-compose.yml' file into individual service files",
	Long:  `Based on the 'docker-compose.yml' file, service files will be created in the 'services' folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("decompose called")

		//Read the flags
		buildDirectory, _ := cmd.Flags().GetString("directory")
		templateFileName, _ := cmd.Flags().GetString("template")
		composeFileName, _ := cmd.Flags().GetString("compose")
		forceOverwrite, _ := cmd.Flags().GetBool("force")
		yamlMode, _ := cmd.Flags().GetBool("yaml-mode")

		//Show the parameters
		fmt.Printf("Build directory: %v\n", buildDirectory)
		fmt.Printf("Template file: %v\n", templateFileName)
		fmt.Printf("Services directory: %v\n", logic.ServicesDirectoryConst)
		fmt.Printf("Compose file: %v\n", composeFileName)
		fmt.Printf("Force overwrite: %v\n", cmd.Flags().Lookup("force").Value.String())
		templateFilePath := filepath.Join(buildDirectory, templateFileName)
		serviceDirectoryPath := filepath.Join(buildDirectory, logic.ServicesDirectoryConst)
		composeFilePath := filepath.Join(buildDirectory, composeFileName)

		exists, err := path.IsExist(buildDirectory)
		if err != nil {
			cobra.CheckErr(err)
		}
		if !exists {
			fmt.Printf("Build directory '%v' not exists\n", buildDirectory)
			return
		}

		exists, err = path.IsExist(composeFilePath)
		if err != nil {
			cobra.CheckErr(err)
		}
		if !exists {
			fmt.Printf("Compose file '%v' not exists\n", composeFileName)
			return
		}

		exists, err = path.IsExist(templateFilePath)
		if err != nil {
			cobra.CheckErr(err)
		}
		if exists && !forceOverwrite {
			fmt.Printf("Template file '%v' already exists. Overwrite[y/N]?", templateFilePath)
			answer := input.AskForYesOrNot("y", "N")
			if !answer {
				fmt.Println("Operation canceled")
				return
			}
			// Create backup of existing file before overwriting
			if err := path.BackupExistingFile(templateFilePath); err != nil {
				fmt.Printf("Error creating backup: %v\n", err)
				return
			}
		}

		exists, err = path.IsExist(serviceDirectoryPath)
		if err != nil {
			cobra.CheckErr(err)
		}
		if exists && !forceOverwrite {
			fmt.Printf("Service directory '%v' already exists. Overwrite[y/N]?", serviceDirectoryPath)
			answer := input.AskForYesOrNot("y", "N")
			if !answer {
				fmt.Println("Operation canceled")
				return
			}
			// Create backup of existing directory before overwriting
			if err := path.BackupExistingDirectory(serviceDirectoryPath); err != nil {
				fmt.Printf("Error creating backup: %v\n", err)
				return
			}
		}

		if !yamlMode {
			fmt.Println("Text mode")
			decomposer := text.NewServiceDecomposer(
				composeFilePath,      // fileSrc
				templateFilePath,     // fileTemplate
				serviceDirectoryPath, // servicesDir
			)
			if err := decomposer.Decompose(); err != nil {
				cobra.CheckErr(err)
			}
		} else {
			fmt.Println("Yaml mode")
			decomposer := yaml.NewServiceDecomposer(
				composeFilePath,      // fileSrc
				templateFilePath,     // fileTemplate
				serviceDirectoryPath, // servicesDir
			)
			if err := decomposer.Decompose(); err != nil {
				cobra.CheckErr(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(decomposeCmd)

	wd, _ := os.Getwd()
	decomposeCmd.Flags().StringP("directory", "d", wd, "Specify the directory to build")
	decomposeCmd.Flags().StringP("template", "t", logic.TemplateFileNameDefaultConst, "Specify the template file to build")
	decomposeCmd.Flags().StringP("compose", "c", logic.ComposeFileNameConst, "Specify the compose file to build")
	decomposeCmd.Flags().BoolP("force", "f", false, "Force overwrite")
	decomposeCmd.Flags().BoolP("yaml-mode", "", false, "Use YAML mode for processing")
}
