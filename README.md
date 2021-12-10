# Raft

### Background

In replicated databases that follows a leader-follower architecture, read requests are served by the follower nodes while write requests are served by the leader node. On receiving a write request, the leader first writes the new data to its local storage and then sends the data change to all of its followers as part of a replication log. The follower takes the log and updates its local copy of the database, by applying the writes in the same order as they were processed on the leader. 

The system continues to work this way untill the leader node goes down and there's no one to server the write requests anymore. At this point, one of the follower needs to be promoted to be the new leader. The client is reconfigured to send writes to the new leader & the other followers start consuming data changes from the new leader. 

