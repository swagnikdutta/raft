package raft

import (
	"fmt"
	"log"
)

type ConsensusModule struct {
	server        *Server
	currentTerm   int
	votedFor      interface{}
	votesInFavour int // not sure if this should be a part of the state
}

type RequestVoteArgs struct {
	CandidateId string
	Term        int
}
type RequestVoteReply struct {
	Response int
}

// Methods
func (cm *ConsensusModule) log(format string, args ...interface{}) {
	format = fmt.Sprintf("[ %v ] ", cm.server.id) + format
	log.Printf(format, args...)
}

func (cm *ConsensusModule) ChangeState(nextState string) {
	cm.server.state = nextState

	if nextState == CANDIDATE {
		cm.log("Becoming a candidate")
		cm.currentTerm += 1        // use mutex
		cm.votedFor = cm.server.id // use mutex
		cm.startElections()        // should I wait here? why wait for election to be over?
	} else if nextState == string(LEADER) {
		cm.log("Becoming a leader")
	}
}

func (cm *ConsensusModule) startElections() {
	cm.log("Starting elections")
	args := RequestVoteArgs{
		CandidateId: cm.server.id,
		Term:        cm.currentTerm,
	}
	reply := RequestVoteReply{
		Response: 0,
	}
	cm.votesInFavour += 1 // use mutex

	// this cannot be synchronous, cannot wait for the reply of one peer before requesting another peer
	for peerId, peerClient := range cm.server.peerClients {
		cm.log("Sending RequestVote RPC to %v", peerId)
		err := peerClient.Call("ConsensusModule.RequestVote", args, &reply)
		if err != nil {
			cm.log("Error happened while sending RequestVote RPC. Error: %+v", err)
		}
		cm.log("Response from peer %v is %v", peerId, reply.Response)

		if reply.Response == 1 {
			cm.votesInFavour += 1 // use mutex
		}
	}

	if hasMajorityVotes(cm) {
		cm.log("Will become leader")
		cm.ChangeState(string(LEADER)) // again, should I wait here? or make it run in a separate goroutine
	}
}

// procedures

func (cm *ConsensusModule) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	cm.log("Received RequestVote RPC from %v", args.CandidateId)

	// this is where I need a mutex, so that it doesn't vote 2 members
	// add additional checks here, if the term is valid.
	if cm.votedFor == nil {
		cm.votedFor = cm.server.id
		reply.Response = 1
	}
	return nil
}

// Functions

func hasMajorityVotes(cm *ConsensusModule) bool {
	return 2*cm.votesInFavour > (len(cm.server.peerIds) + 1)
}

func NewConsensusModule(server *Server) *ConsensusModule {
	cm := &ConsensusModule{
		currentTerm:   1,
		server:        server,
		votedFor:      nil,
		votesInFavour: 0,
	}
	return cm
}
