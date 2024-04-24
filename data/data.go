package data

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"runtime"
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

var MemCache = "https://hour.schmied.us"
var IndexedShortCache = map[string][]byte{}
var ListedCache = make([]string, 0)
var LastSnapshot = ""

// This should only get access from this process or Docker container
// It is ideally the $(HOME)
var Dir = "/tmp"

func Setup() {
	if os.Getenv("IMPLEMENTATION") == "" {
		_ = os.Setenv("IMPLEMENTATION", "./res/demo.txt")
	}

	// We do a very conservative approach to logging.
	// We log the configuration input, but nothing else.
	// Extensive logging just opens vulnerabilities.
	// It also costs a lot with diminishing returns.
	// We rather make sure that the run can be reproduced.
	implementation := os.Getenv("IMPLEMENTATION")
	if implementation != "" {
		var rc func() = nil
		var lines *bufio.Scanner = nil
		if strings.HasPrefix(implementation, "http") {
			resp, _ := http.Get(implementation)

			if resp != nil && resp.Body != nil {
				lines = bufio.NewScanner(resp.Body)
				rc = func() {
					resp.Body.Close()
				}
			}
		} else {
			x, _ := os.Open(implementation)
			lines = bufio.NewScanner(x)
			rc = func() {
				x.Close()
			}
		}
		defer rc()
		for lines.Scan() {
			line := strings.TrimSpace(lines.Text())
			var value string
			n, _ := fmt.Sscanf(line, "Set apikey to %s value.", &value)
			if n == 1 {
				_ = os.WriteFile(path.Join(Dir, "apikey"), EnglangFetch(value), 0600)
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
			}
			n, _ = fmt.Sscanf(line, "Set certificate.pem to %s value.", &value)
			if n == 1 {
				_ = os.WriteFile(path.Join(Dir, "certificate.pem"), EnglangFetch(value), 0600)
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
			}
			n, _ = fmt.Sscanf(line, "Set key.pem to %s value.", &value)
			if n == 1 {
				_ = os.WriteFile(path.Join(Dir, "key.pem"), EnglangFetch(value), 0600)
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
			}
			n, _ = fmt.Sscanf(line, "Set backend to %s value.", &value)
			if n == 1 {
				MemCache = value
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
			}
			n, _ = fmt.Sscanf(line, "Load memory snapshot from %s value.", &value)
			if n == 1 {
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
				LastSnapshot = value
				snapshotBytes := EnglangFetch(value)
				snapshotReader := bytes.NewBuffer(snapshotBytes)
				snapshotLines := bufio.NewScanner(snapshotReader)
				for snapshotLines.Scan() {
					lineNumber := 0
					value = snapshotLines.Text()
					n, _ = fmt.Sscanf(value, "Block number %06d is hashed as %s value.", &lineNumber, &value)
					if n > 1 {
						value = value[len(value)-64-len("/.tig") : len(value)]
						ListedCache = append(ListedCache, value)
						target, err := url.Parse(MemCache)
						if err != nil {
							return
						}
						target.Path = fmt.Sprintf("%s", value)
						target.RawQuery = ""
						data := EnglangFetch(target.String())
						if data != nil && len(data) > 0 && len(data) < 512 {
							IndexedShortCache[value] = data
							fmt.Println(string(data))
						}
					}
				}
			}
			minutes := int64(0)
			n, _ = fmt.Sscanf(line, "Save memory snapshot every %d minutes.", &minutes)
			if n == 1 {
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
				go func() {
					i := 0
					start := time.Now()
					o := time.Duration(0)
					for {
						time.Sleep(time.Duration(minutes) * time.Minute)
						buf := bytes.NewBuffer([]byte{})
						for k, v := range ListedCache {
							buf.WriteString(fmt.Sprintf("Block number %06d is hashed as %s value.", k, v) + "\n")
						}
						buf1 := EnglangPoke(MemCache+"&format=%25s", buf.Bytes())
						root, _, _ := strings.Cut(MemCache, "?")
						LastSnapshot = root + string(buf1)
						fmt.Printf(".")
						i++
						if i%10 == 0 {
							fmt.Printf("\n")
							m := time.Now().Sub(start) / 24 * time.Hour
							if m > o+1 {
								o = m
								if len(LastSnapshot) > 5 {
									fmt.Printf("%s... set as snapshot.\n", LastSnapshot[len(LastSnapshot)-5:])
								}
							}
						}
					}
				}()
			}
			var decimal int
			n, _ = fmt.Sscanf(line, "Listen http on %d port.", &decimal)
			if n == 1 {
				go func() {
					_ = http.ListenAndServe(":7777", nil)
				}()
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
			}
			n, _ = fmt.Sscanf(line, "Listen https on %d port with key.pem and certificate.pem set.", &decimal)
			if n == 1 {
				go func() {
					_, err := os.Stat(path.Join(Dir, "key.pem"))
					if err == nil {
						_, err = tls.LoadX509KeyPair(path.Join(Dir, "certificate.pem"), path.Join(Dir, "key.pem"))
						if err != nil {
							fmt.Println(err)
						}
						err = http.ListenAndServeTLS(":7777", path.Join(Dir, "certificate.pem"), path.Join(Dir, "key.pem"), nil)
					}
				}()
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
			}
			n, _ = fmt.Sscanf(line, "Response hello on %s path.", &value)
			if n == 1 {
				http.HandleFunc(value, func(writer http.ResponseWriter, request *http.Request) {
					_, _ = io.WriteString(writer, "<body>Hello World!</body>\n")
				})
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
			}
			var value1 string
			n, _ = fmt.Sscanf(line, "Proxy %s on %s path.", &value1, &value)
			if n == 2 {
				http.HandleFunc(value, EnglangProxy(value, value1))
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
			}
			n, _ = fmt.Sscanf(line, "Log on %s path and check modifications on %s path.", &value, &value1)
			if n == 2 {
				http.HandleFunc(value, EnglangLog())
				http.HandleFunc(value1, func(writer http.ResponseWriter, request *http.Request) {
					for {
						select {
						case item := <-Modifications:
							_, _ = writer.Write(item)
						default:
							return
						}
					}
				})
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
			}
			n, _ = fmt.Sscanf(line, "Response bursts on %s path.", &value)
			if n == 1 {
				http.HandleFunc(value, EnglangBurst(value))
				_, file, line1, _ := runtime.Caller(0)
				fmt.Printf("File: %s\nLine: %d\n", file, line1)
				fmt.Println(line)
			}
		}
	}
}

