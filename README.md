# Raft

### Implementation

I started this project while attending [The Recurse Center](https://www.recurse.com/) in the Fall of 2021. It's purely for educational purposes so I can better understand distributed systems. As of now, only the leader election part has been completed.

A cluster with three  nodes. The server with `id = 101` wins the election to become the leader and starts sending heartbeats to its followers - `102` and `103` at fixed intervals. The heartbeat resets the timer on the followers, preventing them from transitioning into Candidate state.

![](https://github.com/swagnikdutta/repository-assets/blob/main/leader-election-raft.gif)


