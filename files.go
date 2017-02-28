package main

import (
	"fmt"
	"io/ioutil"
	"os"
	fp "path/filepath"
)

type Dir struct {
	Root      string
	Dirs      []string
	Files     []string
	Images    []string
	AbsDirs   []string
	AbsFiles  []string
	AbsImages []string
}

func NewDir(path string) *Dir {
	fi, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if !fi.IsDir() {
		return nil
	} else {
		dir := Dir{Root: path}
		files, _ := ioutil.ReadDir(path)
		for _, fi := range files {
			absPath := fp.Join(path, fi.Name())
			relPath := fi.Name()
			if fi.IsDir() {
				dir.Dirs = append(dir.Dirs, relPath)
				dir.AbsDirs = append(dir.AbsDirs, absPath)
			} else {
				dir.Files = append(dir.Files, relPath)
				dir.AbsFiles = append(dir.AbsFiles, absPath)

				if fp.Base(relPath) != ".DS_Store" {
					dir.Images = append(dir.Images, relPath)
					dir.AbsImages = append(dir.AbsImages, absPath)
				}

				//switch strings.ToLower(fp.Ext(relPath)) {
				//case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
				//dir.Images = append(dir.Images, relPath)
				//dir.AbsImages = append(dir.AbsImages, absPath)
				//default:
				//}
			}
		}
		return &dir
	}
}

func size2text(size int64) string {
	const ratio = 1024
	size_float := float64(size)
	units := []string{"B", "KB", "MB", "GB", "TB", "EB"}

	index := 0
	for ; size_float > ratio; index += 1 {
		size_float /= ratio
	}
	return fmt.Sprintf("%.2f %s", size_float, units[index])
}

func get_size(path string) int64 {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	}
	return fileInfo.Size()
}

func some_files_size_int64(files []string) (total int64) {
	for _, path := range files {
		total += get_size(path)
	}
	return
}

func some_sub_dir_images_size_int64(dirs []string) (total int64) {
	for _, path := range dirs {
		tmp := NewDir(path)
		total = total + some_files_size_int64(tmp.AbsImages) + some_sub_dir_images_size_int64(tmp.AbsDirs)
	}
	return
}

func file_size_str(path string) string {
	return size2text(get_size(path))
}

func some_files_size_str(files []string) string {
	var total int64
	for _, file := range files {
		total += get_size(file)
	}
	return size2text(total)
}

func dir_images_size_str(dir string) string {
	tmp := NewDir(dir)
	return size2text(some_files_size_int64(tmp.AbsImages) + some_sub_dir_images_size_int64(tmp.AbsDirs))
}

func hasPhoto(path string) bool {
	dir := NewDir(path)
	if len(dir.Images) > 0 {
		return true
	} else {
		for _, subpath := range dir.AbsDirs {
			if hasPhoto(subpath) {
				return true
			}
		}
	}
	return false
}
