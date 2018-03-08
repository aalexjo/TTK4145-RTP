package network

import (

	"fmt"
	"os"

	"./network/bcast"
	"./network/localip"
	"./network/peers"
	"../Status"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

func Network(StatusUpdate chan<- status.UpdateMsg, NetworkUpdate <-chan status.UpdateMsg, ElevID string, id string) {

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	TXupdate := make(status.UpdateMsg)
	//TXstate := make(status.StatusStruct) //used when node needs to sync with network

	// Start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, TXchannel) //TODO: fix ports
	go bcast.Receiver(16569, NetworkUpdate)

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if p.Lost != ""{
				update := status.UpdateMsg{
					MsgType: 6,
					Elevator: p.Lost,
				}
				TXchannel <- update
				StatusUpdate <- update
			}
			if p.New != ""{
				//TODO: transmit full state information
			}
		case update := <-NetworkUpdate:
			TXchannel <- update
			StatusUpdate <- update
	}
}
