# Elevator design
>This is the design part of the elevator project.

### Network
Every elevator operates as an individual node inn a peer-to-peer based network module. We have chosen to use the UDP-protocol to do this, because it has all the features we think we need. However since it does not include a packet-receive acknowledge functionality, we plan on implementing this as a part of the network module.  


### Redundancy
- Local copies on each HDD
- Checksum on UDP packets
- Loss of packets, resend, ack, 
- 
