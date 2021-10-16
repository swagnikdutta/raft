package raft

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
)

type ConsensusModule struct {
	mu sync.Mutex
	wg sync.WaitGroup

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
		cm.startElections() // should I wait here? why wait for election to be over?
	} else if nextState == string(LEADER) {
		cm.log("Becoming a leader")
		cm.doLeaderThings()
	}
}

func (cm *ConsensusModule) doLeaderThings() {
	cm.log("Sending heartbeats")
	cm.mu.Lock()
	defer cm.mu.Unlock()
}

func (cm *ConsensusModule) startElections() {
	cm.log("Starting elections")
	cm.mu.Lock()
	// defer cm.mu.Unlock() // this resulted in a lot of problems due to race conditions IMO

	cm.currentTerm += 1
	cm.votedFor = cm.server.id
	cm.votesInFavour += 1

	args := RequestVoteArgs{
		CandidateId: cm.server.id,
		Term:        cm.currentTerm,
	}
	reply := RequestVoteReply{
		Response: 0,
	}

	for peerId, peerClient := range cm.server.peerClients {
		cm.wg.Add(1)

		go func(peerId string, peerClient *rpc.Client, wg *sync.WaitGroup) {
			defer cm.wg.Done()
			cm.log("Sending RequestVote RPC to %v", peerId)
			// this can be made prettier
			err := peerClient.Call("ConsensusModule.RequestVote", args, &reply)
			if err != nil {
				cm.log("Error happened while sending RequestVote RPC. Error: %+v", err)
			}
			cm.log("Response from peer %v is %v. Total votes: %v", peerId, reply.Response, cm.votesInFavour)

			if reply.Response == 1 {
				cm.votesInFavour += 1 // use mutex. But how?
			}
		}(peerId, peerClient, &cm.wg)
	}

	cm.wg.Wait() // symbolises that the goroutines are done executing
	cm.mu.Unlock()

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
