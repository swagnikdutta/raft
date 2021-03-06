package raft

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

var uniq int = 100

type Server struct {
	id      string
	peerIds []string
	cm      *ConsensusModule
	timer   *time.Timer
	state   string
	cluster *Cluster

	listener    net.Listener           // to listen for incoming connections
	rpcServer   *rpc.Server            // to server incoming connections
	peerClients map[string]*rpc.Client // a client object for each peer server it wants to communicate with

	wg sync.WaitGroup

	ticker *time.Ticker
	done   chan bool

	logs *os.File
}

// Methods

func (s *Server) log(format string, args ...interface{}) {
	format = fmt.Sprintf("[ %v ]\t %v \t", s.id, s.state) + format + "\n"
	log.Printf(format, args...)
	fmt.Fprintf(s.logs, format, args...)
}

func (s *Server) ConnectToPeers(peerServers []*Server) {
	for _, peerServer := range peerServers {
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

func (s *Server) setTimer() {
	start, end := GetTimeoutRange()
	interval := start + rand.Intn(end)
	s.ticker = time.NewTicker(time.Duration(interval) * time.Second)
	s.log("Timeout in %v seconds", interval)
}

func (s *Server) StartElectionTimer(wg *sync.WaitGroup) {
	defer wg.Done()
	s.setTimer()

	for {
		select {
		case <-s.done:
			s.ticker.Stop()
			return
		case <-s.ticker.C:
			s.log("Timeout!")
			// TODO: state change should happen in a separate thread
			if s.state == FOLLOWER {
				s.cm.becomeCandidate()
			} else if s.state == CANDIDATE {
				// From the visualization, if the candidate times out, it increases its term and stays as a candidate.
				// The time interval for Request vote RPC is reset as well (observed this).
				// TODO: This part is not handled yet - because I never let this scenario happen.
			} else if s.state == LEADER {
				// Leader has no timer on it. This is not handled I think.
				// TODO: remove timer when a node becomes a leader, maybe use the done channel. Once done, remove this else block.
			}
			s.setTimer()
		}
	}
}

// Functions

func generateNewId() string {
	uniq += 1
	return strconv.Itoa(uniq)
	// return uuid.New().String()
}

func NewServer(serverCount int, wg *sync.WaitGroup) *Server {
	var err error

	server := new(Server)
	server.id = generateNewId()
	server.state = FOLLOWER
	server.logs, _ = os.Create(server.id + ".txt")

	server.peerClients = make(map[string]*rpc.Client)
	server.rpcServer = rpc.NewServer()

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
			// handling the connection in a new goroutine and go back to accepting new incoming connections.
			server.wg.Add(1)
			go func(c net.Conn) {
				defer server.wg.Done()
				server.rpcServer.ServeConn(c) // TODO: shut this down at some point
			}(conn)
		}
	}()

	server.cm = NewConsensusModule(server)
	server.rpcServer.Register(server.cm)
	return server
}
