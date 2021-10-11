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
	cm      *ConsensusModule
	timer   *time.Timer
	state   string

	listener    net.Listener           // to listen for incoming connections
	rpcServer   *rpc.Server            // to server incoming connections
	peerClients map[string]*rpc.Client // a client object for each peer server it wants to communicate with

	wg sync.WaitGroup
}

// Methods

func (s *Server) ConnectToPeers(peerServers []*Server) {
	for i := 0; i < len(peerServers); i++ {
		peerServer := peerServers[i]

		if s.peerClients[peerServer.id] == nil {
			listenerAddress := peerServer.listener.Addr()
			client, err := rpc.Dial(listenerAddress.Network(), listenerAddress.String())

			if err != nil {
				log.Fatal(err)
			}
			s.peerClients[peerServer.id] = client
		}
	}
}

func (s *Server) SetTimer(wg *sync.WaitGroup) {
	s.timer = time.NewTimer(randomizedTimeout(s.id) * time.Second)
	go s.HandleElectionTimeout(wg)
}

func (s *Server) HandleElectionTimeout(wg *sync.WaitGroup) {
	defer wg.Done()
	<-s.timer.C

	fmt.Println("Timeout happended for server: ", s.id)
	s.cm.ChangeState(CANDIDATE)
}

// Functions

func randomizedTimeout(serverId string) time.Duration {
	interval := 5 + rand.Intn(6) // interval 5-10 seconds
	fmt.Printf("Timeout set for server %v is %v seconds\n", serverId, interval)
	return time.Duration(interval)
}

func generateNewId() string {
	return uuid.New().String()
}

func NewServer(serverCount int, wg *sync.WaitGroup) *Server {
	var err error

	server := new(Server)
	server.id = generateNewId()
	server.state = FOLLOWER

	server.peerClients = make(map[string]*rpc.Client)
	server.rpcServer = rpc.NewServer()
	// server.rpcServer.Register(server.cm)

	// Attaching listener
	server.listener, err = net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	// Start listening for incoming connections
	server.wg.Add(1)
	go func() {
		defer server.wg.Done()

		for {
			conn, err := server.listener.Accept() // TODO: close the connection at some point

			if err != nil {
				log.Fatal(err)
			}
			// handle the connection in a new goroutine and go back to accepting new incoming connections.
			server.wg.Add(1)
			go func(c net.Conn) {
				defer server.wg.Done()
				server.rpcServer.ServeConn(c) // TODO: shut this down at some point
			}(conn)
		}
	}()

	server.cm = NewConsensusModule(server)
	return server
}
