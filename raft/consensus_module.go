package raft

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type ConsensusModule struct {
	mu sync.Mutex
	wg sync.WaitGroup

	server        *Server
	currentTerm   int
	votedFor      interface{}
	votesInFavour int // not sure if this should be a part of the state

	/* Relevant to only leader */
	ticker *time.Ticker
	done   chan bool
}

type RequestVoteArgs struct {
	From        string
	CandidateId string // this will be changed to args.From
	Term        int
}
type RequestVoteReply struct {
	Granted bool
	Term    int
}
type AppendEntriesArgs struct {
	From string
	To   string
	Term int
}
type AppendEntriesReply struct {
	From    string
	To      string
	Term    int
	Success bool // false when peer responds negatively to leader's heartbeat (got a new leader)
}

// Methods
func (cm *ConsensusModule) log(format string, args ...interface{}) {
	format = fmt.Sprintf("[ %v ]\t %v \t", cm.server.id, cm.server.state) + format
	log.Printf(format, args...)
}

// procedures

// expect cm.mu to be locked already
func (cm *ConsensusModule) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	// cm.log("Received RequestVote RPC from %v", args.CandidateId) // this will be changed to args.From

	if cm.server.state == CANDIDATE {
		// If the receiver of the RequestVote RPC is a candidate,

		// cm.log("args.Term %v, cm.currentTerm %v", args.Term, cm.currentTerm) // debugging why candidate voted for competing candidate

		if args.Term > cm.currentTerm {
			/*
				If the candidate requesting the vote has a term greater than the current term of the receiving candidate, then
				1. Update currentTerm to the latest known currentTerm
				2. Step down from candidate state and become a follower
				3. Grant vote to the candidate (requester)
			*/
			cm.currentTerm = args.Term
			cm.becomeFollower()
			cm.votedFor = args.CandidateId // this will be changed to args.From
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
			cm.becomeFollower()
			cm.votedFor = args.CandidateId // this will be changed to args.From
			// prepare the reply
			reply.Granted = true
			reply.Term = cm.currentTerm

			// TODO
			// make sure to terminate the heartbeat sending goroutine
			cm.done <- true
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

// expect cm.mu to be locked already
func (cm *ConsensusModule) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	reply.From = cm.server.id
	reply.To = args.From

	if args.Term < cm.currentTerm {
		reply.Term = cm.currentTerm
		reply.Success = false
		cm.log("Received heartbeat from leader %v", args.From)

	} else {
		reply.Term = args.Term
		reply.Success = true
		start, end := GetTimeoutRange()
		interval := start + rand.Intn(end) // [start, start + end)

		cm.server.ticker.Reset(time.Duration(interval) * time.Second)
		cm.log("Received heartbeat from leader %v, timeout reset to %v seconds", args.From, interval)
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
