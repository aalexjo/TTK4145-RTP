package network

import (
	"fmt"

	"../Status"
	"./network/acknowledge"
	"./network/peers"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
func Network(StatusUpdate chan<- status.UpdateMsg, StatusRefresh chan<- status.StatusStruct, StatusBroadcast <-chan status.StatusStruct, NetworkUpdate <-chan status.UpdateMsg, id string) {
	var peerlist peers.PeerUpdate

	newUpdate := make(chan status.UpdateMsg)
	newStatus := make(chan status.StatusStruct)
	ackPeerUpdate := make(chan peers.PeerUpdate)
	acknowledge.ID = id
	go acknowledge.Ack(newUpdate, newStatus, ackPeerUpdate)
	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(16016, id, peerTxEnable)
	go peers.Receiver(16016, peerUpdateCh)

	fmt.Println("Started network")
	for {
		select {
		case peerlist = <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerlist.Peers)
			fmt.Printf("  New:      %q\n", peerlist.New)
			fmt.Printf("  Lost:     %q\n", peerlist.Lost)

			ackPeerUpdate <- peerlist
			if peerlist.Lost != "" {
				update := status.UpdateMsg{
					MsgType:  5,
					Elevator: peerlist.Lost,
				}
				//acknowledge.SendUpdate(update)
				StatusUpdate <- update
			}
			if peerlist.New != "" {
				acknowledge.SendStatus(<-StatusBroadcast)
				//fmt.Println(<-StatusBroadcast)
			}
		case update := <-NetworkUpdate:
			if update.MsgType == 8 { //update.Direction == "stop" && update.MsgType == 3 {
				fmt.Println("disable TX")
				peerTxEnable <- false
			} else {
				peerTxEnable <- true
			}
			acknowledge.SendUpdate(update)
			StatusUpdate <- update

		case update := <-newUpdate:
			if update.Elevator != id {
				StatusUpdate <- update
			}

		case status := <-newStatus:
			StatusRefresh <- status
		}
	}
}
