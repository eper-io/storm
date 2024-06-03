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
	y := in
	x := bytes.IndexByte(in, '\n')
	u := ""
	if x != -1 {
		z := x + 1
		x = bytes.IndexByte(in[z:], '\n')
		u = string(in[z : z+x])
		x = z + x + 1
	} else {
		// Fallback path
		(*out).WriteString(fmt.Sprintf("Shard: %d\nTime:%s\nIn:\n%s\nOut:\n%s\n", shard, time.Now().Format(time.RFC3339Nano), string(in), "Hello World!"))
	}
	if x != -1 {
		y = in[x:]
	}
	request := bytes.NewBuffer(y)
	req, _ := http.NewRequest("PUT", u, request)
	var z serverlessHttpWriter
	MyHttpHandler(&z, req)
	out.WriteString(fmt.Sprintf("Shard: %d\nTime:%s\n%s", shard, time.Now().Format(time.RFC3339Nano), string(z.out.Bytes())))
}

func MyHttpHandler(out http.ResponseWriter, in *http.Request) {
	x, _ := ioutil.ReadAll(in.Body)
	_, _ = io.WriteString(out, fmt.Sprintf("In:%s\n", string(x)))
	_, _ = io.WriteString(out, fmt.Sprintf("Out:%s\n", "Hello World!"))
}
