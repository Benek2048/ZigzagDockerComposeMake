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
	"github.com/Benek2048/ZigzagDockerComposeMake/internal/helper/path"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Service map[string]interface{}

type DockerCompose struct {
	Services map[string]Service `yaml:"services"`
}

// decomposeCmd represents the decompose command
var decomposeCmd = &cobra.Command{
	Use:   "decompose",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("decompose called")

		//Read the flags
		buildDirectory, _ := cmd.Flags().GetString("directory")
		templateFileName, _ := cmd.Flags().GetString("template")
		composeFileName, _ := cmd.Flags().GetString("compose")
		forceOverwrite, _ := cmd.Flags().GetBool("force")

		//Show the parameters
		fmt.Printf("Build directory: %v\n", buildDirectory)
		fmt.Printf("Template file: %v\n", templateFileName)
		fmt.Printf("Services directory: %v\n", servicesDirectoryConst)
		fmt.Printf("Compose file: %v\n", composeFileName)
		fmt.Printf("Force overwrite: %v\n", cmd.Flags().Lookup("force").Value.String())
		templateFilePath := filepath.Join(buildDirectory, templateFileName)
		serviceDirectoryPath := filepath.Join(buildDirectory, servicesDirectoryConst)
		composeFilePath := filepath.Join(buildDirectory, composeFileName)

		exists, err := path.IsExist(composeFilePath)
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
			fmt.Printf("Compose file '%v' already exists. Overwrite[y/N]?", composeFileName)
			answer := input.AskForYesOrNot("y", "N")
			if !answer {
				fmt.Println("Operation canceled")
				return
			}
		}

		exists, err = path.IsExist(buildDirectory)
		if err != nil {
			fmt.Printf("Build directory '%v' not exists\n", buildDirectory)
			return
		}
		file, err := os.Open("docker-compose.yml")
		if err != nil {
			cobra.CheckErr(err)
		}
		defer func(file *os.File) {
			err := file.Close()
			cobra.CheckErr(err)
		}(file)

		err = os.Mkdir(serviceDirectoryPath, 0755)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}

		noComments, _ := cmd.Flags().GetBool("format")
		splitWorkByFlags(file, noComments)
	},
}

// splitWorkByFlags function which splits work based on flags
// splitWorkByFlags is a function that takes a file pointer and a boolean flag as input. It splits the work of decomposing a file based on the value of the flag. If the flag is true, it calls the DecomposeWithoutComments function. If the flag is false, it calls the DecomposeWithComments function.
func splitWorkByFlags(file *os.File, noComments bool) {
	if noComments {
		decomposeWithoutComments(file)
	} else {
		decomposeWithComments(file)
	}
}

func decomposeWithoutComments(file *os.File) {
	var compose DockerCompose
	decoder := yaml.NewDecoder(file)
	err := decoder.Decode(&compose)
	if err != nil {
		panic(err)
	}

	for name, service := range compose.Services {
		yamlData, err := marshalServiceData(name, service)
		if err != nil {
			panic(err)
		}

		err = writeServiceToFile(name, yamlData)
		if err != nil {
			panic(err)
		}
	}
}

func marshalServiceData(name string, service Service) ([]byte, error) {
	return yaml.Marshal(map[string]Service{name: service})
}

func writeServiceToFile(name string, data []byte) error {
	serviceFile, err := os.Create(filepath.Join("services", name+".yml"))
	if err != nil {
		return err
	}
	defer func(serviceFile *os.File) {
		err := serviceFile.Close()
		cobra.CheckErr(err)
	}(serviceFile)

	_, err = serviceFile.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func decomposeWithComments(file *os.File) {
	scanner := bufio.NewScanner(file)

	re := regexp.MustCompile(`(?s)^\s{2}(\w+):\s*$`)
	startRe := regexp.MustCompile(`(?s)^services:\s*$`)
	endRe := regexp.MustCompile(`(?s)^\w+:\s*$`)
	startWriting := false

	var err error
	var currentFile *os.File
	var builder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)

		if startRe.MatchString(line) {
			startWriting = true
			continue
		}
		if endRe.MatchString(line) {
			startWriting = false
			if currentFile != nil {
				_, err = currentFile.WriteString(builder.String())
				if err != nil {
					fmt.Println(err)
					return
				}
				err := currentFile.Close()
				if err != nil {
					return
				}
				//currentFile = nil
				builder.Reset()
			}
			continue
		}

		if startWriting && match != nil {
			if currentFile != nil {
				_, err = currentFile.WriteString(builder.String())
				if err != nil {
					fmt.Println(err)
					return
				}
				err = currentFile.Close()
				if err != nil {
					return
				}
				//currentFile = nil
				builder.Reset()
			}
			newFile := filepath.Join("services", match[1]+".yml")
			currentFile, err = os.Create(newFile)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Creating file: ", newFile)
		}

		if startWriting && currentFile != nil {
			builder.WriteString(line + "\n")
		}
	}

	if currentFile != nil {
		if builder.String() == "" {
			fmt.Printf("WARNING: '%v' file is already closed and stringBuilder is empty\n", currentFile.Name())
		} else {
			_, err = currentFile.WriteString(builder.String())
			if err != nil {
				fmt.Println(err)
				return
			}
			err = currentFile.Close()
			cobra.CheckErr(err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

}

func init() {
	rootCmd.AddCommand(decomposeCmd)

	wd, _ := os.Getwd()
	decomposeCmd.Flags().StringP("directory", "d", wd, "Specify the directory to build")
	decomposeCmd.Flags().StringP("template", "t", templateFileNameDefaultConst, "Specify the template file to build")
	decomposeCmd.Flags().StringP("compose", "c", composeFileNameConst, "Specify the compose file to build")
	decomposeCmd.Flags().BoolP("force", "f", false, "Force overwrite")

	decomposeCmd.Flags().BoolP("format", "", false, "format and delete comments")
}
