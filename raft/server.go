package raft

import (
	"fmt"
	"net/rpc"
	"sync"
)

type Entry struct {
	term    int    // the term when the entry was created
	command string // command for the state machine
}

type Log struct {
	logs []Entry
}

type persistent struct {
	log         Log
	currentTerm int
	votedFor    interface{}
}

type volatile struct {
	commitIndex int
	lastApplied interface{}
}

// each server should have an id, so that we can populate the peer array
// the paper talks about candidate id
type Server struct {
	state      string
	persistent persistent
	volatile   volatile
	// peers      []Server
}

func CreateNewServer(wg *sync.WaitGroup) {
	server := &Server{
		state: "FOLLOWER",
		persistent: persistent{
			log:         Log{},
			currentTerm: 0,   // latest term server has seen (initialized to 0 on first boot)
			votedFor:    nil, // candidate id that received vote in current term
		},
		volatile: volatile{
			commitIndex: 0,   // index of the highest log entry known to be committed
			lastApplied: nil, // index of the highest log enty applied to state machine
		},
	}

	rpcserver := rpc.NewServer()
	rpcserver.Register(server)
	// now all the methods of type server will become procedures automatically

	doSomething()
	defer wg.Done()
	fmt.Println("creating a new server")
}

// methods to write. Leaders should send AppendRPC,
// candidates will send RequestVoteRPC to all servers
