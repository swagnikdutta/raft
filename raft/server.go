package raft

import (
	"fmt"
	"log"
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

	peerClients map[string]*rpc.Client

	wg sync.WaitGroup
}

// Methods

func (s *Server) ConnectToPeers(peers []*Server) {
	for i := 0; i < len(peers); i++ {
		peer := peers[i]

		if s.peerClients[peer.id] == nil {
			listenerAddress := peer.listener.Addr()
			fmt.Println("Connecting with peer server")
			client, err := rpc.Dial(listenerAddress.Network(), listenerAddress.String())

			if err != nil {
				fmt.Println("Error connecting to peer", err)
			}
			s.peerClients[peer.id] = client
		}
	}
}

// Functions

func NewServer(serverCount int, wg *sync.WaitGroup) *Server {
	server := new(Server)
	server.id = uuid.New().String()
	server.state = FOLLOWER
	server.peerClients = make(map[string]*rpc.Client)
	server.CM = &ConsensusModule{
		currentTerm: 1,
	}

	// Attaching listener
	var err error
	server.listener, err = net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Error while attaching listener on the server")
		log.Fatal(err)
	}

	// Start listening for incoming connections
	server.wg.Add(1)
	go func() {
		defer server.wg.Done()

		for {
			conn, err := server.listener.Accept() // blocks
			fmt.Println("Got a connection", conn)
			if err != nil {
				fmt.Println("Error listening for incoming connections")
				log.Fatal(err)
			}
			// server.wg.Add(1)
			// go func() {
			// 	defer server.wg.Done()
			// 	server.rpcServer.ServeConn(conn)
			// }()
		}
	}()

	return server
}
