package status

import (
	"os"
)

type UpdateMsg struct {
	//encoding of relevant update information
	//copy from network module
	MsgType int
	// 0 = hallRequest
	// 1 = behaviour
	// 2 = arrivedAtfloor
	// 3 = newDirection
	// 4 = cabRequest
	Elevator int //used in all other than 0
	Floor    int //used in 0,2,4
	Button   int //used in 0

}

type Status_Struct struct {
	HallRequests [][]bool
	States       []State
}

type State struct {
	Behaviour   []int //change to enum-ish?
	Floor       []int
	Direction   []int
	CabRequests []bool
}

func Status(ElevStatus chan<- Status_struct, NewUpdate <-chan UpdateMsg, InternalUpdate chan<- UpdateMsg, ExternalUpdate <-chan UpdateMsg) {
	file, err := os.OpenFile("status.txt", os.O_RDWR, 0777)
	check(err)

	var status Status_Struct

	for {
		select {
			case message := <-ExternalUpdate:
				//update file with new message
				//update Status_Struct with new message
			case order := <-NewOrder:
				//give new order to network module
				//update file with new message (or wait for confirmation from network?)
				//update status struct with the new order

		}

	}

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
