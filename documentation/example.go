package main

import (
	"fmt"
	"math/rand"
	"storm/data"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// This example creates a sample table, where you can run a query.
// It uses a columnar index.

func main() {
	x := time.Now()
	i := 0
	er := 0
	e := make(chan time.Duration)
	for n := 0; n < 1*1024*1000; {
		x = x.Add(-1 * time.Hour)
		temp := rand.NormFloat64()*5 + 36.0
		y := fmt.Sprintf("The temperature at %s was %f degrees Fahrenheit.", x.Format("2006010215"), temp)
		a := fmt.Sprintf("The temperature at %s was high.", x.Format("2006010215"))
		if temp < 36.3 {
			a = ""
		}
		i++
		go func(s string, d chan time.Duration) {
			c := time.Now()
			z := data.EnglangPoke(data.MemCache+"?format=%25s", []byte(s))
			if string(z) == "" {
				er++
			}
			d <- time.Now().Sub(c)
		}(y, e)
		i++
		go func(s string, d chan time.Duration) {
			c := time.Now()
			z := data.EnglangPoke(data.MemCache+"?format=%25s", []byte(s))
			if string(z) == "" {
				er++
			}
			d <- time.Now().Sub(c)
		}(a, e)
		n = n + 200
		time.Sleep(100 * time.Microsecond)
	}
	for k := 0; k < i; k++ {
		<-e
	}
	fmt.Println(i*200, i, er, "T", (time.Now().Sub(x)).String())
	//ret := data.EnglangPoke(data.MemCache)
}
