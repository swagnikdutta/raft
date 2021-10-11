package raft

import "fmt"

type ConsensusModule struct {
	server      *Server
	currentTerm int
}

// Methods
func (cm *ConsensusModule) ChangeState(nextState string) {
	cm.server.state = nextState
	fmt.Printf("Server %v became a Candidate\n", cm.server.id)
	cm.startElections()
}

func (cm *ConsensusModule) startElections() {
	// this server wants to be the next leader
	fmt.Printf("Server %v is starting elections\n", cm.server.id)
}

// procedures
type Request struct {
}
type Response struct {
}

// Functions

func NewConsensusModule(server *Server) *ConsensusModule {
	cm := &ConsensusModule{
		currentTerm: 1,
		server:      server,
	}
	return cm
}
