# TTK4145-RTP
This is a repository containing the real time programming project in TTK4145.

##TODO:
1. Edit FSM
  * Contain all functions necessary to run single elevator using state information gained from driver and output from cost function
  * Send state updates to Network Module
2. Edit cost function
  * Take in status struct and run cost algorithm, give necessary information to FSM module
3. Network
  * Transmit all information received from FSM
  * Make sure that all packages are received at all active peers

--distant future--
1. Status
  * Write backup to file
  * Initialise from file if needed
2. All
  * Run within try{} catch() blocks and have a back process ready to reboot when need or handle other errors if possible,
