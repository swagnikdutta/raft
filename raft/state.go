package raft

type serverState string

const (
	LEADER    serverState = "leader"
	FOLLOWER              = "follower"
	CANDIDATE             = "candidate"
)

func HandleStateTransition(server *Server, nextState string) {
	if nextState == CANDIDATE {
		server.state = nextState
		StartElections(server)
	}
}
