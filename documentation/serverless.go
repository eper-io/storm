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
	data.RunShardList(api, RunServerlessLambdaBurstOnHttp)
	for {
		// Infinite loop
		// TODO check for modifications
		time.Sleep(10 * time.Second)
	}
}

type serverlessHttpWriter struct {
	header     http.Header
	out        bytes.Buffer
	statusCode int
}

func (s serverlessHttpWriter) Header() http.Header {
	return s.header
}

func (s *serverlessHttpWriter) Write(x []byte) (int, error) {
	return s.out.Write(x)
}

func (s *serverlessHttpWriter) WriteHeader(statusCode int) {
	s.statusCode = statusCode
}

func RunServerlessLambdaBurstOnHttp(out *bytes.Buffer, in []byte, shard int) {
	x := bytes.SplitN(in, []byte{'\n'}, 4)
	if len(x) != 4 {
		// Fallback path
		(*out).WriteString(fmt.Sprintf("Shard: %d\nTime:%s\nIn:\n%s\nOut:\n%s\n", shard, time.Now().Format(time.RFC3339Nano), string(in), "Hello World!"))
		return
	}
	m := string(x[1])
	u := string(x[2])
	request := bytes.NewBuffer(x[3])
	fmt.Println(u, m, request.Len())
	req, _ := http.NewRequest(m, u, request)
	//if req == nil {
	//	req, _ = http.NewRequest(m, u, nil)
	//}
	var z serverlessHttpWriter
	MyHttpHandler(&z, req)
	out.WriteString(fmt.Sprintf("Shard: %d\nTime:%s\n%s", shard, time.Now().Format(time.RFC3339Nano), string(z.out.Bytes())))
}

func MyHttpHandler(out http.ResponseWriter, in *http.Request) {
	x := []byte{}
	if in.Body != nil {
		x, _ = ioutil.ReadAll(in.Body)
	}
	_, _ = io.WriteString(out, fmt.Sprintf("In:%s\n", string(x)))
	_, _ = io.WriteString(out, fmt.Sprintf("Out:%s\n", "Hello World!"))
}
