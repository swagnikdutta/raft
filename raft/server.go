package raft

import "fmt"

// An entry is considered committed when it's safe for that entry to be applied to state machines
type Entry struct {
	term    int8   // the term when the entry was created
	command string // command for the state machine
}

type Log struct {
	logs []Entry
}

type persistent struct {
}
type volatile struct {
}

type Server struct {
	// persistent state: Updated on stable storage before responding to RPCs.
	log Log

	// volatile state
}

func NewServer() {
	fmt.Println("creating a new server")
}
