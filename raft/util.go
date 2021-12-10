package raft

type If bool

func (c If) String(msg1, msg2 string) string {
	if c {
		return msg1
	}
	return msg2
}

func GetTimeoutRange() (int, int) {
	return 3, 5
}
