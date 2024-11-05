// Package logic /*
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
package logic

var (
	// VersionConst will be set during build
	VersionConst = "development"
	// BuildTime will be set during build
	BuildTime = "unknown"
	// GitCommit will be set during build
	GitCommit = "unknown"
)

const (
	// RepositoryURLConst is the URL of the repository
	RepositoryURLConst = "https://github.com/Benek2048/ZigzagDockerComposeMake"

	// TemplateFileNameDefaultConst is the default template filename
	TemplateFileNameDefaultConst = "docker-compose-dcm.yml"

	// ServicesDirectoryConst is the directory containing service definitions
	ServicesDirectoryConst = "services"

	// ComposeFileNameConst is the default output compose filename
	ComposeFileNameConst = "docker-compose.yml"

	// BuildDirectoryConst is the default build directory
	BuildDirectoryConst = "."
)
