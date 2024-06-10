package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"storm/data"
	"sync"
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
	var done sync.WaitGroup
	data.RunShardClient(os.Stdin, &done, MyHttpHandler)
	done.Wait()
}

func MyHttpHandler(out http.ResponseWriter, in *http.Request) {
	x := []byte{}
	if in.Body != nil {
		x, _ = ioutil.ReadAll(in.Body)
	}
	//time.Sleep(time.Duration(rand.Int()%200) * time.Millisecond)
	_, _ = io.WriteString(out, fmt.Sprintf("Shard: %s\nSelected: %s\nTime:%s\n", in.Header.Get("Shard"), in.Header.Get("Selected"), time.Now().Format(time.RFC3339Nano)))
	_, _ = io.WriteString(out, fmt.Sprintf("Path:%s\n", in.URL.String()))
	_, _ = io.WriteString(out, fmt.Sprintf("In:%s\n", string(x)))
	_, _ = io.WriteString(out, fmt.Sprintf("Out:%s\n", "Hello World!"))
	_, _ = io.WriteString(out, fmt.Sprintf("%s\n", "---"))
}
