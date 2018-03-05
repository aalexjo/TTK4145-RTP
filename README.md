# TTK4145-RTP
This is a repository containing the real time programming project in TTK4145.

##TODO:
1. Edit FSM
  * contain all functions necessairy to run single elevator using state information gained from driver and output from cost function
  * Send state updates to Network Module 
2. Edit cost function
  * Take in status struct and run cost algorithm, give necceairy information to FSM module
3. Network
  * transmit all information recived from FSM
  * Make sure that all packages are recived at all active peers

--distant future--
1. Status
  * Write backup to file
  * Initialize from file if needed
2. All
  * Run within try{} catch() blocks and have a back prosess ready to reboot when need or handle other errors if possible,
