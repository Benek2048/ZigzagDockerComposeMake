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
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info.",
	Long:  `Show version information of the program.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s\n", versionConst)
		fmt.Printf("ZigzagDockerComposeMake\n")
		fmt.Printf("https://github.com/Benek2048/ZigzagDockerComposeMake\n")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
