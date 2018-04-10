package acknowledge

import (
	"fmt"
	"time"

	"../../../Status"
	"../bcast"
)

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

var ID string
var seqNo = 0
var updateMessageToSend UpdateMessageStruct
var statusMessageToSend StatusMessageStruct
var sentMessages = new(SentMessages)

//TODO: should these be private variables -> change starting letter to lower case
var TXupdate = make(chan UpdateMessageStruct)
var TXstate = make(chan StatusMessageStruct)
var RXupdate = make(chan UpdateMessageStruct)
var RXstate = make(chan StatusMessageStruct)
var AckSendChan = make(chan AckMsg)
var AckRecChan = make(chan AckMsg)
var TimeoutAckChan = make(chan AckMsg)

//TODO: Differentiate on elevator id

func Ack(newUpdate chan<- status.UpdateMsg, newStatus chan<- status.StatusStruct) {
	sentMessages.UpdateMessages = make(map[int]status.UpdateMsg)
	sentMessages.StatusMessages = make(map[int]status.StatusStruct)
	sentMessages.NumberOfTimesSent = make(map[int]int)
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
			newUpdate <- update.Message
		case status := <-RXstate:
			ackMessage := AckMsg{
				Id:      ID,
				SeqNo:   status.SeqNo,
				MsgType: 1,
			}
			AckSendChan <- ackMessage
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

func SendUpdate(update status.UpdateMsg) {
	seqNo += 1
	updateMessageToSend.Message = update
	updateMessageToSend.SeqNo = seqNo
	TXupdate <- updateMessageToSend
	sentMessages.UpdateMessages[seqNo] = update
	sentMessages.NumberOfTimesSent[seqNo] = 1

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

func SendStatus(statusUpdate status.StatusStruct) {
	seqNo += 1
	statusMessageToSend.Message = statusUpdate
	statusMessageToSend.SeqNo = seqNo
	TXstate <- statusMessageToSend
	sentMessages.StatusMessages[seqNo] = statusMessageToSend.Message
	sentMessages.NumberOfTimesSent[seqNo] = 1
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

func ackTimer(TimeoutAckChan chan<- AckMsg, ackStruct AckStruct) {
	for {
		select {
		case <-ackStruct.AckTimer.C:
			TimeoutAckChan <- ackStruct.AckMessage
			return
		}
	}
}
