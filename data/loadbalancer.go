package data

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// This is an educational serverless lambda implementation with bursts.
// We use bursts that are pull model instead of push model of callable functions.
// Bursts allow super flexible and dynamic use of scripts running on expensive GPU machines.
// This saves money, so that you can buy even more GPUs.
// Not leaving an external endpoint enhances the security above the current level of tech companies.
// Burst runners call out from servers, so port scanners cannot even find out what they do.
// Eventually all HTTPS traffic is to be replaced with atomic datagrams.
// You may notice less error handling. We expect an operating environment of above than average reliability.
// Also, shards aggregate all GET responses but forward PUT only to one shard.
// This replaces mutexes, so that it is easy to handle integrity.
// Each shard consumes requests in a row, so it is atomic by design. The same data goes to the same shard.
// Performance can be adjusted by setting 100, 10000, 1 million shards respectively.
// It is implemented with Tig, but it is easy to attach MSSQL, S3, Redis, Mongodb, Cassandra, on the other side.
// You noticed that we stick to less than few hundred lines of code per feature.
// We remove the extra complexity of empirical arts of the programming platforms.
// This enables cheaper artificial intelligence training solutions.
// TODO collect PUT with delayed write and data compressing

// Use this for testing:
// curl -X 'PUT' -d 'abcdef' 'http://127.0.0.1:7777/portal/abcd'
// curl -X 'GET' 'http://127.0.0.1:7777/portal/abcde'

func EnglangLoadBalancing(path1 string, shardList string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		EnglangSetMime(writer, request)

		shards := shardList
		// Build shard count
		m := uint64(0)
		x := bufio.NewScanner(bytes.NewBufferString(shards))
		for x.Scan() {
			y := x.Text()
			if strings.TrimSpace(y) != "" {
				m++
			}
		}
		if m == 0 {
			http.Error(writer, "No shards", http.StatusServiceUnavailable)
			return
		}
		var results = make([]*chan []byte, 0)

		// Build shard index
		shard := ""
		rBody, _ := io.ReadAll(request.Body)
		// We use absolute to save dev time
		// rPath := "/" + request.URL.Path + request.URL.Path[len(path):]
		rPath := request.URL.Path
		shard = GetShard(rPath, rBody, m)

		// Send request to shards
		x = bufio.NewScanner(bytes.NewBufferString(shards))
		selected := rand.Uint32() % uint32(m)
		n := 0
		for x.Scan() {
			shardAddress := x.Text()
			if strings.TrimSpace(shardAddress) == "" {
				continue
			}
			useThisShard := false
			if request.Method == "GET" {
				// Aggregate of all shards
				useThisShard = true
			}
			if !useThisShard && strings.Contains(shardAddress, shard) {
				// Return from the selected shard
				useThisShard = true
			}
			if useThisShard {
				sent := bytes.NewBufferString(fmt.Sprintf("%16x%16x\n", rand.Uint64(), rand.Uint64()))
				query := ""
				if request.URL.RawQuery != "" {
					query = "?" + request.URL.RawQuery
				}
				sent.WriteString(fmt.Sprintf("Selected shard is %d ."+"\n", selected))
				sent.WriteString(request.Method + "\n")
				sent.WriteString(rPath + query + "\n")
				sent.Write(rBody)
				sentBytes := sent.Bytes()
				ch := make(chan []byte)
				results = append(results, &ch)
				n++
				go func(shardAddress string, sentBytes []byte, put *chan []byte) {
					var recvBytes []byte
					// Just use the good old Ethernet algorithm
					// It has been working for decades for datacenter networks.
					start := time.Now()
					for {
						ret := TmpPut(shardAddress+"?setifnot=1&format=%25s", sentBytes)
						if len(ret) > 0 {
							break
						}
						if time.Now().Sub(start).Seconds() > 10 {
							// TODO error message
							return
						}
						time.Sleep(time.Duration(rand.Int()%8) * time.Millisecond)
					}

					for {
						recvBytes = TmpGet(shardAddress)
						// TODO Needed?
						//if len(recvBytes) == 0 {
						//	recvBytes = TmpGet(shardAddress)
						//}
						if bytes.HasPrefix(recvBytes, sentBytes[0:32]) && !bytes.Equal(recvBytes, sentBytes) {
							// Reply
							(*put) <- recvBytes[32:]
							// Acknowledge
							// TODO Needed?
							//TmpDelete(shardAddress)
							TmpPut(shardAddress, []byte(""))
							return
						}
						if !bytes.Equal(recvBytes[0:32], sentBytes[0:32]) {
							// Retry
							// This might sound too unprofessional.
							// The basic idea is that 99.9% of the cases have atomicity, consistency, integrity.
							// Many programming languages assign 50%+ resources and code to solve these.
							// We do it here with just three lines.
							for {
								ret := TmpPut(shardAddress+"?setifnot=1&format=%25s", sentBytes)
								if len(ret) > 0 {
									break
								}
								if time.Now().Sub(start).Seconds() > 10 {
									// TODO error message
									return
								}
							}
							time.Sleep(time.Duration(rand.Int()%8) * time.Millisecond)
							continue
						}
						time.Sleep(time.Duration(rand.Int()%8) * time.Millisecond)
					}
				}(shardAddress, sentBytes, &ch)
			}
		}
		for i := 0; i < n; i++ {
			x := <-(*results[i])
			_, _ = writer.Write(x)
		}
	}
}

func EnglangSetMime(writer http.ResponseWriter, request *http.Request) {
	name := request.URL.Path
	encoding := map[string]string{
		"htm":  "text/html",
		"html": "text/html",
		"png":  "image/png",
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"gif":  "image/gif",
		"css":  "text/css",
		"js":   "text/javascript",
		"txt":  "text/plain",
		"md":   "text/markdown",
	}
	for k, v := range encoding {
		if strings.HasSuffix(name, k) {
			writer.Header().Set("Content-Type", v)
		}
	}
}

func GetShard(path string, a []byte, shardCount uint64) string {
	if shardCount == 0 {
		return ""
	}
	shard := [32]byte{}
	if len(path) > 0 {
		shard = sha256.Sum256([]byte(path))
	}
	if len(a) > 0 && len(path) == 0 {
		shard = sha256.Sum256(a)
	}
	var shardNum uint64
	_ = binary.Read(bytes.NewBuffer([]byte{shard[0], shard[1], shard[2], 0, 0, 0, 0, 0}), binary.LittleEndian, &shardNum)
	return fmt.Sprintf("shard=%d", shardNum%shardCount)
}

func TmpGet(address string) []byte {
	x, _ := http.NewRequest("GET", address, bytes.NewBuffer(nil))
	resp, err := http.DefaultClient.Do(x)
	y := []byte{}
	if err == nil {
		y, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	return y
}

func TmpDelete(address string) []byte {
	x, _ := http.NewRequest("DELETE", address, bytes.NewBuffer(nil))
	resp, err := http.DefaultClient.Do(x)
	y := []byte{}
	if err == nil {
		y, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	return y
}

func TmpPut(address string, put []byte) []byte {
	x, _ := http.NewRequest("PUT", address, bytes.NewBuffer(put))
	resp, err := http.DefaultClient.Do(x)
	y := []byte{}
	if err == nil {
		y, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
	return y
}
