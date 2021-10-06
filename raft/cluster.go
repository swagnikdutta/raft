package raft

import (
	"fmt"
	"sync"
)

type Cluster struct {
	servers []*Server
}

// methods
func (c *Cluster) getServerFromId(id int) *Server {
	var s *Server

	for i := 0; i < len(c.servers); i++ {
		if c.servers[i].id == id {
			s = c.servers[i]
			break
		}
	}
	return s
}

func (c *Cluster) connect(n int) {
	for i := 0; i < n; i++ {
		var peers []*Server

		for id := 0; id < len(c.servers[i].peers); id++ {
			peer := c.getServerFromId(id)
			peers = append(peers, peer)
		}
		c.servers[i].ConnectToPeers(peers)
	}
}

// Functions

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
