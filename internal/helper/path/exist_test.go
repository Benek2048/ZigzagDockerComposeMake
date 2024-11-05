package path

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIsExist verifies the functionality of the IsExist function by testing various scenarios
// including existing files, non-existing files, and different error conditions.
func TestIsExist(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Test cases represents different scenarios we want to verify
	tests := []struct {
		name          string // Name of the test case
		setupFunc     func() // Function to set up the test environment
		path          string // Path to check
		expectedExist bool   // Expected result of existence check
		expectedError bool   // Whether we expect an error
		cleanupFunc   func() // Function to clean up after the test
	}{
		{
			name: "existing file",
			setupFunc: func() {
				// Create a test file
				filePath := filepath.Join(tempDir, "test.txt")
				if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
					t.Fatal(err)
				}
			},
			path:          filepath.Join(tempDir, "test.txt"),
			expectedExist: true,
			expectedError: false,
			cleanupFunc:   func() {},
		},
		{
			name:          "non-existing file",
			setupFunc:     func() {},
			path:          filepath.Join(tempDir, "nonexistent.txt"),
			expectedExist: false,
			expectedError: false,
			cleanupFunc:   func() {},
		},
		{
			name: "existing directory",
			setupFunc: func() {
				// Create a test directory
				dirPath := filepath.Join(tempDir, "testdir")
				if err := os.Mkdir(dirPath, 0755); err != nil {
					t.Fatal(err)
				}
			},
			path:          filepath.Join(tempDir, "testdir"),
			expectedExist: true,
			expectedError: false,
			cleanupFunc:   func() {},
		},
		{
			name: "file with no read permissions",
			setupFunc: func() {
				// Create a file with no read permissions
				filePath := filepath.Join(tempDir, "noperm.txt")
				if err := os.WriteFile(filePath, []byte("test content"), 0000); err != nil {
					t.Fatal(err)
				}
			},
			path:          filepath.Join(tempDir, "noperm.txt"),
			expectedExist: true,
			expectedError: false,
			cleanupFunc: func() {
				// Restore permissions to allow cleanup
				err := os.Chmod(filepath.Join(tempDir, "noperm.txt"), 0644)
				if err != nil {
					return
				}
			},
		},
	}

	// Iterate through all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the test environment
			tt.setupFunc()

			// Clean up after the test
			defer tt.cleanupFunc()

			// Call the function being tested
			exists, err := IsExist(tt.path)

			// Verify error expectation
			if (err != nil) != tt.expectedError {
				t.Errorf("IsExist() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			// Verify existence expectation
			if exists != tt.expectedExist {
				t.Errorf("IsExist() = %v, expected %v", exists, tt.expectedExist)
			}
		})
	}
}
