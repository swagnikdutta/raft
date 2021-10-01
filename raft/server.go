package raft

type Server struct {
	id    int
	peers []int
}

func CreateNewServer() *Server {
	s := new(Server)
	// set other server properties here
	return s
}
