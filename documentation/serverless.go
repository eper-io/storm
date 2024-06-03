package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"storm/data"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Example request:
// curl -X 'PUT' -d 'abcdef' 'http://127.0.0.1:7777/portal/abcd'
// Example request:
// curl -X 'GET' 'http://127.0.0.1:7777/portal/abcd'

func main() {
	api := "https://hour.schmied.us/df94d5658feda65e9d5cdac6bcd50b8012c835ab884f0a74c5fa46e396b05ae7.tig"
	snapshot := data.RunShardListHttp(api, MyHttpHandler)
	for {
		// Infinite loop
		time.Sleep(10 * time.Second)
		current := data.TmpGet(api)
		// Check for modifications
		if !bytes.Equal(current, snapshot) {
			return
		}
	}
}

func MyHttpHandler(out http.ResponseWriter, in *http.Request) {
	x := []byte{}
	if in.Body != nil {
		x, _ = ioutil.ReadAll(in.Body)
	}
	_, _ = io.WriteString(out, fmt.Sprintf("In:%s\n", string(x)))
	_, _ = io.WriteString(out, fmt.Sprintf("Out:%s\n", "Hello World!"))
}
