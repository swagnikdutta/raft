# Raft

This is an educational project to understand how consensus works. It features a toy implementation of the raft protocol. 

It is partially complete, till the leader election step. 

The gif below demonstrates a cluster with three  nodes. The server with `id = 101` wins the election to become the leader and starts sending heartbeats to its followers - `102` and `103` at fixed intervals. The heartbeat resets the timer on the followers, preventing them from transitioning into the candidate state.



![](https://github.com/swagnikdutta/repository-assets/blob/main/leader-election-raft.gif)


