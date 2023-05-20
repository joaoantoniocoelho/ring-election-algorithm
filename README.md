# ring-election-algorithm
Ring Election Algorithm for my Parallel and Distributed Processing class

Implement the Distributed Election Algorithm based on a logica ring with Go, the code must have the following requirements:
> 1. Based on the use of a logical ring;
> 2. Each process knows the entire ring, but sends messages only to the next active process in the direction of the ring;
> 3. When the process detects that the coordinator is no longer active (this can be simulated with a message from an external process), it sends an election message on the ring containing its id (disputing this vacant coordination;
> 4. At each step, the process that receives the election message includes its id in the message and passes it on through the ring.
> 5. At the end of a complete turn, the process that initiated the election receives the message and chooses the one with the highest id as the new coordinating process;
> 6. A new message is sent through the ring to let everyone know the new coordinator;

It's also necessary to model a way to show that it works and simulate fails so that response is not always the same (nodes that stop working).

<img src="https://imgur.com/a/F1CyZDf" />
