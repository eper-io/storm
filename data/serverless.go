package data

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func RunShardClient(instructions io.Reader, done *sync.WaitGroup, handler http.HandlerFunc) {
	implementation := bufio.NewScanner(instructions)
	for implementation.Scan() {
		command := strings.TrimSpace(implementation.Text())
		shard := -1
		api := "noapiurl"
		n, _ := fmt.Sscanf(command, "Run shard id %d from api pointed by %s key.", &shard, &api)
		if n == 2 {
			fmt.Println(command)
		}
		done.Add(1)
		go func(api string, shard int) {
			snapshot := RunShardListHttp(api, shard, handler)
			for {
				// Check for api version modification let restart to update, if needed
				time.Sleep(10 * time.Second)
				current := ServerlessGet(api)
				if !bytes.Equal(current, snapshot) {
					//fmt.Println(string(current), string(snapshot))
					done.Done()
					return
				}
			}
		}(ApiCache+api, shard)
	}
}

func RunShardListHttp(shardListKey string, shardIndex int, lambda http.HandlerFunc) []byte {
	return RunShardList(shardListKey, shardIndex, func(out *bytes.Buffer, in []byte, i int) {
		RunServerlessLambdaBurstOnHttp(out, in, i, lambda)
	})
}

func RunShardList(shardListKey string, shardIndex int, lambda func(out *bytes.Buffer, in []byte, i int)) []byte {
	currentShards := ServerlessGet(shardListKey)
	if string(currentShards) == "" {
		return currentShards
	}
	shards := string(ServerlessGet(shardListKey))
	list := string(ServerlessGet(shards))
	x := bufio.NewScanner(bytes.NewBufferString(list))
	n := 0
	for x.Scan() {
		t := strings.TrimSpace(x.Text())
		if strings.TrimSpace(t) != "" {
			n++
		}
	}
	for i := 0; i < n; i++ {
		if shardIndex == -1 || shardIndex == i {
			//fmt.Println("Running shard", i)
			RunSingleShard(list, i, lambda)
		}
	}
	return currentShards
}

func RunSingleShard(list string, shardIndex int, lambda func(out *bytes.Buffer, in []byte, i int)) {
	x := bufio.NewScanner(bytes.NewBufferString(list))
	n := 0
	for x.Scan() {
		shardAddress := x.Text()
		time.Sleep(1 * time.Millisecond)
		if n == shardIndex {
			go func(shardAddress string, shardIndex int) {
				RunShard(shardAddress, shardIndex, lambda)
			}(shardAddress, n)
		}
		n++
	}
}

func RunShard(shardAddress string, i int, lambda func(out *bytes.Buffer, in []byte, i int)) {
	ServerlessPut(shardAddress, []byte("ack"))
	sentBytes := []byte{}
	for {
		recvBytes := ServerlessGet(shardAddress)
		if len(recvBytes) > 0 && len(sentBytes) > 0 && bytes.Equal(recvBytes, sentBytes) {
			continue
		}
		if len(recvBytes) == 0 {
			sentBytes = []byte{}
			continue
		}
		if len(recvBytes) > 32 {
			x := bytes.NewBuffer(recvBytes[0:32])
			lambda(x, recvBytes, i)
			sentBytes = x.Bytes()
			ServerlessPut(shardAddress, sentBytes)
			time.Sleep(time.Duration(rand.Int()%3) * time.Millisecond)
		}
		time.Sleep(time.Duration(rand.Int()%8) * time.Millisecond)
	}
}

func newServerlessResponseWriter(b *bytes.Buffer) serverlessHttpWriter {
	ret := serverlessHttpWriter{}
	ret.out = b
	ret.header = http.Header{}
	ret.statusCode = http.StatusNotImplemented
	return ret
}

type serverlessHttpWriter struct {
	header     http.Header
	out        *bytes.Buffer
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

func RunServerlessLambdaBurstOnHttp(out *bytes.Buffer, in []byte, shard int, httpFunc http.HandlerFunc) {
	x := bytes.SplitN(in, []byte{'\n'}, 5)
	if len(x) != 5 {
		// Fallback path
		(*out).WriteString(fmt.Sprintf("Shard: %d\nTime:%s\nIn:\n%s\nOut:\n%s\n", shard, time.Now().Format(time.RFC3339Nano), string(in), "Hello World!"))
		return
	}
	var selected int
	s := string(x[1])
	_, _ = fmt.Sscanf(s, "Selected shard is %d .", &selected)
	m := string(x[2])
	u := string(x[3])
	request := bytes.NewBuffer(x[4])
	req, _ := http.NewRequest(m, u, request)
	req.Header.Set("Shard", strconv.Itoa(shard))
	req.Header.Set("Selected", strconv.Itoa(selected))
	z := newServerlessResponseWriter(out)
	httpFunc(&z, req)
}

func MockHttpHandler(out http.ResponseWriter, in *http.Request) {
	x := []byte{}
	if in.Body != nil {
		x, _ = ioutil.ReadAll(in.Body)
	}
	_, _ = io.WriteString(out, fmt.Sprintf("Shard: %s\nSelected: %s\nTime:%s\n", in.Header.Get("Shard"), in.Header.Get("Selected"), time.Now().Format(time.RFC3339Nano)))
	_, _ = io.WriteString(out, fmt.Sprintf("Path:%s\n", in.URL.String()))
	_, _ = io.WriteString(out, fmt.Sprintf("In:%s\n", string(x)))
	_, _ = io.WriteString(out, fmt.Sprintf("Out:%s\n", "Hello World!"))
	_, _ = io.WriteString(out, fmt.Sprintf("%s\n", "---"))
}

func ServerlessGet(address string) []byte {
	x, _ := http.NewRequest("GET", address, bytes.NewBuffer(nil))
	resp, err := http.DefaultClient.Do(x)
	y := []byte{}
	if err == nil {
		y, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	return y
}

func ServerlessDelete(address string) []byte {
	x, _ := http.NewRequest("DELETE", address, bytes.NewBuffer(nil))
	resp, err := http.DefaultClient.Do(x)
	y := []byte{}
	if err == nil {
		y, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	return y
}

func ServerlessPut(address string, put []byte) []byte {
	x, _ := http.NewRequest("PUT", address, bytes.NewBuffer(put))
	resp, err := http.DefaultClient.Do(x)
	y := []byte{}
	if err == nil {
		y, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	return y
}