var bursts = map[string]*chan string{}
var burstRun sync.Mutex

func EnglangBurst(path string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		path1 := request.URL.Path[len(path):]
		channel, ok := bursts[path1]
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

var Modifications = make(chan []byte)

func EnglangLog() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Print(".")
		if r.Method == "PUT" || r.Method == "POST" || r.Method == "DELETE" {
			x := bytes.Buffer{}
			_, _ = io.Copy(&x, r.Body)
			_, _ = io.WriteString(&x, "\n")
			y := x.Bytes()
			go func(x []byte) {
				Modifications <- x
			}(y)
		}
	}
}

func EnglangProxy(prefix string, remoteURL string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		remoteURL := remoteURL
		targetURL, err := url.Parse(remoteURL)
		if err != nil {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		r.Host = targetURL.Host
		r.URL.Path = r.URL.Path[len(prefix):]
		proxy.ServeHTTP(w, r)
	}
}

func EnglangFetch(url string) (ret []byte) {
	resp, _ := http.Get(url)
	if resp != nil && resp.Body != nil {
		ret, _ = io.ReadAll(resp.Body)
	}
	return
}

func EnglangPoke(url string, input []byte) (ret []byte) {
	ret = []byte{}
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(input))
	if err != nil {
		// Do you need an error? Check the server side.
		// Logs are an attack surface.
		return
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := client.Do(req)
	if err != nil {
		// Do you need an error? Check the server side.
		// Logs are an attack surface.
		return
	}
	ret, _ = io.ReadAll(res.Body)
	_ = res.Body.Close()
	return
}

func EnglangSet(url string, input []byte) (ret []byte) {
	ret = []byte{}
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(input))
	if err != nil {
		// Do you need an error? Check the server side.
		// Logs are an attack surface.
		return
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := client.Do(req)
	if err != nil {
		// Do you need an error? Check the server side.
		// Logs are an attack surface.
		return
	}
	ret, _ = io.ReadAll(res.Body)
	_ = res.Body.Close()
	return
}

func EnglangDrop(url string) bool {
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		// Do you need an error? Check the server side.
		// Logs are an attack surface.
		return false
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	res, err := client.Do(req)
	if err != nil {
		// Do you need an error? Check the server side.
		// Logs are an attack surface.
		return false
	}
	if res.StatusCode != http.StatusOK {
		return false
	}
	_ = res.Body.Close()
	return true
}

func EnglangGetFields(s string, f ...string) (y []string) {
	x := append(make([]string, 0), f...)
	y = make([]string, 0)
	for k, v := range x {
		if v == "" {
			y = append(y, s)
			return
		}
		b, e, q := strings.Cut(s, v)
		if !q {
			return
		}
		if k > 0 {
			y = append(y, b)
		}
		s = e
	}
	return
}

func EnglangGetBlocks(s []byte, f ...[]byte) (y [][]byte) {
	x := append(make([][]byte, 0), f...)
	y = make([][]byte, 0)
	for k, v := range x {
		if v == nil {
			y = append(y, s)
			return
		}
		b, e, q := bytes.Cut(s, v)
		if !q {
			return
		}
		if k > 0 {
			y = append(y, b)
		}
		s = e
	}
	return
}