package raft

import (
	"math/rand"
	"time"
)

type Server struct {
	id    int
	peers []int
	CM    *ConsensusModule
	timer *time.Timer
}

func randomizedTimeout() time.Duration {
	return time.Duration(2 + rand.Intn(6))
}

func NewServer(serverId, serverCount int) *Server {
	server := new(Server)
	server.id = serverId
	for i := 0; i < serverCount; i++ {
		if i != serverId {
			server.peers = append(server.peers, i)
		}
	}
	server.CM = &ConsensusModule{}
	server.timer = time.NewTimer(randomizedTimeout() * time.Second)
	return server
}
