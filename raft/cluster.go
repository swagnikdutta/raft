package raft

import (
	"fmt"
	"sync"
)

type Cluster struct {
	servers []*Server
}

// Methods
func (c *Cluster) FindServerById(id string) *Server {
	var s *Server

	for i := 0; i < len(c.servers); i++ {
		if c.servers[i].id == id {
			s = c.servers[i]
			break
		}
	}
	return s
}

func (c *Cluster) populatePeers(n int) {
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if c.servers[i].id != c.servers[j].id {
				c.servers[i].peerIds = append(c.servers[i].peerIds, c.servers[j].id)
			}
		}
	}
}

func (c *Cluster) connect(n int) {
	for i := 0; i < n; i++ {
		var peerServers []*Server
		server := c.servers[i]

		for j := 0; j < len(server.peerIds); j++ {
			peerId := server.peerIds[j]
			peerServers = append(peerServers, c.FindServerById(peerId))
		}
		c.servers[i].ConnectToPeers(peerServers)
	}
	fmt.Println("All peers connected!")
}

func (c *Cluster) runTimerOnServers(n int) {
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < len(c.servers); i++ {
		go c.servers[i].RunElectionTimer(&wg) // validate this once
	}
	// wg.Wait() This will block
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
			cluster.servers[i] = NewServer(n, &wg, cluster) // trying to make cm call a method of cluster
		}(i)
	}
	wg.Wait()
	cluster.populatePeers(n)
	cluster.connect(n)
	cluster.runTimerOnServers(n)
}
