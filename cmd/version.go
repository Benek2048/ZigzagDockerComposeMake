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
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info.",
	Long:  `Show version information of the program and compilation.`,
	Run: func(cmd *cobra.Command, args []string) {
		const colWidth = 15
		fmt.Printf("%-*s%s\n", colWidth, "Project:", "ZigzagDockerComposeMake")
		fmt.Printf("%-*s%s\n", colWidth, "Version:", logic.VersionConst)
		fmt.Printf("%-*s%s (UTC)\n", colWidth, "Build Time:", logic.BuildTime)
		fmt.Printf("%-*s%s\n", colWidth, "Git Commit:", logic.GitCommit)
		fmt.Printf("%-*s%s\n", colWidth, "Repository:", logic.RepositoryURLConst)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
