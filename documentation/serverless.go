package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"storm/data"
	"time"
)

func main() {
	TmpMemory := "https://hour.schmied.us"
	key := "55E2C4BE-A96C-46A2-AADD-80715E0A16CD"
	pathSetKey := fmt.Sprintf("/%x.tig", sha256.Sum256([]byte(key)))
	currentShards := data.TmpGet(TmpMemory + pathSetKey)
	if string(currentShards) == "" {
		fmt.Println(fmt.Sprintf("no shards at %s => %s", key, fmt.Sprintf(TmpMemory)+pathSetKey))
		os.Exit(1)
	}
	shards := string(data.TmpGet(TmpMemory + pathSetKey))
	list := string(data.TmpGet(shards))
	x := bufio.NewScanner(bytes.NewBufferString(list))
	n := 0
	for x.Scan() {
		shardAddress := x.Text()
		time.Sleep(1 * time.Millisecond)
		go func(shardAddress string, shardIndex int) {
			data.RunShard(shardAddress, shardIndex, RunServerlessLambdaBurst)
		}(shardAddress, n)
		n++
	}
	time.Sleep(100 * time.Hour)
}

func RunServerlessLambdaBurst(out *bytes.Buffer, in []byte, shard int) {
	(*out).WriteString(fmt.Sprintf("Shard: %d\nTime:%s\nIn:\n%s\nOut:\n%s\n", shard, time.Now().Format(time.RFC3339Nano), string(in), "Hello World!"))
}
