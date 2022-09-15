package roles

type RaftRole int

const (
	Follower  RaftRole = 1
	Candidate RaftRole = 2
	Leader    RaftRole = 3
)
