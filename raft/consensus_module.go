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
	votesInFavour int

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

// Procedures

// expect cm.mu to be locked already
func (cm *ConsensusModule) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {

	if cm.server.state == CANDIDATE {
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
		if args.Term > cm.currentTerm {
			/*
				If follower's currentTerm is lesser than the currentTerm of the candidate requesting vote
				1. Update the currentTerm
				2. Grant vote to candidate (requester)
			*/
			cm.currentTerm = args.Term
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
	} else if cm.server.state == LEADER {
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

// cm.mu is locked ---- NOOOOOOOO got error -> fatal error: sync: unlock of unlocked mutex
func (cm *ConsensusModule) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	cm.log("Received heartbeat from leader %v", args.From)

	reply.From = cm.server.id
	reply.To = args.From
	/*
		The receiver of a heartbeat can be
		1) Follower
		2) Candidate
		3) Obsolete leader - A leader who got disconnected initially, but
			came back to life later and now thinks of itself as the legitimate leader
	*/
	if args.Term < cm.currentTerm {
		/*
			If the leader sending the heartbeat has an obsolete term,
			reject the heartbeat right away, and let the sender know of the updated term (known to this server).
		*/
		reply.Term = cm.currentTerm
		reply.Success = false
		cm.log("Received heartbeat from leader %v", args.From)
	} else {
		/*
			If the term of the leader sending heartbeat is >= the term known to this server
			1) update currentTerm - receiver can be leader(obsolete), candidate, follower
			2) If the receiver of the heartbeat is a candidate or an obsolete leader, step down to be follower
			3) reset the timer
			TODO: // check if there's any definite pattern to do the above
		*/

		if args.Term > cm.currentTerm {
			/*
				TLDR: First heartbeat?

				On receiving the first heartbeat, the currentTerm of the receiver will be synced with that of the leader.
				Further heartbeats won't change the value of the receiver's currentTerm as it is in sync with that of the leader, for the entirety of this term.
				We can prevent unnecessary writes by putting the assignment inside this block
			*/
			cm.currentTerm = args.Term
			cm.log("Updated currentTerm to %v", args.Term)
		}

		// TODO: put this in a separate function, make it a one line thing
		start, end := GetTimeoutRange()
		interval := start + rand.Intn(end) // [start, start + end)
		cm.server.ticker.Reset(time.Duration(interval) * time.Second)
		cm.log("Timeout reset to %v seconds", interval)

		if cm.server.state == CANDIDATE || cm.server.state == LEADER {
			cm.log("Stepping down as %v, becoming a follower", cm.server.state)
			// TODO: need serious debugging
			// cm.mu.Unlock() // because the becomeFollower function locks the mutex, that was the idea, but uncommenting this line gives error
			cm.becomeFollower()
		}

		reply.Term = args.Term
		reply.Success = true
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
