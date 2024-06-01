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

func EnglangLoadBalancing(path string, shardList string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
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
		var results = make([]*chan []byte, 0)

		// Build shard index
		shard := ""
		rBody, _ := io.ReadAll(request.Body)
		rPath := "/" + request.URL.Path[len(path):]
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
				query := ""
				if request.URL.RawQuery != "" {
					query = "?" + request.URL.RawQuery
				}
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
							time.Sleep(time.Duration(rand.Int()%8) * time.Millisecond)
							continue
						}
						if bytes.HasPrefix(recvBytes, sentBytes[0:32]) && !bytes.Equal(recvBytes, sentBytes) {
							// Reply
							(*put) <- recvBytes[32:]
							// Acknowledge
							TmpPut(shardAddress, []byte("ack"))
							return
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

func RunShardIndex(shardKey string, shardIndex int, lambda func(out *bytes.Buffer, in []byte, i int)) {
	currentShards := TmpGet(shardKey)
	if string(currentShards) == "" {
		return
	}
	shards := string(TmpGet(shardKey))
	list := string(TmpGet(shards))
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
	TmpPut(shardAddress, []byte("ack"))
	sentBytes := []byte{}
	for {
		recvBytes := TmpGet(shardAddress)
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
			TmpPut(shardAddress, sentBytes)
			time.Sleep(time.Duration(rand.Int()%3) * time.Millisecond)
		}
		time.Sleep(time.Duration(rand.Int()%8) * time.Millisecond)
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
