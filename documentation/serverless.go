package main

import (
	"bufio"
	"bytes"
	"fmt"
	"storm/data"
	"time"
)

func main() {
	shards := data.ShardList
	x := bufio.NewScanner(bytes.NewBufferString(shards))
	for x.Scan() {
		shardAddress := x.Text()
		time.Sleep(1 * time.Second)
		go func(shardAddress string) {
			data.RunShard(shardAddress, RunServerlessLambdaBurst)
		}(shardAddress)
	}
	time.Sleep(100 * time.Hour)
}

func RunServerlessLambdaBurst(out *bytes.Buffer, in []byte) {
	(*out).WriteString(fmt.Sprintf("Hello World! %s %d\n", time.Now().Format(time.RFC3339Nano), len(string(in))))
}
