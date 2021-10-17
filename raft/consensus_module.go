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
	Granted bool
}

// Methods
func (cm *ConsensusModule) log(format string, args ...interface{}) {
	format = fmt.Sprintf("[ %v ] ", cm.server.id) + format
	log.Printf(format, args...)
}

func (cm *ConsensusModule) ChangeState(nextState string) {
	cm.mu.Lock()
	cm.server.state = nextState

	if nextState == CANDIDATE {
		cm.log("Becoming a candidate")
		cm.startElections()
	} else if nextState == string(LEADER) {
		cm.log("Becoming a leader")
		cm.doLeaderThings()
	}
}

// expect cm.mu to be locked
func (cm *ConsensusModule) doLeaderThings() {
	cm.log("Sending heartbeats")
	defer cm.mu.Unlock()
	return
}

// expect cm.mu to be locked
func (cm *ConsensusModule) startElections() {
	cm.currentTerm += 1
	cm.votedFor = cm.server.id
	cm.votesInFavour += 1
	cm.log("is starting elections, votes for itself. Vote count so far %v", cm.votesInFavour)

	args := RequestVoteArgs{
		CandidateId: cm.server.id,
		Term:        cm.currentTerm,
	}
	reply := RequestVoteReply{
		Granted: false,
	}

	for peerId, peerClient := range cm.server.peerClients {
		cm.wg.Add(1)

		go func(peerId string, peerClient *rpc.Client, wg *sync.WaitGroup) {
			defer cm.wg.Done()
			cm.log("Sending RequestVote RPC to %v", peerId)

			if err := peerClient.Call("ConsensusModule.RequestVote", args, &reply); err == nil {
				if reply.Granted == true {
					cm.votesInFavour += 1
				}
				cm.log("Response from peer %v is %v. Total votes: %v", peerId, reply.Granted, cm.votesInFavour)
			}
		}(peerId, peerClient, &cm.wg)
	}

	cm.wg.Wait()
	hasMajorityVotes := 2*cm.votesInFavour > (len(cm.server.peerIds) + 1)
	// Reasons for unlocking here
	// 1) Need the goroutines sending RequestVote RPC to complete so that vote can be safely counted
	// 2) To maintain code consistency in change state function - starting with locking of mutex for both cases - leader, candidate
	cm.mu.Unlock()

	if hasMajorityVotes {
		cm.log("Will become leader")
		cm.ChangeState(string(LEADER)) // should this be a sequential operation? ChangeState leads to different workflow altogether - candidate election or leader sending heartbeats
	}
}

// procedures

// expect cm.mu to be locked already
func (cm *ConsensusModule) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	cm.log("Received RequestVote RPC from %v", args.CandidateId)

	// add additional checks here, if the term is valid.
	if cm.votedFor == nil && cm.votedFor != cm.server.id {
		cm.votedFor = cm.server.id
		reply.Granted = true
	}
	return nil
}

// Functions

func NewConsensusModule(server *Server) *ConsensusModule {
	cm := &ConsensusModule{
		currentTerm:   1,
		server:        server,
		votedFor:      nil,
		votesInFavour: 0,
	}
	return cm
}
