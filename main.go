package main

import "github.com/swagnikdutta/raft/raft"

func main() {
	// create cluster with 3 servers
	raft.CreateCluster(3)
}
