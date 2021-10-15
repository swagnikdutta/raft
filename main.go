package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/swagnikdutta/raft/raft"
)

func main() {
	// create cluster with 3 servers
	log.SetFlags(0)
	rand.Seed(time.Now().UnixNano())
	raft.CreateCluster(3)
}
