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
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/logic"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds a docker-compose.yml file from separate service files",
	Long: `The build command combines separate service definitions and a template 
docker-compose file into a complete docker-compose.yml file. It looks for service 
definitions in the 'services' directory and merges them with the template file 
(docker-compose-dcm.yml) containing shared configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Read the flags
		buildDirectory, _ := cmd.Flags().GetString("directory")
		templateFileName, _ := cmd.Flags().GetString("template")
		composeFileName, _ := cmd.Flags().GetString("compose")
		forceOverwrite, _ := cmd.Flags().GetBool("force")

		// Show the parameters
		fmt.Printf("Build directory: %v\n", buildDirectory)
		fmt.Printf("Template file: %v\n", templateFileName)
		fmt.Printf("Services directory: %v\n", logic.ServicesDirectoryConst)
		fmt.Printf("Compose file: %v\n", composeFileName)
		fmt.Printf("Force overwrite: %v\n", cmd.Flags().Lookup("force").Value.String())

		// Create paths
		templateFilePath := filepath.Join(buildDirectory, templateFileName)
		serviceDirectoryPath := filepath.Join(buildDirectory, logic.ServicesDirectoryConst)
		composeFilePath := filepath.Join(buildDirectory, composeFileName)

		// Create builder with configuration
		builder := logic.NewBuilder(
			templateFilePath,     // template file path
			serviceDirectoryPath, // services directory path
			composeFilePath,      // output file path
			forceOverwrite,       // force overwrite flag
		)

		// Execute the build
		if err := builder.Build(); err != nil {
			cobra.CheckErr(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	wd, _ := os.Getwd()
	buildCmd.Flags().StringP("directory", "d", wd, "Specify the directory to build")
	buildCmd.Flags().StringP("template", "t", logic.TemplateFileNameDefaultConst, "Specify the template file to build")
	buildCmd.Flags().StringP("compose", "c", logic.ComposeFileNameConst, "Specify the compose file to build")
	buildCmd.Flags().BoolP("force", "f", false, "Force overwrite of existing compose file or services folder")
}
