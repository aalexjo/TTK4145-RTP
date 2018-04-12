# TTK4145-RTP
This is a repository containing the real time programming project in TTK4145.

## TODO:
1. Edit FSM
 *Handle broken motor
2. Network
  * Test ack implementation, add induvidule elevator acks

4. watchdog
  * implement wachdog module that tracks all other modules, is able to terminate them if they crash or hang and spawn replacemnets

after
 -Clean up code and comments
 -add function comments

 if time.

 -implement stop button
 -implement obstruction


 ### ISSUES:
 - Broken motor seems to freeze the system
 -Uncomment status broadcast in Network when receiving new peer, this makes the ack module wait for a status message it never receives acks for.
 -Watchdog infinitely spawns new processes when started with bad port input(simulator is not running)
 -^however when running on another computer it seems to be unable to spawn the other process
