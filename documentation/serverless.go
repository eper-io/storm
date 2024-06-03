package main

import (
	"bytes"
	"fmt"
	"storm/data"
	"time"
)

func main() {
	api := "https://hour.schmied.us/df94d5658feda65e9d5cdac6bcd50b8012c835ab884f0a74c5fa46e396b05ae7.tig"
	data.RunShardList(api, RunServerlessLambdaBurst)
	for {
		// Infinite loop
		time.Sleep(10 * time.Second)
	}
}

func RunServerlessLambdaBurst(out *bytes.Buffer, in []byte, shard int) {
	(*out).WriteString(fmt.Sprintf("Shard: %d\nTime:%s\nIn:\n%s\nOut:\n%s\n", shard, time.Now().Format(time.RFC3339Nano), string(in), "Hello World!"))
}
