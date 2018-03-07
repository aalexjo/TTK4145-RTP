# TTK4145-RTP
This is a repository containing the real time programming project in TTK4145.

##TODO:
> Edit FSM
  * contain all functions necessary to run single elevator using state information gained from driver and output from cost function
  * Send state updates to Network Module 
  * thouroughly check that all input variables are correctly used throughout module 

> Network
  * transmit all information received from FSM
  * Make sure that all packages are received at all active peers

>possibly combine cost and status modules

> Check if prefixes such as cost.Cost and status.UpdateMsg are used correctly or if they are needed at all...


--distant future--
1. Status
  * Write backup to file
  * Initialize from file if needed
2. All
  * Run within try{} catch() blocks and have a back prosess ready to reboot when need or handle other errors if possible,


completed
2. Edit cost function
  * Take in status struct and run cost algorithm, give necceary information to FSM module


