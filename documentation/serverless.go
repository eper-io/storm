package main

import (
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
	data.RunShardIndex(TmpMemory+pathSetKey, 0, RunServerlessLambdaBurst)
	data.RunShardIndex(TmpMemory+pathSetKey, 1, RunServerlessLambdaBurst)
	for {
		// Infinite loop
		time.Sleep(10 * time.Second)
	}
}

func RunServerlessLambdaBurst(out *bytes.Buffer, in []byte, shard int) {
	(*out).WriteString(fmt.Sprintf("Shard: %d\nTime:%s\nIn:\n%s\nOut:\n%s\n", shard, time.Now().Format(time.RFC3339Nano), string(in), "Hello World!"))
}
