package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type userStruct struct {
	Idx      int      `json:"-"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Browsers []string `json:"browsers"`
}

func (u userStruct) String() string {
	return fmt.Sprintf("[%d] %s <%s>\n", u.Idx, u.Name, u.Email)
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	/*
		!!! !!! !!!
		обратите внимание - в задании обязательно нужен отчет
		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
		так же обратите внимание на команду в параметром -http
		перечитайте еще раз задание
		!!! !!! !!!
	*/
	file, err := os.Open(filePath)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	if err != nil {
		panic(err)
	}

	seenBrowsers := make(map[string]struct{})
	uniqueBrowsers := 0
	scanner := bufio.NewScanner(file)
	counter := -1
	userPool := sync.Pool{
		New: func() interface{} {
			return &userStruct{}
		},
	}
	var strBuilder strings.Builder
	strBuilder.Grow(1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		user := userPool.Get().(*userStruct)
		err := json.Unmarshal(line, user)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false
		counter++

		for _, browser := range user.Browsers {
			isAndroidInner := strings.Contains(browser, "Android")
			isMSIEInner := strings.Contains(browser, "MSIE")

			if isAndroidInner {
				isAndroid = true
			}

			if isMSIEInner {
				isMSIE = true
			}

			if isAndroidInner || isMSIEInner {
				_, seenBrowser := seenBrowsers[browser]
				if !seenBrowser {
					seenBrowsers[browser] = struct{}{}
					uniqueBrowsers++
				}
			}

			if isAndroid && isMSIE {
				break
			}
		}

		if !(isAndroid && isMSIE) {
			user.Browsers = user.Browsers[:0]
			userPool.Put(user)
			continue
		}

		strBuilder.WriteString(
			fmt.Sprintf(
				"[%d] %s <%s>\n",
				counter,
				user.Name,
				strings.Replace(user.Email, "@", " [at] ", 1),
			),
		)

		user.Browsers = user.Browsers[:0]
		userPool.Put(user)
	}

	fmt.Fprintf(out, "found users:\n%s\n", strBuilder.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
