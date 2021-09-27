package main

import (
	"fmt"
	"sync"

	"github.com/swagnikdutta/raft/raft"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go raft.CreateNewServer(&wg)
	go raft.CreateNewServer(&wg)
	wg.Wait()

	fmt.Println("test")
}
