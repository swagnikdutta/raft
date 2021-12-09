package raft

import (
	"fmt"
	"sync"
)

type Cluster struct {
	servers []*Server
	wg      sync.WaitGroup
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
	c.wg.Add(n)
	for _, server := range c.servers {
		go server.StartElectionTimer(&c.wg) // validate this once
	}
	c.wg.Wait()
}

// Functions

func CreateCluster(n int) {
	cluster := &Cluster{}
	cluster.servers = make([]*Server, n)

	// iterate n times, create n servers, a

	for i := 0; i < n; i++ {
		cluster.wg.Add(1)
		go func(i int) {
			defer cluster.wg.Done()
			cluster.servers[i] = NewServer(n, &cluster.wg) // trying to make cm call a method of cluster
		}(i)
	}
	cluster.wg.Wait()
	cluster.populatePeerInfo(n)
	cluster.connectAllServers(n)
	cluster.runTimerOnServers(n)
}
