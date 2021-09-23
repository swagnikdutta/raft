package main

import (
	"fmt"

	"github.com/swagnikdutta/raft/raft"
)

func main() {
	raft.NewServer()
	fmt.Println("test")
}
