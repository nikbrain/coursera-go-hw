package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const BrowserAndroid = "Android"
const BrowserMSIE = "MSIE"

type User struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Browsers []string `json:"browsers"`
}

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	defer file.Close()

	r := bufio.NewScanner(file)

	user := User{}
	seenBrowsers := make(map[string]bool, 200)

	fmt.Fprintln(out, "found users:")

	for i := 0; r.Scan(); i++ {

		line := r.Bytes()

		if !(bytes.Contains(line, []byte(BrowserAndroid)) || bytes.Contains(line, []byte(BrowserMSIE))) {
			continue
		}

		err = json.Unmarshal(line, &user)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {

			switch {
			case strings.Contains(browser, BrowserAndroid):
				isAndroid = true
			case strings.Contains(browser, BrowserMSIE):
				isMSIE = true
			default:
				continue
			}

			seenBrowsers[browser] = true
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email := strings.Replace(user.Email, "@", " [at] ", -1)
		fmt.Fprintln(out, "["+strconv.Itoa(i)+"] "+user.Name+" <"+email+">")
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}
