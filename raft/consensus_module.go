package raft

import "fmt"

type ConsensusModule struct {
}

func StartElections(server *Server) {
	fmt.Printf("%+v is starting elections\n", server)
}
