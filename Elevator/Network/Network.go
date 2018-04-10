package network

import (
	"fmt"
	"time"

	"../Status"
	"./network/peers"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.

func Network(StatusUpdate chan<- status.UpdateMsg, StatusRefresh chan<- status.StatusStruct, StatusBroadcast <-chan status.StatusStruct, NetworkUpdate <-chan status.UpdateMsg, id string) {
	var peerlist peers.PeerUpdate

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

			if peerlist.Lost != "" {
				update := status.UpdateMsg{
					MsgType:  5,
					Elevator: peerlist.Lost,
				}
				acknowledge.SendUpdate(update)
				StatusUpdate <- update
			}
			if peerlist.New != "" {
				acknowledge.SendStatus(<-StatusBroadcast)
			}
		case update := <-NetworkUpdate:
			acknowledge.SendUpdate(update)
			StatusUpdate <- update

		case update := <-RXupdate:
			ackMessage := AckMsg{
				Id:      id,
				SeqNo:   update.SeqNo,
				MsgType: 0,
			}
			AckSendChan <- ackMessage
			if update.Message.Elevator != id {
				StatusUpdate <- update.Message
			}
		case update := <-RXstate:
			StatusRefresh <- update.Message
			ackMessage := AckMsg{
				Id:      id,
				SeqNo:   update.SeqNo,
				MsgType: 1,
			}
			AckSendChan <- ackMessage

			//Acks and timeouts
		case notReceivedAck := <-TimeoutAckChan:
			switch notReceivedAck.MsgType {
			case 0: //UpdateMessages
				_, ok := sentMessages.UpdateMessages[notReceivedAck.SeqNo]
				if ok && (sentMessages.NumberOfTimesSent[notReceivedAck.SeqNo] < 5) { //Send på nytt hvis sendt 5 ganger
					fmt.Println("No ack - packet loss - resending...")

					updateMessageToSend.Message = sentMessages.UpdateMessages[notReceivedAck.SeqNo]
					updateMessageToSend.SeqNo = notReceivedAck.SeqNo
					TXupdate <- updateMessageToSend
					sentMessages.NumberOfTimesSent[notReceivedAck.SeqNo]++
					newAckStruct := AckStruct{
						AckMessage: AckMsg{
							Id:      notReceivedAck.Id,
							SeqNo:   notReceivedAck.SeqNo,
							MsgType: 0,
						},
						AckTimer: time.NewTimer(15 * time.Millisecond),
					}
					go ackTimer(TimeoutAckChan, newAckStruct)
				}
			case 1: //StatusMessages
				_, ok := sentMessages.StatusMessages[notReceivedAck.SeqNo]
				if ok && (sentMessages.NumberOfTimesSent[notReceivedAck.SeqNo] < 5) {
					fmt.Println("No ack - packet loss - resending...")

					statusMessageToSend.Message = sentMessages.StatusMessages[notReceivedAck.SeqNo]
					statusMessageToSend.SeqNo = notReceivedAck.SeqNo
					TXstate <- statusMessageToSend
					sentMessages.NumberOfTimesSent[notReceivedAck.SeqNo]++
					newAckStruct := AckStruct{
						AckMessage: AckMsg{
							Id:      notReceivedAck.Id,
							SeqNo:   notReceivedAck.SeqNo,
							MsgType: 1,
						},
						AckTimer: time.NewTimer(15 * time.Millisecond),
					}
					go ackTimer(TimeoutAckChan, newAckStruct)
				}
			}
		case recAck := <-AckRecChan:
			delete(sentMessages.NumberOfTimesSent, recAck.SeqNo) //Delete from NumberOfTimesSent
			switch recAck.MsgType {
			case 0: //UpdateMessages
				delete(sentMessages.UpdateMessages, recAck.SeqNo) //Delete from UpdateMessages
			case 1: //StatusMessages
				delete(sentMessages.StatusMessages, recAck.SeqNo) //Delete from StatusMessages
			}
		}
	}
}

func sendUpdate()

func ackTimer(TimeoutAckChan chan<- AckMsg, ackStruct AckStruct) {
	for {
		select {
		case <-ackStruct.AckTimer.C:
			TimeoutAckChan <- ackStruct.AckMessage
			return
		}
	}
}
