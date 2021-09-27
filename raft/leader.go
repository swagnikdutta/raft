package raft

import "fmt"

type LeaderVolatile struct {
	nextIndex  int
	matchIndex int
}

type Leader struct {
	Server         Server
	LeaderVolatile LeaderVolatile
}

func doSomething() {
	fmt.Println("Inside doSomething")
	leader := new(Leader)
	fmt.Printf("%+v\n", leader)
}
