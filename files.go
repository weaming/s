package main

import (
	"github.com/weaming/disk-analysis/filetree"
)

func totalSizeOfNodes(nodes []*filetree.FileNode) (total int64) {
	for _, node := range nodes {
		total += node.TotalSize
	}
	return
}

func totalSizeStrOfNodes(nodes []*filetree.FileNode) string {
	size := totalSizeOfNodes(nodes)
	return filetree.HumanSize(size, 1000)
}

func hasFile(node *filetree.FileNode) bool {
	return len(node.Files) > 0
}
