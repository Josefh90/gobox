package gobox_utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileNode represents a single file or directory in a hierarchical structure.
// It will be serialized into JSON.
type FileNode struct {
	Name     string      `json:"name"`               // The name of the file or directory
	IsDir    bool        `json:"isDir"`              // True if this node is a directory
	Children []*FileNode `json:"children,omitempty"` // Child nodes (only if IsDir is true)
}

// DirToJSONOptions defines optional parameters for DirToJSON.
type DirToJSONOptions struct {
	PrintJSON bool // If true, print the resulting JSON to stdout
}

// DirToJSON converts a directory structure into a tree of FileNode objects.
// It recursively scans the given root directory and builds a tree representation,
// which can be serialized into JSON.
//
// Parameters:
//   - root: string — The full path of the directory or file to scan.
//   - options: optional settings (e.g., PrintJSON)
//
// Returns:
//   - *FileNode — A pointer to the root FileNode struct representing the file or folder hierarchy.
//   - error — An error if something went wrong while accessing the filesystem.
//
// Example:
//
//	Calling DirToJSON("myfolder") might output:
//	{
//	  "name": "myfolder",
//	  "isDir": true,
//	  "children": [
//	    {
//	      "name": "file1.txt",
//	      "isDir": false
//	    },
//	    {
//	      "name": "subdir",
//	      "isDir": true,
//	      "children": [
//	        {
//	          "name": "nested.txt",
//	          "isDir": false
//	        }
//	      ]
//	    }
//	  ]
//	}
func DirToJSON(root string, options *DirToJSONOptions) (*FileNode, error) {
	// Get information about the root path
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	// Create the FileNode for the current path
	node := &FileNode{
		Name:  info.Name(),
		IsDir: info.IsDir(),
	}

	// If it's not a directory, return the node as-is (leaf node)
	if !info.IsDir() {
		return node, nil
	}

	// Read all entries inside the directory
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	// Recursively convert each child entry
	for _, entry := range entries {
		childPath := filepath.Join(root, entry.Name())
		childNode, err := DirToJSON(childPath, options)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, childNode)
	}

	// Optional: print JSON representation to stdout for debugging
	// Only print JSON if requested via options
	if options != nil && options.PrintJSON {
		out, err := json.MarshalIndent(node, "", "  ")
		if err != nil {
			fmt.Println("JSON Error:", err)
			return nil, err
		}
		fmt.Println(string(out))
	}

	return node, nil
}
