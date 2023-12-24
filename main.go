package main

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

func main() {
	s := spinner.New(spinner.CharSets[39], 100*time.Millisecond)
	s.Start()
	time.Sleep(4 * time.Second)
	s.Stop()
}

func loop() {
	max, iter := 255, 2
	for i := 0; i < iter; i++ {
		iterateNums(max)
	}
}

func iterateNums(n int) {
	for i := 0; i < n; i++ {
		fmt.Printf("%d\t", i)
	}
}
