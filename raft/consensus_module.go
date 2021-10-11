package raft

import (
	"fmt"
)

type ConsensusModule struct {
	server      *Server
	currentTerm int
}

type RequestVoteArgs struct {
	candidateId string
}
type RequestVoteReply struct{}

// Methods
func (cm *ConsensusModule) ChangeState(nextState string) {
	cm.server.state = nextState
	fmt.Printf("Server %v became a Candidate\n", cm.server.id)

	// What does a server do when it becomes a candidate?
	// it increases the current term and starts elections
	cm.currentTerm += 1
	cm.startElections()
}

func (cm *ConsensusModule) startElections() {
	fmt.Printf("Server %v is starting elections\n", cm.server.id)
	args := RequestVoteArgs{
		candidateId: cm.server.id,
	}
	reply := RequestVoteReply{}

	for _, peerClient := range cm.server.peerClients {
		fmt.Printf("%v\n", peerClient)
		// Sending Request vote RPC to each peer client
		peerClient.Call("ConsensusModule.RequestVote", args, reply)
	}
}

// procedures

func (cm *ConsensusModule) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	fmt.Printf("Server %v is about to send Request RPC votes\n", cm.server.id)
	return nil
}

// Functions

func NewConsensusModule(server *Server) *ConsensusModule {
	cm := &ConsensusModule{
		currentTerm: 1,
		server:      server,
	}
	return cm
}
