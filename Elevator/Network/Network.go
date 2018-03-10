package network

import (
	"fmt"
	//"os"
	"time"

	"./network/bcast"
	//"./network/localip"
	"../Status"
	"./network/peers"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

/*---------------------Ack message struct---------------------
Id: ID of the Elevator
SeqNo: Sequence number of the UDP packet
MsgType: Type of UDP message that was sent/received out of:
	- 0: UpdateMessages
	- 1: StatusMessages
-------------------------------------------------------------*/
type AckMsg struct {
	Id      string
	SeqNo   int
	MsgType int
}

type SentMessages struct {
	UpdateMessages    map[int]status.UpdateMsg
	StatusMessages    map[int]status.StatusStruct
	NumberOfTimesSent map[int]int
}

type AckStruct struct {
	AckMessage AckMsg
	AckTimer   *time.Timer
}

func Network(StatusUpdate chan<- status.UpdateMsg, StatusRefresh chan<- status.StatusStruct, StatusBroadcast <-chan status.StatusStruct, NetworkUpdate <-chan status.UpdateMsg, id string) {
	var SeqNo = 0 //Denne må kanskje være et annet sted. Global??
	var peerlist peers.PeerUpdate

	sentMessages := new(SentMessages)
	sentMessages.UpdateMessages = make(map[int]status.UpdateMsg)
	sentMessages.StatusMessages = make(map[int]status.StatusStruct)
	sentMessages.NumberOfTimesSent = make(map[int]int)

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
	AckSendChan := make(chan AckMsg)
	AckRecChan := make(chan AckMsg)
	TimeoutAckChan := make(chan AckMsg)

	// Start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, TXupdate, TXstate, AckSendChan) //TODO: fix ports
	go bcast.Receiver(16569, RXupdate, StatusRefresh, AckRecChan)

	//Timeout channel ---- CHANGE TO A BETTER TIMEOUT VALUE???-------
	//ackInterval := time.NewTimer(15 * time.Millisecond)
	//ackTimeout := time.NewTimer(50 * time.Millisecond)
	//ackInterval.Stop()
	//ackTimeout.Stop()

	fmt.Println("Started")
	for {
		select {
		case peerlist = <-peerUpdateCh:
			AckSendChan <- ackMsg
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerlist.Peers)
			fmt.Printf("  New:      %q\n", peerlist.New)
			fmt.Printf("  Lost:     %q\n", peerlist.Lost)

			if peerlist.Lost != "" {
				update := status.UpdateMsg{
					MsgType:  5,
					Elevator: peerlist.Lost,
				}
				TXupdate <- update
				newAckStruct := AckStruct{
					AckMessage: AckMsg{
						Id: id,
						//SeqNo: //TODO: Add sqeuence numbers to every message?
						MsgType: 0,
					},
					AckTimer: time.NewTimer(15 * time.Millisecond),
				}
				go ackTimer(TimeoutAckChan, newAckStruct)
				StatusUpdate <- update
			}
			if peerlist.New != "" {
				//TXstate <- StatusBroadcast
				//ackTimeout.Reset(100 * time.Millisecond)
			}
		case update := <-NetworkUpdate:
			TXupdate <- update
			newAckStruct := AckStruct{
				AckMessage: AckMsg{
					Id: id,
					//SeqNo: //TODO: Add sqeuence numbers to every message?
					MsgType: 0,
				},
				AckTimer: time.NewTimer(15 * time.Millisecond),
			}
			go ackTimer(TimeoutAckChan, newAckStruct)

			StatusUpdate <- update

			//Case when ack is not received

			//Acks and timeouts
		case notReceivedAck := <-TimeoutAckChan:
			switch notReceivedAck.MsgType {
			case 0: //UpdateMessages
				_, ok := sentMessages.UpdateMessages[notReceivedAck.SeqNo]
				if ok {
					fmt.Println("No ack - packet loss - resending...")

					TXupdate <- sentMessages.UpdateMessages[notReceivedAck.SeqNo]
					newAckStruct := AckStruct{
						AckMessage: AckMsg{
							Id: notReceivedAck.Id,
							//SeqNo: //TODO: Add sqeuence numbers to every message?
							MsgType: 0,
						},
						AckTimer: time.NewTimer(15 * time.Millisecond),
					}
					go ackTimer(TimeoutAckChan, newAckStruct)
				}
			case 1: //StatusMessages
				fmt.Println("No ack - packet loss - resending...")

				_, ok := sentMessages.StatusMessages[notReceivedAck.SeqNo]
				if ok {
					TXstate <- sentMessages.StatusMessages[notReceivedAck.SeqNo]
					newAckStruct := AckStruct{
						AckMessage: AckMsg{
							Id: notReceivedAck.Id,
							//SeqNo: //TODO: Add sqeuence numbers to every message?
							MsgType: 0,
						},
						AckTimer: time.NewTimer(15 * time.Millisecond),
					}
					go ackTimer(TimeoutAckChan, newAckStruct)
				}
			}
		case recAck := <-AckRecChan:
			switch recAck.MsgType {
			case 0: //UpdateMessages
				delete(sentMessages.UpdateMessages, recAck.SeqNo)
			case 1: //StatusMessages
				delete(sentMessages.StatusMessages, recAck.SeqNo)
			}
		}
	}
}

func ackTimer(TimeoutAckChan chan<- AckMsg, ackStruct AckStruct) {
	for {
		select {
		case <-ackStruct.AckTimer.C:
			TimeoutAckChan <- ackStruct.AckMessage
			return
		}
	}
}
