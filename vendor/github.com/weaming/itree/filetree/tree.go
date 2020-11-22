package filetree

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	TYPE_DIR  = "dir"
	TYPE_FILE = "file"
)

type FileNode struct {
	Name      string
	Extension string
	AbsPath   string
	RelPath   string
	Size      int64
	TotalSize int64
	HumanSize string
	Type      string

	Parent *FileNode
	Dirs   []*FileNode
	Files  []*FileNode

	Images   []*FileNode
	Children []*FileNode
}

func NewFileNode(path string, root string, parent *FileNode, sortBySize bool) *FileNode {
	// parse arguments: absolute pathes
	path, e := filepath.Abs(path)
	fatalErr(e)
	root, e = filepath.Abs(root)
	fatalErr(e)

	fi, err := os.Stat(path)
	if !IsFileOrDir(fi) {
		return nil
	}
	fatalErr(err)

	// get abs/rel path
	absPath, err := filepath.Abs(path)
	fatalErr(err)
	relPath, err := filepath.Rel(root, path)
	fatalErr(err)

	rv := &FileNode{
		Name:      fi.Name(),
		Extension: filepath.Ext(fi.Name()),
		AbsPath:   absPath,
		RelPath:   relPath,
		Size:      fi.Size(),
		Parent:    parent,
		Type:      TYPE_FILE,

		Dirs:     []*FileNode{},
		Files:    []*FileNode{},
		Images:   []*FileNode{},
		Children: []*FileNode{},
	}

	if fi.IsDir() {
		rv.Type = TYPE_DIR
		// println(rv.Size) // dir size is not 0 !!
		rv.Size = 0

		files, err := ioutil.ReadDir(path)
		fatalErr(err)

		for _, fi := range files {
			absPath, err := filepath.Abs(filepath.Join(path, fi.Name()))
			fatalErr(err)

			childFile := NewFileNode(absPath, root, rv, sortBySize)
			if childFile == nil {
				continue
			}

			if fi.IsDir() {
				// children
				rv.Children = append(rv.Children, childFile)
				rv.Dirs = append(rv.Dirs, childFile)
			} else {
				// children
				rv.Children = append(rv.Children, childFile)
				rv.Files = append(rv.Files, childFile)

				switch strings.ToLower(childFile.Extension) {
				case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
					rv.Images = append(rv.Images, childFile)
				default:
				}
			}
		}
	}

	rv.TotalSize = rv.totalSize()
	if sortBySize {
		rv.Sort()
	}
	return rv
}

func (p *FileNode) Sort() {
	sort.Stable(sort.Reverse(p))
}

func (p *FileNode) totalSize() int64 {
	rv := p.Size
	if p.Type == TYPE_FILE {
		return rv
	} else {
		for _, x := range p.Files {
			rv += x.Size
		}

		for _, x := range p.Dirs {
			rv += x.totalSize()
		}
		return rv
	}
}

// Len is part of sort.Interface.
func (p *FileNode) Len() int {
	return len(p.Children)
}

// Swap is part of sort.Interface.
func (p *FileNode) Swap(i, j int) {
	p.Children[i], p.Children[j] = p.Children[j], p.Children[i]
}

// Less is part of sort.Interface.
func (p *FileNode) Less(i, j int) bool {
	return p.Children[i].TotalSize < p.Children[j].TotalSize
}

func fatalErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func IsFileOrDir(fi os.FileInfo) bool {
	if fi == nil {
		return false
	}
	return fi.Mode().IsDir() || fi.Mode().IsRegular()
}
