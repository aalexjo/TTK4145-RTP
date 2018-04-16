# TTK4145-RTP
This is a repository containing the real time programming project in TTK4145.

Our design is based on a peer to peer principle where all elevators(nodes) have identical information if they are connected to the network. By using this assumtion we can use an identical cost function locally at each elevator and have them arrive at the same conclusion for what elevator is to execute what order. When a new elevator connects to the network it shares its entire state information with the network, and the network transmits their inforamtion back in order for everyone to have a consensus. 

The program flow is seen below

![Program diagram](https://github.com/aalexjo/TTK4145-RTP/blob/master/Design/SanntidDiagram%20(2).png)

The main task of the different modules are as follows:
-Status: Save the current information of internal and network states
-Cost: calculate what elevator is to execute what hall requests
-FSM: Run the elevator to its assigned floors and control lighting, recive button presses and other updates from the elevator.
-Network: Transmit updates to other nodes on the netork and update status module.
 -Acknowledge: Garantuee delivery of updates to other connected elevators, even in the case of a bad network connection.
 
 Each module, except Status, has a short memory span; meaning that they do not store state information locally, but rely on updates from other modules. FSM recives continues updates from Cost which recives continues updates from Status, these are only in scope for a short time before being discarded.
