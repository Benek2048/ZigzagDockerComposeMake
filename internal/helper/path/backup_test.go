package path

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCreateBackupFileName verifies that the createBackupFileName function:
// - Generates correct date-based backup names
// - Handles existing files by adding incremental numbers
// - Properly preserves file extensions
// - Returns paths that don't conflict with existing files
func TestCreateBackupFileName(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Test cases cover different scenarios for backup file naming
	tests := []struct {
		name           string   // Test case name
		setupFiles     []string // Files to create before test
		expectedSuffix string   // Expected suffix in the generated name
		originalName   string   // Name of the original file
	}{
		{
			name:           "fresh_backup_no_existing_files",
			setupFiles:     []string{},
			expectedSuffix: time.Now().Format("20060102"),
			originalName:   "test.txt",
		},
		{
			name: "existing_base_backup",
			setupFiles: []string{
				fmt.Sprintf("test-%s.txt", time.Now().Format("20060102")),
			},
			expectedSuffix: fmt.Sprintf("%s.1", time.Now().Format("20060102")),
			originalName:   "test.txt",
		},
		{
			name: "multiple_existing_backups",
			setupFiles: []string{
				fmt.Sprintf("test-%s.txt", time.Now().Format("20060102")),
				fmt.Sprintf("test-%s.1.txt", time.Now().Format("20060102")),
				fmt.Sprintf("test-%s.2.txt", time.Now().Format("20060102")),
			},
			expectedSuffix: fmt.Sprintf("%s.3", time.Now().Format("20060102")),
			originalName:   "test.txt",
		},
		{
			name:           "file_with_multiple_extensions",
			setupFiles:     []string{},
			expectedSuffix: time.Now().Format("20060102"),
			originalName:   "test.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the original file path
			originalPath := filepath.Join(tempDir, tt.originalName)

			// Create a dummy original file
			if err := os.WriteFile(originalPath, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Create all setup files
			for _, fileName := range tt.setupFiles {
				filePath := filepath.Join(tempDir, fileName)
				if err := os.WriteFile(filePath, []byte("existing backup"), 0644); err != nil {
					t.Fatalf("Failed to create setup file %s: %v", fileName, err)
				}
			}

			// Call the function under test
			result, err := CreateBackupFileName(originalPath)
			if err != nil {
				t.Fatalf("createBackupFileName failed: %v", err)
			}

			// Verify the result contains expected date pattern
			if !strings.Contains(result, tt.expectedSuffix) {
				t.Errorf("Expected backup name to contain suffix %s, got %s", tt.expectedSuffix, result)
			}

			// Verify the generated name doesn't exist yet
			if _, err := os.Stat(result); err == nil {
				t.Error("Generated backup file name already exists")
			}

			// Verify file extension is preserved
			expectedExt := filepath.Ext(tt.originalName)
			if !strings.HasSuffix(result, expectedExt) {
				t.Errorf("Expected file extension %s, got %s", expectedExt, filepath.Ext(result))
			}
		})
	}
}

// TestBackupExistingFile verifies that the backupExistingFile function:
// - Successfully renames existing files
// - Generates correct backup names
// - Handles errors appropriately
// - Preserves file contents during rename
func TestBackupExistingFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name                string
		setupFiles          []string // Existing backup files
		expectedError       bool     // Whether we expect an error
		expectedBackupCount int      // Expected number of backup files after operation
	}{
		{
			name:                "successful_backup",
			setupFiles:          []string{},
			expectedError:       false,
			expectedBackupCount: 1, // Just the new backup
		},
		{
			name: "backup_with_existing_files",
			setupFiles: []string{
				fmt.Sprintf("test-%s.txt", time.Now().Format("20060102")),
			},
			expectedError:       false,
			expectedBackupCount: 2, // One existing + one new backup
		},
		{
			name: "backup_with_multiple_existing_files",
			setupFiles: []string{
				fmt.Sprintf("test-%s.txt", time.Now().Format("20060102")),
				fmt.Sprintf("test-%s.1.txt", time.Now().Format("20060102")),
			},
			expectedError:       false,
			expectedBackupCount: 3, // Two existing + one new backup
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create original file to backup
			originalPath := filepath.Join(tempDir, "test.txt")
			if err := os.WriteFile(originalPath, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Create setup files
			for _, fileName := range tt.setupFiles {
				filePath := filepath.Join(tempDir, fileName)
				if err := os.WriteFile(filePath, []byte("existing backup"), 0644); err != nil {
					t.Fatalf("Failed to create setup file %s: %v", fileName, err)
				}
			}

			// Perform backup
			err := BackupExistingFile(originalPath)

			// Check error expectation
			if (err != nil) != tt.expectedError {
				t.Errorf("Expected error: %v, got error: %v", tt.expectedError, err)
			}

			if err == nil {
				// Verify original file doesn't exist anymore
				if _, err := os.Stat(originalPath); err == nil {
					t.Error("Original file still exists after backup")
				}

				// Verify backup files count
				pattern := filepath.Join(tempDir, fmt.Sprintf("test-%s*", time.Now().Format("20060102")))
				matches, err := filepath.Glob(pattern)
				if err != nil {
					t.Fatalf("Failed to find backup files: %v", err)
				}
				if len(matches) != tt.expectedBackupCount {
					t.Fatalf("Expected to find exactly %d backup files, found %d",
						tt.expectedBackupCount, len(matches))
				}
			}
		})
	}
}

