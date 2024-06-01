package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"storm/data"
	"time"
)

func main() {
	if os.Getenv("IMPLEMENTATION") == "" {
		_ = os.Setenv("IMPLEMENTATION", "./res/demo.txt")
	}
	shards := string(data.TmpGet("https://data.schmied.us/295b5019b70eb5f8d145b2ab27a2f26b1346a404fab5c1b3712d44bead215bd7.tig"))
	fmt.Println(shards)
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
