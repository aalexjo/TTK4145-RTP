package network

import (

	"fmt"
	//"os"

	"./network/bcast"
	//"./network/localip"
	"./network/peers"
	"../Status"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

func Network(StatusUpdate chan<- status.UpdateMsg, StatusRefresh chan<- status.StatusStruct, StatusBroadcast <-chan status.StatusStruct, NetworkUpdate <-chan status.UpdateMsg, id string) {

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	TXupdate := make(chan status.UpdateMsg)
	TXstate := make(chan status.StatusStruct)
	RXupdate := make(chan status.UpdateMsg)
	//RXstate := make(chan status.StatusStruct)
	// Start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, TXupdate, TXstate) //TODO: fix ports
	go bcast.Receiver(16569, RXupdate, StatusRefresh)

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
					MsgType: 5,
					Elevator: p.Lost,
				}
				TXupdate <- update
				StatusUpdate <- update
			}
			if p.New != ""{
				//TXstate <- StatusBroadcast
			}
		case update := <-NetworkUpdate:
			TXupdate <- update
			StatusUpdate <- update
			fmt.Println(update)
		
		case update := <-RXupdate:
			if update.Elevator != id{
				StatusUpdate <- update
			}
		}
	}
}
