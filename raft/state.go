package raft

type serverState string

const (
	LEADER    serverState = "leader"
	FOLLOWER              = "follower"
	CANDIDATE             = "candidate"
)
