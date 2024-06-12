package main

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

import (
	"fmt"
	"storm/data"
	"time"
)

func main() {
	data.Setup()
	fmt.Println("Started at", time.Now())
	fmt.Println("Go and run the serverless fab at ", "./documentation/serverless.sh")
	fmt.Println(`Example request:
curl -X 'PUT' -d 'abcdef' 'http://127.0.0.1:7777/portal/abcd'
Example request:
curl -X 'GET' 'http://127.0.0.1:7777/portal/abcd'`)
	for {
		// TODO running logic in a forked child process cleans up memory for long-haul
		time.Sleep(10 * time.Minute)
		fmt.Println()
	}
}
