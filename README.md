# Raft

**Note:** This project is still under development.

### Background

In replicated databases that follows a leader-follower architecture, read requests are served by the follower nodes while write requests are served by the leader node. On receiving a write request, the leader first writes the new data to its local storage and then sends the data change to all of its followers as part of a replication log. The follower takes the log and updates its local copy of the database, by applying the writes in the same order as they were processed on the leader. 

The system continues to work this way untill the leader node goes down and there's no one to server the write requests anymore. At this point, one of the follower needs to be promoted to be the new leader. The client is reconfigured to send writes to the new leader & the other followers start consuming data changes from the new leader. 

The process of electing a new leader requires all the nodes to reach a _consensus_, agree on one thing - who among them will be the next leader. 

[Raft](https://raft.github.io/raft.pdf) is a well known distributed consensus algorithm, widely accepted owing to it's ease of understandability. 

### Implementation

I started this project while I was attending [The Recurse Center](https://www.recurse.com/) in the Fall of 2021. It's purely for educational purposes so that I can understand distributed systems better and definitely not intended for production use. The implementation, like the algorithm itself followed a modular approach; it was planned in three phases. 
- Leader election :white_check_mark:
- Log replication :x:
- Safety :x:

As of now, only the leader election part has been completed and the rest are being developed.

<!-- ### How to run

```bash
git clone https://github.com/swagnikdutta/raft.git
cd raft
go run main.go
```
 -->
 
 ### See in action
 
 


