package raft

import (
	"fmt"
)

type ConsensusModule struct {
	server      *Server
	currentTerm int
	votedFor    interface{}
}

type RequestVoteArgs struct {
	CandidateId string
}
type RequestVoteReply struct {
	Response int
}

// Methods
func (cm *ConsensusModule) ChangeState(nextState string) {
	cm.server.state = nextState

	if nextState == CANDIDATE {
		fmt.Printf("Server %v became a Candidate\n", cm.server.id)
		cm.currentTerm += 1
		cm.votedFor = cm.server.id
		cm.startElections()
	} else if nextState == string(LEADER) {
		fmt.Printf("Server %v became a Leader\n", cm.server.id)
	}
}

func (cm *ConsensusModule) startElections() {
	fmt.Printf("Server %v is starting elections\n", cm.server.id)
	args := RequestVoteArgs{
		CandidateId: cm.server.id,
	}
	reply := RequestVoteReply{
		Response: 0,
	}
	votesInFavour := 1 // it voted for itself

	fmt.Printf("Server %v sending RequestVote RPC\n", cm.server.id)

	for peerId, peerClient := range cm.server.peerClients {
		_ = peerId
		err := peerClient.Call("ConsensusModule.RequestVote", args, &reply)
		if err != nil {
			fmt.Println("Error occurred while sending Request Vote RPC")
			fmt.Println(err)
		}
		fmt.Printf("Response from peer %v is %v\n", peerId, reply.Response)

		if reply.Response == 1 {
			votesInFavour += 1
		}
	}

	if hasMajorityVotes(votesInFavour, cm) {
		fmt.Printf("Server %v will become leader\n", cm.server.id)
		cm.ChangeState(string(LEADER))
	}
}

// procedures

func (cm *ConsensusModule) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	fmt.Printf("Server %v received RequestVote RPC from %v\n", cm.server.id, args.CandidateId)

	// this is where I need a mutex, so that it doesn't vote 2 members
	// add additional checks here, if the term is valid.
	if cm.votedFor == nil {
		cm.votedFor = cm.server.id
		reply.Response = 1
	}
	return nil
}

// Functions

func hasMajorityVotes(votesInFavour int, cm *ConsensusModule) bool {
	return 2*votesInFavour > (len(cm.server.peerIds) + 1)
}

func NewConsensusModule(server *Server) *ConsensusModule {
	cm := &ConsensusModule{
		currentTerm: 1,
		server:      server,
		votedFor:    nil,
	}
	return cm
}
