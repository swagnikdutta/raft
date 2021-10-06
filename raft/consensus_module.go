package raft

import "fmt"

type ConsensusModule struct {
	currentTerm int
}

// procedures
type ArgType struct{}
type ReplyType struct{}

func (cm *ConsensusModule) DoSomething(ArgType, *ReplyType) error {
	return nil
}

func StartElections(server *Server) {
	fmt.Printf("Server with id: %v becomes a Candidate and starting elections\n", server.id)
	server.CM.currentTerm += 1

	// Next steps
	// the candidate will be sending request vote RPCs to it's peers
	// check if it has received the majority of votes
	// but first all peers need to be connected to one another, they need to listen for rpc calls.
}
