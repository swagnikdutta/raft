package raft

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Server struct {
	id    int
	peers []int
	CM    *ConsensusModule
	timer *time.Timer
	state string
}

func randomizedTimeout(serverId int) time.Duration {
	interval := 5 + rand.Intn(6) // interval 5-10 seconds
	fmt.Printf("Timeout set for server %d is %d seconds\n", serverId, interval)
	return time.Duration(interval)
}

func handleElectionTimeout(server *Server, wg *sync.WaitGroup) {
	defer wg.Done()
	<-server.timer.C
	fmt.Println("Timeout happended for server: ", server.id)
	HandleStateTransition(server, CANDIDATE)
}

func NewServer(serverId, serverCount int, wg *sync.WaitGroup) *Server {
	server := new(Server)
	server.id = serverId
	server.state = FOLLOWER
	for i := 0; i < serverCount; i++ {
		if i != serverId {
			server.peers = append(server.peers, i)
		}
	}
	server.CM = &ConsensusModule{
		currentTerm: 1,
	}
	server.timer = time.NewTimer(randomizedTimeout(server.id) * time.Second)
	go handleElectionTimeout(server, wg)
	return server
}
