package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func getLastItemName(path string, printFiles bool) string {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for i := len(files) - 1; i >= 0; i-- {
		if printFiles {
			return files[i].Name()
		}
		if files[i].IsDir() {
			return files[i].Name()
		}
	}
	return ""
}

func getRightPrevStick(prevStick string, currentStick string) string {
	if strings.Contains(prevStick, currentStick) {
		return prevStick
	}
	old := "└───"
	if currentStick == old {
		old = "├───"
	}

	return strings.Replace(prevStick, old, currentStick, 1)
}

func setTabToStick(stick string, tab string) string {
	stickAsSlice := strings.Split(stick, "\t")
	stickAsSliceLen := len(stickAsSlice)
	betterStickAsSlice := make([]string, stickAsSliceLen+1)
	var necessarySetLast bool
	for idx, elem := range stickAsSlice {
		if necessarySetLast {
			betterStickAsSlice[len(betterStickAsSlice)-1] = elem
			break
		}
		if stickAsSliceLen-idx == 1 {
			betterStickAsSlice[idx] = tab
			betterStickAsSlice[idx+1] = elem
			necessarySetLast = true
		} else {
			betterStickAsSlice[idx] = elem
		}
	}
	return strings.Join(betterStickAsSlice, "\t")
}

func dirTreeRec(writer io.Writer, path string, printFiles bool, prevStick string) error {
	var stick string
	var tab string
	var fileName string
	var size int

	lastItemName := getLastItemName(path, printFiles)
	if lastItemName == "" {
		return nil
	}
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
		return err
	}
	for _, file := range files {
		if !file.IsDir() && !printFiles {
			continue
		}

		stick = "├───"
		fileName = file.Name()
		if lastItemName == fileName {
			stick = "└───"
		}

		if prevStick != "" {
			tab = ""
			if strings.Contains(prevStick, "├───") {
				tab = "│"
			}
			stick = getRightPrevStick(prevStick, stick)
			if strings.Contains(stick, "\t") {
				stick = setTabToStick(stick, tab)
			} else {
				tab = "\t"
				if strings.Contains(prevStick, "├───") {
					tab = "│\t"
				}
				stick = tab + stick
			}
		}

		if !file.IsDir() {
			lstat, _ := os.Lstat(path + "/" + fileName)
			size = int(lstat.Size())
			if size == 0 {
				fileName += " (empty)"
			} else {
				fileName += " (" + strconv.Itoa(size) + "b)"
			}
		}
		fmt.Fprintln(writer, stick+fileName)

		if file.IsDir() {
			dirTreeRec(writer, path+"/"+fileName, printFiles, stick)
		}
	}

	return nil
}

func dirTree(writer io.Writer, path string, printFiles bool) error {
	return dirTreeRec(writer, path, printFiles, "")
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	//err := dirTree(out, "testdata", false)
	if err != nil {
		panic(err.Error())
	}
}
