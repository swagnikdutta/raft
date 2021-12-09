package raft

import (
	"fmt"
	"sync"
)

type Cluster struct {
	servers []*Server
}

// Methods
func (c *Cluster) findServerById(id string) *Server {
	var s *Server

	for _, server := range c.servers {
		if server.id == id {
			s = server
			break
		}
	}
	return s
}

func (c *Cluster) populatePeerInfo(n int) {
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			c.servers[i].peerIds = append(c.servers[i].peerIds, c.servers[j].id)
			c.servers[j].peerIds = append(c.servers[j].peerIds, c.servers[i].id)
		}
	}
}

func (c *Cluster) connectAllServers(n int) {
	for i := 0; i < n; i++ {
		var peerServers []*Server

		for _, peerId := range c.servers[i].peerIds {
			peerServers = append(peerServers, c.findServerById(peerId))
		}
		c.servers[i].ConnectToPeers(peerServers)
	}
	fmt.Println("All peers connected!")
}

func (c *Cluster) runTimerOnServers(n int) {
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < len(c.servers); i++ {
		go c.servers[i].StartElectionTimer(&wg) // validate this once
	}
	wg.Wait()
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
			cluster.servers[i] = NewServer(n, &wg) // trying to make cm call a method of cluster
		}(i)
	}
	wg.Wait()
	cluster.populatePeerInfo(n)
	cluster.connectAllServers(n)
	cluster.runTimerOnServers(n)
}
