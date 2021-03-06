//
// stat.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package bbos

import (
	"fmt"
	"os"
	"time"

	"github.com/markkurossi/backup/lib/tree"
	"github.com/markkurossi/blackbox-os/kernel/process"
)

type FileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	element tree.Element
}

func (info *FileInfo) Name() string {
	return info.name
}

func (info *FileInfo) Size() int64 {
	return info.size
}

func (info *FileInfo) Mode() os.FileMode {
	return info.mode
}

func (info *FileInfo) ModTime() time.Time {
	return info.modTime
}

func (info *FileInfo) IsDir() bool {
	return (info.mode & os.ModeDir) != 0
}

func (info *FileInfo) Sys() interface{} {
	return info.element
}

func Stat(p *process.Process, name string) (os.FileInfo, error) {
	path, err := p.ResolvePath(name)
	if err != nil {
		return nil, err
	}
	element, err := tree.DeserializeID(path[len(path)-1].ID, p.FS.Zone)
	if err != nil {
		return nil, err
	}
	switch el := element.(type) {
	case *tree.Directory:
		return &FileInfo{
			name:    path[len(path)-1].Name,
			mode:    os.ModeDir,
			element: element,
		}, nil

	case tree.File:
		return &FileInfo{
			name:    path[len(path)-1].Name,
			size:    el.Size(),
			element: element,
		}, nil

	default:
		return nil, fmt.Errorf("Invalid element %T", element)
	}
}

func ReadDir(p *process.Process, dirname string) ([]os.FileInfo, error) {
	info, err := Stat(p, dirname)
	if err != nil {
		return nil, err
	}
	dir, ok := info.Sys().(*tree.Directory)
	if !ok {
		return nil, fmt.Errorf("File '%s' is not a directory", dirname)
	}

	var result []os.FileInfo
	for _, entry := range dir.Entries {
		i, err := Stat(p, fmt.Sprintf("%s/%s", dirname, entry.Name))
		if err != nil {
			return nil, err
		}
		ii, ok := i.(*FileInfo)
		if ok {
			ii.modTime = time.Unix(0, entry.ModTime)
		}
		result = append(result, i)
	}

	return result, nil
}
