package acknowledge

import (
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

var seqNo = 0
var updateMessageToSend UpdateMessageStruct
var statusMessageToSend StatusMessageStruct
var sentMessages = new(SentMessages)
sentMessages.UpdateMessages = make(map[int]status.UpdateMsg)
sentMessages.StatusMessages = make(map[int]status.StatusStruct)
sentMessages.NumberOfTimesSent = make(map[int]int)

//TODO: should these be private variables -> change starting letter to lower case
var TXupdate = make(chan UpdateMessageStruct)
var TXstate = make(chan StatusMessageStruct)
var RXupdate = make(chan UpdateMessageStruct)
var RXstate = make(chan StatusMessageStruct)
var AckSendChan = make(chan AckMsg)
var AckRecChan = make(chan AckMsg)
var TimeoutAckChan = make(chan AckMsg)

//TODO: Differentiate on elevator id

func ack() {
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
			Id:      id,
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
			Id:      id,
			SeqNo:   seqNo,
			MsgType: 1,
		},
		AckTimer: time.NewTimer(15 * time.Millisecond),
	}
	go ackTimer(TimeoutAckChan, newAckStruct)
}
