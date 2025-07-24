package gobox_utils

import (
	"fmt"
	"os"
)

type FileNode struct {
	Name     string     `json:"name"`
	isDir    bool       `json:"isDir"`
	Children []FileNode `json:"children,omitempty"`
}

func DirToJSON(root string) (*FileNode, error) {

	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	node := &FileNode{
		Name:  info.Name(),
		isDir: info.IsDir(),
	}

	if !info.IsDir() {
		return node, nil
	}

	fmt.Println(info)
	fmt.Println(root)

	return node, nil

}
