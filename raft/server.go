package raft

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Server struct {
	id      string
	peerIds []string
	CM      *ConsensusModule
	timer   *time.Timer
	state   string

	listener  net.Listener
	rpcServer *rpc.Server
}

// methods
func (s *Server) ConnectToPeers(peers []*Server) {
	// Find listener address of each of the peer server and connect with them
	for i := 0; i < len(peers); i++ {
		// connect server (s.id) with server (peers[i].id) which is listening on peer[i].listener.Addr()
		// client, err := s.rpcServer.Dial(peers[i].listener.Addr().Network())
	}
}

func randomizedTimeout(serverId int) time.Duration {
	interval := 5 + rand.Intn(6) // interval 5-10 seconds
	fmt.Printf("Timeout set for server %d is %d seconds\n", serverId, interval)
	return time.Duration(interval)
}

// func handleElectionTimeout(server *Server, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	<-server.timer.C
// 	fmt.Println("Timeout happended for server: ", server.id)
// 	HandleStateTransition(server, CANDIDATE)
// }

func NewServer(serverCount int, wg *sync.WaitGroup) *Server {
	server := new(Server)
	server.id = uuid.New().String()
	server.state = FOLLOWER
	server.CM = &ConsensusModule{
		currentTerm: 1,
	}

	// Attaching RPC server
	// server.rpcServer = rpc.NewServer()
	// server.rpcServer.Register(server.CM)
	// server.rpcServer.Register(server)  // can it do that?

	// Attaching listener
	var err error
	server.listener, err = net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Error while attaching listener on the server")
		log.Fatal(err)
	}

	// Start listening for incoming connections
	go func() {
		for {
			conn, err := server.listener.Accept()
			if err != nil {
				fmt.Println("Error listening for incoming connections")
				log.Fatal(err)
			}
			// what to do with this conn object?
			_ = conn
		}
	}()

	// I think the timer should be set only when the peers are connected to one another.
	// first we must connect the peers, so that when the follower becomes a candidate, it can start the elections, i.e send RPCs immediately
	//
	// server.timer = time.NewTimer(randomizedTimeout(server.id) * time.Second)
	// go handleElectionTimeout(server, wg)
	return server
}
