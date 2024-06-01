package data

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// ## References
// ### Modern Columnar Databases
// ISBN-13: 9783319373898
// ISBN-13: 978-0131873254

var bursts = map[string]*chan string{}
var burstRun sync.Mutex
var commitFrequency = 10 * time.Second

func EnglangSearch(unique string) func(http.ResponseWriter, *http.Request) {
	go func() {
		for {
			time.Sleep(commitFrequency)
			g := bytes.Buffer{}
			h := bytes.Buffer{}
			for k, _ := range bursts {
				EnglangSearchIndex(k, "PERSIST", &g, &h)
			}
		}
	}()
	return func(writer http.ResponseWriter, request *http.Request) {
		method := request.Method
		f := io.Writer(writer)
		e := bytes.NewBuffer([]byte{})
		_, _ = io.Copy(e, request.Body)
		_ = request.Body.Close()
		path1 := request.URL.Path[len(unique):]
		if EnglangSearchIndex(path1, method, e, f) {
			return
		}
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func EnglangSearchIndex(path string, method string, request *bytes.Buffer, response io.Writer) bool {
	pathSetKey := fmt.Sprintf("/%x.tig", sha256.Sum256([]byte(path)))
	pathGetKey := pathSetKey
	burstRun.Lock()
	channel, ok := bursts[path]
	burstRun.Unlock()
	if !ok {
		burst := make(chan string)
		burstRun.Lock()
		bursts[path] = &burst
		burstRun.Unlock()
		channel = &burst
	}
	if method == "PUT" || method == "POST" {
		// add small "columnar" item
		go func(x io.Reader, y *chan string) {
			z, _ := io.ReadAll(x)
			*y <- string(z)
		}(request, channel)
		return true
	}
	if method == "PERSIST" {
		// persist or commit
		a := bytes.Buffer{}
		first := true
		start := time.Now()
		for time.Now().Sub(start) < commitFrequency {
			select {
			case msg := <-*channel:
				if first {
					first = false
					burstRun.Lock()
					commit := string(EnglangFetch(MemCache + pathGetKey))
					burstRun.Unlock()
					if commit != "" {
						d := EnglangFetch(commit)
						a.Write(d)
					}
				}
				// Potential logarithmic insertion location
				_, _ = a.WriteString(msg)
				_, _ = io.WriteString(&a, "\n")
			default:
				start = time.Now().Add(-time.Hour)
			}
		}
		z := a.Bytes()
		if len(z) > 0 {
			b := fmt.Sprintf(MemCache+"/%x.tig", sha256.Sum256(z))
			// Persist
			EnglangPoke(MemCache, z)
			// Commit
			burstRun.Lock()
			key := string(EnglangPoke(MemCache+pathSetKey+"?format=*", []byte(b)))
			if key != pathGetKey {
				fmt.Println(key)
			}
			burstRun.Unlock()
		}
	}
	if method == "GET" {
		burstRun.Lock()
		commit := string(EnglangFetch(MemCache + pathGetKey))
		burstRun.Unlock()
		_, _ = response.Write(EnglangFetch(commit))
		return true
	}
	return false
}

func EnglangBurst(path string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		path1 := request.URL.Path[len(path):]
		burstRun.Lock()
		channel, ok := bursts[path1]
		burstRun.Unlock()
		if !ok {
			burst := make(chan string)
			burstRun.Lock()
			bursts[path1] = &burst
			burstRun.Unlock()
			channel = &burst
			if path1 == "hello" {
				go func(burst chan string) {
					for i := 0; i < 10; i++ {
						burst <- fmt.Sprintf("Hello World! %d", i)
					}
				}(burst)
			}
		}
		if request.Method == "PUT" || request.Method == "POST" {
			x := bytes.NewBuffer([]byte{})
			_, _ = io.Copy(x, request.Body)
			_ = request.Body.Close()
			go func(x io.Reader, y *chan string) {
				z, _ := io.ReadAll(x)
				*y <- string(z)
			}(x, channel)
			return
		}
		if request.Method == "GET" || request.Method == "DELETE" {
			start := time.Now()
			for time.Now().Sub(start) < 3*time.Second {
				select {
				case msg := <-*channel:
					_, _ = io.WriteString(writer, msg)
					if request.Method == "DELETE" {
						burstRun.Lock()
						delete(bursts, path1)
						burstRun.Unlock()
					}
					return
				default:
				}
			}
			http.NotFound(writer, request)
			return
		}
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
	}
}
