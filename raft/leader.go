package raft

import (
	"net/rpc"
	"sync"
	"time"
)

func (cm *ConsensusModule) becomeLeader() {
	cm.mu.Lock()

	cm.server.state = LEADER
	cm.log("Became a leader, sending heartbeats")

	// this is the first heartbeat after becoming the leader, not sure if I should wait for it to finish
	cm.wg.Add(1) // I am not sure if I need this
	go cm.sendHeartbeats(false)
	cm.wg.Wait() // I am not sure if I need this

	cm.wg.Add(1)
	go cm.sendHeartbeats(true)
	cm.wg.Wait()
	cm.mu.Unlock()
}

func (cm *ConsensusModule) sendHeartbeatToPeer(peerId string, peerClient *rpc.Client, wg *sync.WaitGroup, args AppendEntriesArgs, reply *AppendEntriesReply) {
	defer cm.wg.Done()

	cm.log("Sending heartbeat to %v", peerId)
	args.To = peerId

	if err := peerClient.Call("ConsensusModule.AppendEntries", args, reply); err == nil {
		if reply.Success {
			cm.log("Still the leader, yayyy!")
		} else {
			cm.currentTerm = reply.Term
			cm.becomeFollower()
			cm.done <- true
		}
		// cm.log("Heartbeat response from peer %v is %v ", peerId, reply.Success)
	}
}

func (cm *ConsensusModule) sendHeartbeats(periodic bool) {
	defer cm.wg.Done()

	args := AppendEntriesArgs{
		From: cm.server.id,
		Term: cm.currentTerm,
	}
	reply := AppendEntriesReply{}
	if !periodic {
		for peerId, peerClient := range cm.server.peerClients {
			cm.wg.Add(1)
			go cm.sendHeartbeatToPeer(peerId, peerClient, &cm.wg, args, &reply)
		}
		return
	}
	cm.ticker = time.NewTicker(2 * time.Second)
	for {
		select {
		case <-cm.done:
			cm.ticker.Stop()
			return
		case <-cm.ticker.C:
			for peerId, peerClient := range cm.server.peerClients {
				cm.wg.Add(1)
				go cm.sendHeartbeatToPeer(peerId, peerClient, &cm.wg, args, &reply)
			}
		}
	}
}
