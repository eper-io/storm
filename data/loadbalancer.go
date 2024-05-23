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
	"sync"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func EnglangLoadBalancing(path string, servers string) func(http.ResponseWriter, *http.Request) {
	// TODO collect PUT with delayed write and compressing data

	shardList := ""
	var balancerLock = sync.Mutex{}
	go func() {
		// TODO
		time.Sleep(100 * time.Millisecond)
		for i := 0; i < 2; i++ {
			time.Sleep(1 * time.Second)
			go func(i int) {
				rand.Seed(time.Now().UnixNano() + int64(i))
				shardAddress := fmt.Sprintf(MemCache+"/%x.tig?shard=%d", sha256.Sum256([]byte(fmt.Sprintf("%16x%16x", rand.Uint64(), rand.Uint64()))), i)
				req, _ := http.NewRequest("PUT", "http://127.0.0.1:7777"+path+"registernode", bytes.NewBuffer([]byte(shardAddress)))
				_, _ = http.DefaultClient.Do(req)
				go func(shardAddress string) {
					TmpPut(shardAddress, []byte("ack"))
					sentBytes := []byte{}
					for {
						recvBytes := TmpGet(shardAddress)
						if len(recvBytes) > 0 && len(sentBytes) > 0 && bytes.HasPrefix(recvBytes, sentBytes) {
							continue
						}
						if len(recvBytes) == 0 {
							sentBytes = []byte{}
							continue
						}
						if len(recvBytes) > 32 {
							x := bytes.NewBuffer(recvBytes[0:32])
							x.WriteString("Hello World! " + time.Now().Format(time.RFC3339Nano) + "\n")
							sentBytes = x.Bytes()
							TmpPut(shardAddress, sentBytes)
							sentBytes = sentBytes[0:32]
							time.Sleep(time.Duration(rand.Int()%3) * time.Millisecond)
						}
						time.Sleep(time.Duration(rand.Int()%8) * time.Millisecond)
					}
				}(shardAddress)
			}(i)
		}
	}()

	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "PUT" && request.URL.Path == path+"registernode" {
			const shardLifeTime = 100 * time.Hour
			body, err := io.ReadAll(request.Body)
			if err != nil {
				return
			}
			node := strings.Trim(string(body), " \r\n\t")
			balancerLock.Lock()
			shardList = shardList + node + "\n"
			balancerLock.Unlock()
			go func(node string) {
				time.Sleep(shardLifeTime)
				balancerLock.Lock()
				x := bufio.NewScanner(bytes.NewBufferString(shardList))
				z := ""
				for x.Scan() {
					y := x.Text()
					if y != node {
						z = z + y + "\n"
					}
				}
				shardList = z
				balancerLock.Unlock()
			}(node)
			return
		}
		var results = make(chan []byte)

		balancerLock.Lock()
		shards := shardList
		balancerLock.Unlock()
		// Build shard count
		m := uint64(0)
		x := bufio.NewScanner(bytes.NewBufferString(shards))
		for x.Scan() {
			y := x.Text()
			if strings.TrimSpace(y) != "" {
				m++
			}
		}

		// Build shard index
		shard := ""
		rBody, _ := io.ReadAll(request.Body)
		rPath := request.URL.Path[len(path):]
		shard = GetShard(rPath, rBody, m)

		// Send request to shards
		x = bufio.NewScanner(bytes.NewBufferString(shards))
		n := 0
		for x.Scan() {
			shardAddress := x.Text()
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
				sent.WriteString(rPath + "\n")
				sent.Write(rBody)
				sentBytes := sent.Bytes()
				n++
				go func(shardAddress string, sentBytes []byte, put chan []byte) {
					var recvBytes []byte
					// Just use the good old Ethernet algorithm
					// It has been working for decades for datacenter networks.
					TmpPut(shardAddress, sentBytes)
					for {
						recvBytes = TmpGet(shardAddress)
						if len(recvBytes) == 0 {
							// TODO
							recvBytes = TmpGet(shardAddress)
						}
						if len(recvBytes) < 32 {
							// Retry
							// This might sound too unprofessional.
							// The basic idea is that 99.9% of the cases have atomicity, consistency, integrity.
							// Many programming languages assign 50%+ resources and code to solve these.
							// We do it here with just three lines.
							TmpPut(shardAddress, sentBytes)
							continue
						}
						if bytes.HasPrefix(recvBytes, sentBytes[0:32]) && !bytes.Equal(recvBytes, sentBytes) {
							// Reply
							put <- recvBytes[32:]
							// Acknowledge
							TmpPut(shardAddress, []byte("ack"))
							return
						}
						time.Sleep(time.Duration(rand.Int()%8) * time.Millisecond)
					}
				}(shardAddress, sentBytes, results)
			}
		}
		for i := 0; i < n; i++ {
			x := <-results
			_, _ = writer.Write(x)
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
