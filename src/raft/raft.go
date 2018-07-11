package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//


import (
	"../labrpc"
	"fmt"
	"sync"
	"math/rand"
	"time"
)

// import "bytes"
// import "encoding/gob"



//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
//
type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}


type LeaderReply struct {
	Rep       int
}
//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        	sync.Mutex
	peers     	[]*labrpc.ClientEnd
	persister 	*Persister
	me        	int // index into peers[]
	state	 	string //mine
	CurrentTerm	  	int //mine
	votedFor	[1000] int //mine
	candidateID int //mine
	finalVotedFor [1000] int //mine
	voteCount [1000] int //mine
	leaderTerm int //mine
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool
	term = rf.CurrentTerm //mine
	if rf.finalVotedFor[term] == rf.me{
		isleader = true
	} else{
		isleader = false
	}
	return term, isleader
}
/*func (rf *Raft) GetState1() (int, int) {
	var term int
	var votedFor int
	term = rf.CurrentTerm //mine
	votedFor = rf.votedFor
	return term, votedFor
}*/

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here.
	// Example:
	// w := new(bytes.Buffer)
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	// Your code here.
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)
}


type AppendEntries struct {
	Term int
	LeaderID int
}

//
// example RequestVote RPC arguments structure.
//
type RequestVoteArgs struct {
	// Your data here.
	Term int
	CandidateID int
}

//
// example RequestVote RPC reply structure.
//
type RequestVoteReply struct {
	// Your data here.
	VoteGranted bool
	Term int
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here.
	var CallerTerm = args.Term
	var CalleeTerm = rf.CurrentTerm
	if CalleeTerm > CallerTerm{
		reply.Term = CalleeTerm
		reply.VoteGranted = false
	}else if CalleeTerm < CallerTerm{
		rf.CurrentTerm = CallerTerm
		rf.votedFor[CallerTerm] = args.CandidateID
		reply.Term = CallerTerm
		reply.VoteGranted = true
	}else{
		if rf.votedFor[CallerTerm] == -1{
			rf.votedFor[CallerTerm] = args.CandidateID
			reply.Term = CallerTerm
			reply.VoteGranted = true
		}else{
			reply.Term = CalleeTerm
			reply.VoteGranted = false
		}
	}
	if rf.candidateID != args.CandidateID{
		rand.Seed(time.Now().UnixNano())
		r := rand.Intn(500)
		var t = time.Duration(r) * time.Millisecond
		time.Sleep(t)
	}
}


func (rf *Raft) AppendEntries(args AppendEntries, reply *LeaderReply) {
	var LeaderTerm = args.Term
	var LeaderID = args.LeaderID
	rf.finalVotedFor[LeaderTerm] = LeaderID
	//rf.CurrentTerm = LeaderTerm
	rf.voteCount[LeaderTerm] = 0
	rf.votedFor[LeaderTerm] = LeaderID
	fmt.Println("Append Entries")
	fmt.Printf("Follower Id: %v, term: %v, Leader ID: %v\n", rf.candidateID,LeaderTerm, LeaderID)
	if rf.me != LeaderID{
		rand.Seed(time.Now().UnixNano())
		r := rand.Intn(500)
		var t = time.Duration(r) * time.Millisecond
		time.Sleep(t)
	}
	reply.Rep = 1
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// returns true if labrpc says the RPC was delivered.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args AppendEntries, reply *LeaderReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}


//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true


	return index, term, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.state = "follower"
	rf.CurrentTerm = -1
	rf.votedFor[0] = -1
	rf.finalVotedFor[0] = -1
	rf.voteCount[0] = 0
	rf.candidateID = me
	rf.leaderTerm = -1

	//time.Sleep(2000 * time.Millisecond)
	//for i := 0; i < 2; i++ {
	go func() {

		r := rand.Intn(1000)
		var t = time.Duration(r) * time.Millisecond
		time.Sleep(t)
		//fmt.Printf("Index: %v , time: %v\n", me, t)
		rf.CurrentTerm = rf.CurrentTerm + 1
		var cterm = rf.CurrentTerm

		var result RequestVoteReply
		for j := 0; j < 3 ; j++{
			rf.sendRequestVote(j,RequestVoteArgs{cterm,rf.candidateID},&result)
			if result.VoteGranted == true{
				rf.voteCount[cterm] = rf.voteCount[cterm] + 1
				if rf.voteCount[cterm] == 2{
					fmt.Printf("Leader Id: %v, Term: %v\n", rf.candidateID, cterm)
					var rep LeaderReply
					for m:=0; m < 3; m++{
						rf.sendAppendEntries(m,AppendEntries{cterm,rf.candidateID},&rep)
					}
					for cterm >= rf.CurrentTerm{
						r := rand.Intn(2)
						var t = time.Duration(r) * time.Millisecond
						time.Sleep(t)
						for m:=0; m < 3; m++{
							rf.sendAppendEntries(m,AppendEntries{cterm,rf.candidateID},&rep)
						}
					}
				}
			}else{
				rf.CurrentTerm = result.Term
			}
		}


	}()
	//	}
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())


	return rf
}