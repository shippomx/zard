package release

import (
	"os"
	"testing"
)

func TestZipTarget(t *testing.T) {
	// targetdir = "/Users/devine/code/golang/i18n/dist_json"
	// Test case: target directory does not exist
	_, err := ZipTarget("")
	if err == nil {
		t.Errorf("Expected error when target directory does not exist")
	}
	// Test case: create temporary directory for target directory
	targetDir, err := os.MkdirTemp(os.TempDir(), "test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(targetDir)
	t.Log(targetDir)
	// Test case: target directory is empty
	zipPath, err := ZipTarget(targetDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	defer os.Remove(zipPath)

	// Test case: target directory contains files
	file1, err := os.CreateTemp(targetDir, "test1.txt")
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file1.Close()

	file2, err := os.CreateTemp(targetDir, "test2.txt")
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file2.Close()

	zipPath, err = ZipTarget(targetDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	t.Log(zipPath)
	defer os.Remove(zipPath)

	// Test case: zip file is valid
	zipFile, err := os.Open(zipPath)
	if err != nil {
		t.Errorf("Failed to open zip file: %v", err)
	}
	defer zipFile.Close()
}
