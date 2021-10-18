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
	Term    int
}

// Methods
func (cm *ConsensusModule) log(format string, args ...interface{}) {
	format = fmt.Sprintf("[ %v ]\t %v \t", cm.server.id, cm.server.state) + format
	log.Printf(format, args...)
}

func (cm *ConsensusModule) ChangeState(nextState string) {
	cm.mu.Lock()
	// if state transition already done (unique scenario, then ignore the rest of the code, unlock the mutex and leave)
	// this scenario arise when, candidate already moved back to follower state upon receive negative ack from followers
	// but state change was triggered again due to negative acknowledgement of a competing candidate (with higher term obviously, hence the negative ack)
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
		// Term: ,
	}

	for peerId, peerClient := range cm.server.peerClients {
		cm.wg.Add(1)

		go func(peerId string, peerClient *rpc.Client, wg *sync.WaitGroup) {
			defer cm.wg.Done()
			cm.log("Sending RequestVote RPC to %v", peerId)

			if err := peerClient.Call("ConsensusModule.RequestVote", args, &reply); err == nil {
				if reply.Granted == true {
					// candidate got vote
					cm.votesInFavour += 1
				} else {
					/*
						If the candidate didn't get the vote, chances are it's currentTerm is out of date.
						Scenario 1: peer that denied the vote is a candidate with a higher term
						Scenario 2: peer that denied the vote is a follower with a higher term
						Scenario 3: peer that denied the vote is a follower which granted vote to a competing candidate i.e this candidate and the competiting candidate are running elections for the same term (currentTerm)

						The only actionable part here is for scenarios 1 and 2,
						where this candidate updates its currentTerm and steps down to be a follower
					*/
					if reply.Term > cm.currentTerm {
						cm.currentTerm = reply.Term
						cm.server.state = FOLLOWER
					}
				}
				cm.log("Response from peer %v is %v. Total votes: %v", peerId, reply.Granted, cm.votesInFavour)
			}
		}(peerId, peerClient, &cm.wg)
	}

	cm.wg.Wait()
	hasMajorityVotes := 2*cm.votesInFavour > (len(cm.server.peerIds) + 1)
	// Reasons for unlocking here
	// 1) Need the goroutines sending RequestVote RPC to complete so that votes can be safely calculated
	// 2) To maintain code consistency in ChangeState function - starting with locking of mutex for both cases - leader, candidate
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

	if cm.server.state == CANDIDATE {
		// If the receiver of the RequestVote RPC is a candidate,
		if args.Term > cm.currentTerm {
			/*
				If the candidate requesting the vote has a term greater than the current term of the receiving candidate, then
				1. Update currentTerm to the latest known currentTerm
				2. Step down from candidate state and become a follower
				3. Grant vote to the candidate (requester)
			*/
			cm.currentTerm = args.Term
			cm.server.state = FOLLOWER
			cm.votedFor = args.CandidateId
			// prepare the reply
			reply.Granted = true
			reply.Term = cm.currentTerm
		} else {
			/*
				If the currentTerm of this candidate (receiver) is higher than the currentTerm of the candidate requesting vote,
				1. Deny the vote, update the requester with the higher currentTerm it knows
			*/
			reply.Granted = false
			reply.Term = cm.currentTerm
		}
	} else if cm.server.state == FOLLOWER {
		// If the receiver of the RequestVote RPC is a follower
		if args.Term > cm.currentTerm {
			/*
				If follower's currentTerm is lesser than the currentTerm of the candidate requesting vote
				1. Update the currentTerm
				2. Grant vote to candidate (requester)
			*/
			cm.currentTerm = args.Term
			// prepare the reply
			reply.Granted = true
			reply.Term = cm.currentTerm
		} else {
			/*
				If follower's currentTerm is higher than the currentTerm of the candidate requesting vote,
				1. Deny the vote, update the candidate requesting for vote with the higher value of currentTerm known to the follower.
			*/
			reply.Granted = false
			reply.Term = cm.currentTerm
		}
	} else {
		// If the receiver of the RequestVote RPC is a leader,

		// Leader, on discovering some other node with a higher term number, steps down and becomes a follower
		if args.Term > cm.currentTerm {
			/*
				If the candidate requesting the vote has a term greater than the current term of the leader, then
				1. Update currentTerm to the latest known currentTerm
				2. Step down from leadership and become a follower
				3. Grant vote to the candidate (requester)
			*/
			cm.currentTerm = args.Term
			cm.server.state = FOLLOWER
			cm.votedFor = args.CandidateId
			// prepare the reply
			reply.Granted = true
			reply.Term = cm.currentTerm
		} else {
			/*
				If the currentTerm of the leader is higher than the currentTerm of the candidate requesting vote,
				1. Deny the vote, update the candidate with the higher value of currentTerm known to the leader
			*/
			reply.Term = cm.currentTerm
			reply.Granted = false
		}
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
