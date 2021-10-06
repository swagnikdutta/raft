package raft

import (
	"fmt"
	"sync"
)

type Cluster struct {
	servers []*Server
}

func (c *Cluster) connect(n int) {
	for i := 0; i < n; i++ {
		c.servers[i].ConnectToPeers()
	}
}

func CreateCluster(n int) {
	cluster := &Cluster{}
	cluster.servers = make([]*Server, n)

	// iterate n times, create n servers, a
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cluster.servers[i] = NewServer(i, n, &wg)
		}(i)
	}
	wg.Wait()
	fmt.Println("came here")
	// Now i know for sure that all the clusters have been created.
	// Write code for connecting the peers to one another
	// If I am creating a cluster, I cannot return a disconnected cluster, so all code about connecting the peers to one another
	// must go in this file itself.
	cluster.connect(n)
}
