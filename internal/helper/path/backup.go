package path

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CreateBackupFileName generates a unique backup file name with date and optional index.
// It tries base name with date first (filename-YYYYMMDD.ext), then adds incremental
// numbers (filename-YYYYMMDD.1.ext, filename-YYYYMMDD.2.ext, etc.) until it finds
// a name that doesn't exist.
func CreateBackupFileName(originalPath string) (string, error) {
	// Get current date in YYYYMMDD format
	date := time.Now().Format("20060102")

	// Create base backup name with date
	ext := filepath.Ext(originalPath)
	baseNameWithoutExt := originalPath[:len(originalPath)-len(ext)]
	baseBackupName := fmt.Sprintf("%s-%s%s", baseNameWithoutExt, date, ext)

	// Check if base backup name is available
	_, err := os.Stat(baseBackupName)
	if err != nil && os.IsNotExist(err) {
		return baseBackupName, nil
	}

	// Try with incremental numbers until we find an available name
	i := 1
	for {
		candidateName := fmt.Sprintf("%s-%s.%d%s", baseNameWithoutExt, date, i, ext)
		_, err := os.Stat(candidateName)
		if err != nil && os.IsNotExist(err) {
			return candidateName, nil
		}
		i++
	}
}

// BackupExistingFile creates a backup of the existing file with a date-based name.
// If the target backup file already exists, it will try to create a new name
// with an incremental number suffix.
func BackupExistingFile(filePath string) error {
	backupPath, err := CreateBackupFileName(filePath)
	if err != nil {
		return fmt.Errorf("failed to generate backup file name: %v", err)
	}

	err = os.Rename(filePath, backupPath)
	if err != nil {
		return fmt.Errorf("failed to rename file: %v", err)
	}

	return nil
}

// CreateBackupDirectoryName generates a unique backup directory name with date and optional index.
// It tries base name with date first (dirname-YYYYMMDD), then adds incremental
// numbers (dirname-YYYYMMDD.1, dirname-YYYYMMDD.2, etc.) until it finds
// a name that doesn't exist.
func CreateBackupDirectoryName(originalPath string) (string, error) {
	// Get current date in YYYYMMDD format
	date := time.Now().Format("20060102")

	// Create base backup name with date
	baseName := filepath.Base(originalPath)
	parentDir := filepath.Dir(originalPath)
	baseBackupName := filepath.Join(parentDir, fmt.Sprintf("%s-%s", baseName, date))

	// Check if base backup name is available
	_, err := os.Stat(baseBackupName)
	if err != nil && os.IsNotExist(err) {
		return baseBackupName, nil
	}

	// Try with incremental numbers until we find an available name
	i := 1
	for {
		candidateName := filepath.Join(parentDir, fmt.Sprintf("%s-%s.%d", baseName, date, i))
		_, err := os.Stat(candidateName)
		if err != nil && os.IsNotExist(err) {
			return candidateName, nil
		}
		i++
	}
}

// BackupExistingDirectory creates a backup of the existing directory with a date-based name.
// If the target backup directory already exists, it will try to create a new name
// with an incremental number suffix.
func BackupExistingDirectory(dirPath string) error {
	// Verify the path is a directory
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dirPath)
	}

	backupPath, err := CreateBackupDirectoryName(dirPath)
	if err != nil {
		return fmt.Errorf("failed to generate backup directory name: %v", err)
	}

	err = os.Rename(dirPath, backupPath)
	if err != nil {
		return fmt.Errorf("failed to rename directory: %v", err)
	}

	return nil
}
