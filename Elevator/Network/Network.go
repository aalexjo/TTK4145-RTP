package network

import (
	"fmt"
	"time"

	"../Status"
	"./network/bcast"
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

type UpdateMessageStruct struct {
	Message status.UpdateMsg
	SeqNo   int
}

type StatusMessageStruct struct {
	Message status.StatusStruct
	SeqNo   int
}

func Network(StatusUpdate chan<- status.UpdateMsg, StatusRefresh chan<- status.StatusStruct, StatusBroadcast <-chan status.StatusStruct, NetworkUpdate <-chan status.UpdateMsg, id string) {
	var seqNo = 0
	var peerlist peers.PeerUpdate
	var updateMessageToSend UpdateMessageStruct
	var statusMessageToSend StatusMessageStruct

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
	go peers.Transmitter(16016, id, peerTxEnable)
	go peers.Receiver(16016, peerUpdateCh)

	TXupdate := make(chan UpdateMessageStruct)
	TXstate := make(chan StatusMessageStruct)
	RXupdate := make(chan UpdateMessageStruct)
	RXstate := make(chan StatusMessageStruct)
	AckSendChan := make(chan AckMsg)
	AckRecChan := make(chan AckMsg)
	TimeoutAckChan := make(chan AckMsg)

	// Start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, TXupdate, TXstate, AckSendChan) //TODO: fix ports
	go bcast.Receiver(16569, RXupdate, RXstate, AckRecChan)

	//TODO: Differentiate on elevator id

	fmt.Println("Started")
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
				seqNo += 1
				updateMessageToSend.Message = update
				updateMessageToSend.SeqNo = seqNo
				TXupdate <- updateMessageToSend
				sentMessages.UpdateMessages[seqNo] = update
				sentMessages.NumberOfTimesSent[seqNo] += 1
				newAckStruct := AckStruct{
					AckMessage: AckMsg{
						Id:      id,
						SeqNo:   seqNo,
						MsgType: 0,
					},
					AckTimer: time.NewTimer(15 * time.Millisecond),
				}
				go ackTimer(TimeoutAckChan, newAckStruct)

				StatusUpdate <- update
			}
			if peerlist.New != "" {
				seqNo += 1
				statusMessageToSend.Message = <-StatusBroadcast
				statusMessageToSend.SeqNo = seqNo
				TXstate <- statusMessageToSend
				sentMessages.StatusMessages[seqNo] = statusMessageToSend.Message
				sentMessages.NumberOfTimesSent[seqNo] += 1
				newAckStruct := AckStruct{
					AckMessage: AckMsg{
						Id:      id,
						SeqNo:   seqNo,
						MsgType: 0,
					},
					AckTimer: time.NewTimer(15 * time.Millisecond),
				}
				go ackTimer(TimeoutAckChan, newAckStruct)
			}
		case update := <-NetworkUpdate:
			seqNo += 1
			updateMessageToSend.Message = update
			updateMessageToSend.SeqNo = seqNo
			TXupdate <- updateMessageToSend

			//Wait for ack on this message
			sentMessages.UpdateMessages[seqNo] = update
			sentMessages.NumberOfTimesSent[seqNo] += 1
			newAckStruct := AckStruct{
				AckMessage: AckMsg{
					Id:      id,
					SeqNo:   seqNo,
					MsgType: 0,
				},
				AckTimer: time.NewTimer(15 * time.Millisecond),
			}
			go ackTimer(TimeoutAckChan, newAckStruct)

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
							MsgType: 0,
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

func ackTimer(TimeoutAckChan chan<- AckMsg, ackStruct AckStruct) {
	for {
		select {
		case <-ackStruct.AckTimer.C:
			TimeoutAckChan <- ackStruct.AckMessage
			return
		}
	}
}