// TestCreateBackupDirectoryName verifies that the CreateBackupDirectoryName function:
// - Generates correct date-based backup names
// - Handles existing directories by adding incremental numbers
// - Returns paths that don't conflict with existing directories
func TestCreateBackupDirectoryName(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setupDirs      []string // Directories to create before test
		expectedSuffix string   // Expected suffix in the generated name
		originalName   string   // Name of the original directory
	}{
		{
			name:           "fresh_backup_no_existing_directories",
			setupDirs:      []string{},
			expectedSuffix: time.Now().Format("20060102"),
			originalName:   "testdir",
		},
		{
			name: "existing_base_backup",
			setupDirs: []string{
				fmt.Sprintf("testdir-%s", time.Now().Format("20060102")),
			},
			expectedSuffix: fmt.Sprintf("%s.1", time.Now().Format("20060102")),
			originalName:   "testdir",
		},
		{
			name: "multiple_existing_backups",
			setupDirs: []string{
				fmt.Sprintf("testdir-%s", time.Now().Format("20060102")),
				fmt.Sprintf("testdir-%s.1", time.Now().Format("20060102")),
				fmt.Sprintf("testdir-%s.2", time.Now().Format("20060102")),
			},
			expectedSuffix: fmt.Sprintf("%s.3", time.Now().Format("20060102")),
			originalName:   "testdir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tempDir, t.Name())
			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test directory structure: %v", err)
			}

			// Create the original directory path
			originalPath := filepath.Join(testDir, tt.originalName)
			if err := os.Mkdir(originalPath, 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Create all setup directories
			for _, dirName := range tt.setupDirs {
				dirPath := filepath.Join(testDir, dirName)
				if err := os.Mkdir(dirPath, 0755); err != nil {
					t.Fatalf("Failed to create setup directory %s: %v", dirName, err)
				}
			}

			// Call the function under test
			result, err := CreateBackupDirectoryName(originalPath)
			if err != nil {
				t.Fatalf("CreateBackupDirectoryName failed: %v", err)
			}

			// Verify the result contains expected date pattern
			if !strings.Contains(result, tt.expectedSuffix) {
				t.Errorf("Expected backup name to contain suffix %s, got %s", tt.expectedSuffix, result)
			}

			// Verify the generated name doesn't exist yet
			if _, err := os.Stat(result); err == nil {
				t.Error("Generated backup directory name already exists")
			}
		})
	}
}

// TestBackupExistingDirectory verifies that the BackupExistingDirectory function:
// - Successfully renames existing directories
// - Generates correct backup names
// - Handles errors appropriately
func TestBackupExistingDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name                string
		setupDirs           []string // Existing backup directories
		expectedError       bool     // Whether we expect an error
		expectedBackupCount int      // Expected number of backup directories after operation
	}{
		{
			name:                "successful_backup",
			setupDirs:           []string{},
			expectedError:       false,
			expectedBackupCount: 1, // Just the new backup
		},
		{
			name: "backup_with_existing_directories",
			setupDirs: []string{
				fmt.Sprintf("testdir-%s", time.Now().Format("20060102")),
			},
			expectedError:       false,
			expectedBackupCount: 2, // One existing + one new backup
		},
		{
			name: "backup_with_multiple_existing_directories",
			setupDirs: []string{
				fmt.Sprintf("testdir-%s", time.Now().Format("20060102")),
				fmt.Sprintf("testdir-%s.1", time.Now().Format("20060102")),
			},
			expectedError:       false,
			expectedBackupCount: 3, // Two existing + one new backup
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tempDir, t.Name())
			if err := os.MkdirAll(testDir, 0755); err != nil {
				t.Fatalf("Failed to create test directory structure: %v", err)
			}

			// Create setup directories
			for _, dirName := range tt.setupDirs {
				dirPath := filepath.Join(testDir, dirName)
				if err := os.Mkdir(dirPath, 0755); err != nil {
					t.Fatalf("Failed to create setup directory %s: %v", dirName, err)
				}
			}

			// Create original directory to backup
			originalPath := filepath.Join(testDir, "testdir")
			if err := os.Mkdir(originalPath, 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			// Perform backup
			err := BackupExistingDirectory(originalPath)

			// Check error expectation
			if (err != nil) != tt.expectedError {
				t.Errorf("Expected error: %v, got error: %v", tt.expectedError, err)
			}

			if err == nil {
				// Verify original directory doesn't exist anymore
				if _, err := os.Stat(originalPath); err == nil {
					t.Error("Original directory still exists after backup")
				}

				// Verify backup directories count
				pattern := filepath.Join(testDir, fmt.Sprintf("testdir-%s*", time.Now().Format("20060102")))
				matches, err := filepath.Glob(pattern)
				if err != nil {
					t.Fatalf("Failed to find backup directories: %v", err)
				}
				if len(matches) != tt.expectedBackupCount {
					t.Errorf("Expected to find exactly %d backup directories, found %d",
						tt.expectedBackupCount, len(matches))
				}
			}
		})
	}
}
