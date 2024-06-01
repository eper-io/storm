package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"storm/data"
	"time"
)

func main() {
	TmpMemory := "https://hour.schmied.us"
	key := "55E2C4BE-A96C-46A2-AADD-80715E0A16CD"
	pathSetKey := fmt.Sprintf("/%x.tig", sha256.Sum256([]byte(key)))
	RunShardFromKey(TmpMemory+pathSetKey, 0)
	RunShardFromKey(TmpMemory+pathSetKey, 1)
	time.Sleep(100 * time.Hour)
}

func RunShardFromKey(shardKey string, shardIndex int) {
	currentShards := data.TmpGet(shardKey)
	if string(currentShards) == "" {
		return
	}
	shards := string(data.TmpGet(shardKey))
	list := string(data.TmpGet(shards))
	x := bufio.NewScanner(bytes.NewBufferString(list))
	n := 0
	for x.Scan() {
		shardAddress := x.Text()
		time.Sleep(1 * time.Millisecond)
		if n == shardIndex {
			go func(shardAddress string, shardIndex int) {
				data.RunShard(shardAddress, shardIndex, RunServerlessLambdaBurst)
			}(shardAddress, n)
		}
		n++
	}
}

func RunServerlessLambdaBurst(out *bytes.Buffer, in []byte, shard int) {
	(*out).WriteString(fmt.Sprintf("Shard: %d\nTime:%s\nIn:\n%s\nOut:\n%s\n", shard, time.Now().Format(time.RFC3339Nano), string(in), "Hello World!"))
}
