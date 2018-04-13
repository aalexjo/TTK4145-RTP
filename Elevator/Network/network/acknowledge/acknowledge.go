package acknowledge

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"../../../Status"
	"../bcast"
	"../peers"
)

//TODO: fix den første meldinga som sendes
//TODO: Fix maps

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
	StatusMessages    map[int]status.StatusStruct //TODO: Agree on issue with pointer?
	NumberOfTimesSent map[int]int                 //Number of times sent for each sequence no
	NotRecFromPeer    map[int][]string            //Acks not received from active peers per seq. no.
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

var ID string
var PORT string
var seqNo = 0
var updateMessageToSend UpdateMessageStruct
var statusMessageToSend StatusMessageStruct
var sentMessages = new(SentMessages)
var peerlist peers.PeerUpdate

//TODO: should these be private variables -> change starting letter to lower case
var TXupdate = make(chan UpdateMessageStruct)
var TXstate = make(chan StatusMessageStruct)
var RXupdate = make(chan UpdateMessageStruct)
var RXstate = make(chan StatusMessageStruct)
var AckSendChan = make(chan AckMsg)
var AckRecChan = make(chan AckMsg)
var TimeoutAckChan = make(chan AckMsg)

func Ack(newUpdate chan<- status.UpdateMsg, newStatus chan<- status.StatusStruct, peerUpdate <-chan peers.PeerUpdate) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r, " ACK fatal panic, unable to recover. Rebooting...", "go run main.go -init=false -port="+PORT, " -id="+ID)
			err := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go -init=false -port="+PORT+" -id="+ID).Run()
			if err != nil {
				fmt.Println("Unable to reboot process, crashing...")
			}
		}
		os.Exit(0)
	}()
	sentMessages.UpdateMessages = make(map[int]status.UpdateMsg)
	sentMessages.StatusMessages = make(map[int]status.StatusStruct)
	sentMessages.NumberOfTimesSent = make(map[int]int)
	sentMessages.NotRecFromPeer = make(map[int][]string)

	// Start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, TXupdate, TXstate, AckSendChan) //TODO: fix ports
	go bcast.Receiver(16569, RXupdate, RXstate, AckRecChan)

	for {
		select {
		case update := <-RXupdate:
			ackMessage := AckMsg{
				Id:      ID,
				SeqNo:   update.SeqNo,
				MsgType: 0,
			}
			AckSendChan <- ackMessage
			if update.Message.Elevator != ID {
				newUpdate <- update.Message
			}

		case status := <-RXstate:
			ackMessage := AckMsg{
				Id:      ID,
				SeqNo:   status.SeqNo,
				MsgType: 1,
			}
			AckSendChan <- ackMessage
			newStatus <- status.Message
		case notReceivedAck := <-TimeoutAckChan:
			switch notReceivedAck.MsgType {
			case 0: //UpdateMessages
				_, ok := sentMessages.UpdateMessages[notReceivedAck.SeqNo]
				if ok && (sentMessages.NumberOfTimesSent[notReceivedAck.SeqNo] < 10) { //Resend if sent <10 times
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
				if ok && (sentMessages.NumberOfTimesSent[notReceivedAck.SeqNo] < 10) {
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
			//fmt.Println("før ackreceived: ", sentMessages.NotRecFromPeer)
			//fmt.Println("før, updates: ", sentMessages.UpdateMessages)
			//fmt.Println("før, state: ", sentMessages.StatusMessages)
			_, ok := sentMessages.NotRecFromPeer[recAck.SeqNo] //In case the seqno has been deleted unexpectedly
			if ok {
				ind := stringInSlice(recAck.Id, sentMessages.NotRecFromPeer[recAck.SeqNo])
				if ind != -1 {
					sentMessages.NotRecFromPeer[recAck.SeqNo] = removeFromSlice(sentMessages.NotRecFromPeer[recAck.SeqNo], ind)
					if len(sentMessages.NotRecFromPeer[recAck.SeqNo]) == 0 {
						delete(sentMessages.NotRecFromPeer, recAck.SeqNo)
						delete(sentMessages.NumberOfTimesSent, recAck.SeqNo) //Delete from NumberOfTimesSent
						switch recAck.MsgType {
						case 0: //UpdateMessages
							delete(sentMessages.UpdateMessages, recAck.SeqNo) //Delete from UpdateMessages
						case 1: //StatusMessages
							delete(sentMessages.StatusMessages, recAck.SeqNo) //Delete from StatusMessages
						}
					}
				}
				//fmt.Println("etter ackreceived: ", sentMessages.NotRecFromPeer)
				//fmt.Println("etter, updates: ", sentMessages.UpdateMessages)
				//fmt.Println("etter, state: ", sentMessages.StatusMessages)
			}
			//Should delete peer from NotRecFromPeer
		case peerlist = <-peerUpdate:
			if peerlist.Lost != "" {
				//fmt.Println("Før peerupdate: ", sentMessages.NotRecFromPeer) //remove
				for seqNo, peers := range sentMessages.NotRecFromPeer {
					ind := stringInSlice(peerlist.Lost, peers)
					if ind != -1 {
						sentMessages.NotRecFromPeer[seqNo] = removeFromSlice(sentMessages.NotRecFromPeer[seqNo], ind)
					}
				}
				//fmt.Println("etter: ", sentMessages.NotRecFromPeer) //remove
			}
		}
	}
}

func SendUpdate(update status.UpdateMsg) {
	seqNo += 1
	updateMessageToSend.Message = update
	updateMessageToSend.SeqNo = seqNo
	TXupdate <- updateMessageToSend
	if len(peerlist.Peers) != 0 {
		sentMessages.UpdateMessages[seqNo] = update
		sentMessages.NumberOfTimesSent[seqNo] = 1
		sentMessages.NotRecFromPeer[seqNo] = peerlist.Peers

		newAckStruct := AckStruct{
			AckMessage: AckMsg{
				Id:      ID,
				SeqNo:   seqNo,
				MsgType: 0,
			},
			AckTimer: time.NewTimer(15 * time.Millisecond),
		}
		go ackTimer(TimeoutAckChan, newAckStruct)
	}
}

func SendStatus(statusUpdate status.StatusStruct) {
	seqNo += 1
	statusMessageToSend.Message = statusUpdate
	statusMessageToSend.SeqNo = seqNo
	TXstate <- statusMessageToSend
	if len(peerlist.Peers) != 0 {
		sentMessages.StatusMessages[seqNo] = statusMessageToSend.Message
		sentMessages.NumberOfTimesSent[seqNo] = 1
		sentMessages.NotRecFromPeer[seqNo] = peerlist.Peers

		newAckStruct := AckStruct{
			AckMessage: AckMsg{
				Id:      ID,
				SeqNo:   seqNo,
				MsgType: 1,
			},
			AckTimer: time.NewTimer(15 * time.Millisecond),
		}
		go ackTimer(TimeoutAckChan, newAckStruct)
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

//Utilities for arrays in golang
func stringInSlice(a string, list []string) int {
	for ind, b := range list {
		if b == a {
			return ind
		}
	}
	return -1
}

func removeFromSlice(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
