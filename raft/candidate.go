package raft

import (
	"net/rpc"
	"sync"
)

func (cm *ConsensusModule) requestVote(peerId string, peerClient *rpc.Client, wg *sync.WaitGroup, args RequestVoteArgs, reply *RequestVoteReply) {
	defer cm.wg.Done()
	cm.log("Sending RequestVote RPC to peer %v", peerId)

	if err := peerClient.Call("ConsensusModule.RequestVote", args, &reply); err == nil {
		result := If(reply.Granted).String("granted", "denied")

		if reply.Granted == true {
			cm.votesInFavour += 1
		} else {
			/*
				If the candidate didn't get the vote, chances are it's currentTerm is out of date.
				Scenario 1: peer that denied the vote is a candidate with a higher term
				Scenario 2: peer that denied the vote is a follower with a higher term
				Scenario 3: peer that denied the vote is a follower which granted vote to a competing candidate i.e this candidate and the competiting candidate are running elections for the same term (currentTerm)
				In either case, the only actionable thing here is to update the currentTerm and step down to be a follower
			*/
			if reply.Term > cm.currentTerm {
				cm.currentTerm = reply.Term
				// TODO: this shouldn't be the scope of the current thread. We need to make these things run on a separate thread to handle mutexes gracefully.
				cm.becomeFollower()
			}
		}
		cm.log("Peer %v %v vote. Total votesInFavour = %v", peerId, result, cm.votesInFavour)
	}
}

func (cm *ConsensusModule) becomeCandidate() {
	cm.mu.Lock()

	cm.server.state = CANDIDATE
	cm.log("Became a candidate, starting elections ...")
	cm.currentTerm += 1
	cm.votedFor = cm.server.id
	cm.votesInFavour = 1
	cm.log("currentTerm = %v, votesInFavour = %v", cm.currentTerm, cm.votesInFavour)

	args := RequestVoteArgs{
		From: cm.server.id,
		Term: cm.currentTerm,
	}
	reply := RequestVoteReply{}

	for peerId, peerClient := range cm.server.peerClients {
		cm.wg.Add(1)
		go cm.requestVote(peerId, peerClient, &cm.wg, args, &reply)
	}
	cm.wg.Wait()

	/* this calculation must happen before mutex unlock */
	hasMajorityVotes := 2*cm.votesInFavour > (len(cm.server.peerIds) + 1)
	// TODO: maybe the unlocking can be deferred when ChangeState happens in a different thread.
	cm.mu.Unlock()

	if hasMajorityVotes {
		cm.log("Won election with majority votes")
		// TODO: make this run in a separate thread
		cm.becomeLeader()
	}
}
