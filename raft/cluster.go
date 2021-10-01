package raft

type Cluster struct {
	servers []*Server
}

func CreateCluster(n int) {
	cluster := &Cluster{}
	cluster.servers = make([]*Server, n)

	// iterate n times, create n servers, a
	for i := 0; i < n; i++ {
		cluster.servers[i] = CreateNewServer()
	}
}
