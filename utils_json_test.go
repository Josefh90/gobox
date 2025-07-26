package gobox_utils

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"bou.ke/monkey"
)

// TestDirToJSON tests the DirToJSON function with a simple directory structure.
func TestDirToJSON(t *testing.T) {
	// 🧱 1. Create a temporary directory structure
	tmpDir := t.TempDir()

	// /tmpDir/
	// ├── file1.txt
	// └── subdir/
	//     └── file2.txt

	file1 := filepath.Join(tmpDir, "file1.txt")
	subDir := filepath.Join(tmpDir, "subdir")
	file2 := filepath.Join(subDir, "file2.txt")

	err := os.WriteFile(file1, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	err = os.WriteFile(file2, []byte("world"), 0644)
	if err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	// 🧪 2. Call DirToJSON
	options := &DirToJSONOptions{
		PrintJSON: false, // important for clean test output
	}
	node, err := DirToJSON(tmpDir, options)
	if err != nil {
		t.Fatalf("DirToJSON returned error: %v", err)
	}

	// ✅ 3. Validate the root node
	if node.Name != filepath.Base(tmpDir) {
		t.Errorf("expected root name %q, got %q", filepath.Base(tmpDir), node.Name)
	}
	if !node.IsDir {
		t.Error("expected root node to be a directory")
	}
	if len(node.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(node.Children))
	}

	// 🔍 4. Check child names (order is not guaranteed, so check both ways)
	var foundFile1, foundSubdir bool
	for _, child := range node.Children {
		if child.Name == "file1.txt" && !child.IsDir {
			foundFile1 = true
		}
		if child.Name == "subdir" && child.IsDir && len(child.Children) == 1 && child.Children[0].Name == "file2.txt" {
			foundSubdir = true
		}
	}

	if !foundFile1 {
		t.Error("file1.txt not found in root children")
	}
	if !foundSubdir {
		t.Error("subdir/file2.txt not correctly found or structured")
	}
}

// TestDirToJSON_ReadDirError tests the case where os.ReadDir returns an error.
func TestDirToJSON_ReadDirError(t *testing.T) {
	// Create a temp directory (so os.Stat doesn't fail)
	tmpDir := t.TempDir()

	// Patch os.ReadDir to always return an error
	patch := monkey.Patch(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if filepath.Clean(name) == filepath.Clean(tmpDir) {
			return nil, errors.New("mocked os.ReadDir error")
		}
		// fallback to real os.ReadDir for other paths (optional)
		return os.ReadDir(name)
	})
	defer patch.Unpatch()

	options := &DirToJSONOptions{PrintJSON: false}

	_, err := DirToJSON(tmpDir, options)
	if err == nil {
		t.Fatal("expected error from mocked os.ReadDir, got nil")
	}
	if err.Error() != "mocked os.ReadDir error" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

// Example with non-existent child:
func TestDirToJSON_InvalidPath(t *testing.T) {
	options := &DirToJSONOptions{PrintJSON: false}

	// Use a path that almost certainly does not exist
	invalidPath := "/path/that/does/not/exist_1234567890"

	_, err := DirToJSON(invalidPath, options)
	if err == nil {
		t.Fatal("expected error from os.Stat for invalid path, got nil")
	}
}

// Example with broken symlink (works on Unix-like OS):
func TestDirToJSON_RecursiveError(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Create a broken symlink inside subDir
	brokenLink := filepath.Join(subDir, "broken_link")
	err := os.Symlink("/path/that/does/not/exist", brokenLink)
	if err != nil {
		t.Skip("symlink creation not supported on this OS, skipping test")
	}

	// Create a normal file to have at least one child
	file1 := filepath.Join(tmpDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}

	options := &DirToJSONOptions{PrintJSON: false}

	_, err = DirToJSON(tmpDir, options)
	if err == nil {
		t.Fatal("expected error due to broken symlink, got nil")
	}
}

func TestDirToJSON_PrintJSON(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	os.WriteFile(file1, []byte("hello"), 0644)

	options := &DirToJSONOptions{PrintJSON: true}

	// Just check no error returned and JSON is printed (you can capture stdout if needed)
	_, err := DirToJSON(tmpDir, options)
	if err != nil {
		t.Fatalf("DirToJSON returned error: %v", err)
	}
}

func BenchmarkDirToJSON(b *testing.B) {
	// Setup (not part of the benchmark)
	tmpDir := b.TempDir()

	// Create test structure:
	// tmpDir/
	// ├── file1.txt
	// └── subdir/
	//     ├── file2.txt
	//     └── nested/
	//         └── file3.txt

	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "file2.txt"), []byte("world"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir", "nested"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "nested", "file3.txt"), []byte("!"), 0644)

	options := &DirToJSONOptions{
		PrintJSON: false,
	}

	// Reset timer so setup doesn't affect benchmark
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := DirToJSON(tmpDir, options)
		if err != nil {
			b.Fatalf("DirToJSON failed: %v", err)
		}
	}
}
