// Package path /*
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
package path

import (
	"os"
)

// IsExist checks if the given path (file or directory) exists in the file system.
//
// Parameters:
//   - path: The path to the file or directory to check.
//
// Returns:
//   - bool: A boolean indicating whether the path exists (true) or not (false).
//   - error: An error that will be non-nil in case of failures.
func IsExist(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
