package raft

// type interval struct {
// 	start, end int
// }

// var Ma = map[string]interval{
// 	"a": interval{2, 3},
// 	"v": interval{2, 3},
// }

func GetTimeoutRange() (int, int) {
	return 3, 5
}
