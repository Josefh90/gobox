package gobox_utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FileNode struct {
	Name     string      `json:"name"`
	IsDir    bool        `json:"isDir"`
	Children []*FileNode `json:"children,omitempty"` // pointer slice to do add comments
}

func DirToJSON(root string) (*FileNode, error) {

	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	node := &FileNode{
		Name:  info.Name(),
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		return node, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		childPath := filepath.Join(root, entry.Name())
		childNode, err := DirToJSON(childPath)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, childNode)
	}

	out, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		fmt.Println("JSON Error:", err)
		return nil, err
	}

	fmt.Println(string(out))

	return node, nil
}
