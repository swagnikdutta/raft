package raft

import (
	"sync"
)

type Cluster struct {
	servers []*Server
}

func CreateCluster(n int) {
	cluster := &Cluster{}
	cluster.servers = make([]*Server, n)

	// iterate n times, create n servers, a
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			cluster.servers[i] = NewServer(i, n, &wg)
		}(i)
	}
	wg.Wait()
}
