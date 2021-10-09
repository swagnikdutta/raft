package raft

import (
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

	listener    net.Listener           // to listen for incoming connections
	rpcServer   *rpc.Server            // to server incoming connections
	peerClients map[string]*rpc.Client // a client object for each peer it wants to communicate with

	wg sync.WaitGroup
}

// Methods

func (s *Server) ConnectToPeers(peers []*Server) {
	for i := 0; i < len(peers); i++ {
		peer := peers[i]

		if s.peerClients[peer.id] == nil {
			listenerAddress := peer.listener.Addr()
			client, err := rpc.Dial(listenerAddress.Network(), listenerAddress.String())

			if err != nil {
				log.Fatal(err)
			}
			s.peerClients[peer.id] = client
		}
	}
}

// Functions

func NewServer(serverCount int, wg *sync.WaitGroup) *Server {
	var err error

	server := new(Server)
	server.id = uuid.New().String()
	server.state = FOLLOWER
	server.peerClients = make(map[string]*rpc.Client)
	server.CM = &ConsensusModule{
		currentTerm: 1,
	}

	server.rpcServer = rpc.NewServer()
	// server.rpcServer.Register(server.CM)

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

	return server
}
