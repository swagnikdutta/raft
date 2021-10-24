package raft

func (cm *ConsensusModule) becomeFollower() {
	cm.mu.Lock()

	/* If state transition already happened */
	if cm.server.state == FOLLOWER {
		/*
			state change can also be triggered by consecutive, delayed negative acknowledgement from peers.
			It's possible that a candidate/leader had already stepped down to become a follower before receiving the last (negative) heartbeat response.
		*/
		cm.mu.Unlock()
		return
	}

	cm.server.state = FOLLOWER
	// TODO: the unlock can deferred when state change happens in a separate thread
	cm.mu.Unlock()
}
