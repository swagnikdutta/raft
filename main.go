package main

import (
	"math/rand"
	"time"

	"github.com/swagnikdutta/raft/raft"
)

func main() {
	// create cluster with 3 servers
	rand.Seed(time.Now().UnixNano())
	raft.CreateCluster(3)
}
