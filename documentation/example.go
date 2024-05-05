package main

import (
	"fmt"
	"math/rand"
	"storm/data"
	"strings"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// This example creates a sample table, where you can run a query.
// It uses a columnar index. We also run a query that does not last long.

func main() {
	db := "http://127.0.0.1:7777/data/"
	index := "http://127.0.0.1:7777/columnar/temperature"
	t := time.Now()
	r := time.Now()
	n := 0
	er := 0
	e := make(chan time.Duration)
	fmt.Printf("Inserting data...")
	b := 0
	for b < 1*1024*1024 {
		r = r.Add(-1 * time.Hour)
		temp := rand.NormFloat64()*5 + 36.0
		x := fmt.Sprintf("The temperature at %s was %f degrees Fahrenheit.", r.Format("2006010215"), temp)
		y := fmt.Sprintf("The temperature at %s was high.", r.Format("2006010215"))
		if temp < 36.3 {
			y = ""
		}

		// Maintain byte and record counters
		b = b + len(y) + len(x)
		n++

		go func(s string, d chan time.Duration) {
			u := time.Now()
			z := data.EnglangPoke(db+"?format=%25s", []byte(s))
			if string(z) == "" {
				er++
			}
			d <- time.Now().Sub(u)
		}(x, e)
		if y != "" {
			n++
			go func(s string, d chan time.Duration) {
				u := time.Now()
				data.EnglangPoke(index, []byte(s))
				d <- time.Now().Sub(u)
			}(y, e)
		}

		// Do not saturate channels
		time.Sleep(100 * time.Microsecond)
	}
	for k := 0; k < n; k++ {
		<-e
	}
	fmt.Printf("\b\b\b. Done.\n")
	fmt.Println("Kilobytes: ", (b+1023)/1024, "Records:", n, "Errors:", er, "Time per record inserted: ", (time.Now().Sub(t) / time.Duration(n)).String(), "Elapsed to insert: ", time.Now().Sub(t).String())
	fmt.Printf("Waiting for indexing...")
	time.Sleep(10 * time.Second)
	fmt.Printf("\b\b\b. Done. Querying...")
	t = time.Now()
	ret := string(data.EnglangFetch(index))
	d := fmt.Sprintf("The temperature at %s was high.", time.Now().Add(-5*time.Hour).Format("2006010215"))
	if strings.Contains(ret, d) {
		fmt.Printf("\b\b\b. Done. We found an entry for five hours ago. Elapsed to query: %s\n", time.Now().Sub(t).String())
	} else {
		fmt.Printf("\b\b\b. Done. We did not find an entry for five hours ago. Elapsed to query: %s\n", time.Now().Sub(t).String())
	}
}
