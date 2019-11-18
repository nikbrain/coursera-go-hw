package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

type dirTreeWriter struct {
	writer io.Writer
	tabs   []bool
}

func (dw *dirTreeWriter) SetTabs(t []bool) { dw.tabs = t }

func (dw dirTreeWriter) Write(p []byte) (n int, err error) {
	_, _ = dw.writer.Write([]byte(dw.GetTabs()))
	return dw.writer.Write(p)
}

func (dw *dirTreeWriter) WriteDir(file os.FileInfo, isLast bool) error {
	if isLast {
		_, _ = dw.Write([]byte("└───" + file.Name() + "\n"))
		dw.tabs = append(dw.tabs, false)
	} else {
		_, _ = dw.Write([]byte("├───" + file.Name() + "\n"))
		dw.tabs = append(dw.tabs, true)
	}
	return nil
}

func (dw *dirTreeWriter) WriteFile(file os.FileInfo, isLast bool) error {
	size := int(file.Size())
	sizeString := ""
	if size == 0 {
		sizeString = " (empty)"
	} else {
		sizeString = " (" + strconv.Itoa(size) + "b)"
	}
	if isLast {
		_, _ = dw.Write([]byte("└───" + file.Name() + sizeString + "\n"))
	} else {
		_, _ = dw.Write([]byte("├───" + file.Name() + sizeString + "\n"))
	}
	return nil
}

func (dw dirTreeWriter) GetTabs() string {
	prefix := ""
	for _, b := range dw.tabs {
		if b {
			prefix += "│\t"
		} else {
			prefix += "\t"
		}
	}
	return prefix
}

type ByName []os.FileInfo

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }

func dirTree(out io.Writer, path string, printFiles bool) error {
	if _, ok := out.(dirTreeWriter); !ok {
		out = dirTreeWriter{out, []bool{}}
	}

	dir, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cant open path")
	}

	var filesInto []os.FileInfo
	filesInto, err = dir.Readdir(0)
	if err != nil {
		return fmt.Errorf("cant read dir")
	}

	if !printFiles {
		var dirs []os.FileInfo
		for _, f := range filesInto {
			if f.IsDir() {
				dirs = append(dirs, f)
			}
		}
		filesInto = dirs
	}

	sort.Sort(ByName(filesInto))
	for i, file := range filesInto {
		var isLastFile bool
		if i == len(filesInto)-1 {
			isLastFile = true
		}
		if out2, ok := out.(dirTreeWriter); ok {
			if file.IsDir() {
				newPath := path + "/" + file.Name()
				_ = out2.WriteDir(file, isLastFile)

				err = dirTree(out2, newPath, printFiles)
				if err != nil {
					return fmt.Errorf("error read dir %s: %s", newPath, err.Error())
				}
			} else if printFiles {
				_ = out2.WriteFile(file, isLastFile)
			}
		}
	}

	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
