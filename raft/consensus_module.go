package raft

import (
	"fmt"
)

type ConsensusModule struct {
	server      *Server
	currentTerm int
}

type RequestVoteArgs struct {
	CandidateId string
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
		CandidateId: cm.server.id,
	}
	reply := RequestVoteReply{}

	fmt.Printf("Server %v sending RequestVote RPC\n", cm.server.id)

	for peerId, peerClient := range cm.server.peerClients {
		fmt.Printf("\tTo %v\n", peerId)
		err := peerClient.Call("ConsensusModule.RequestVote", args, &reply)
		if err != nil {
			fmt.Println("Error occurred while sending Request Vote RPC")
			fmt.Println(err)
		}
	}
}

// procedures

func (cm *ConsensusModule) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	fmt.Printf("Received RequestVote RPC from %v\n", args.CandidateId)
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
