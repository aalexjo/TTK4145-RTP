package status

import (
	"os"
)

type UpdateMsg struct {
	//encoding of relevant update information
	//copy from network module
	msgType int
	// 0 = hallRequest
	// 1 = behaviour
	// 2 = arrivedAtfloor
	// 3 = newDirection
	// 4 = cabRequest
	elevator int //used in all other than 0
	floor    int //used in 0,2,4
	button   int //used in 0

}

type Status_Struct struct {
	hallRequests [][]bool
	states       []state
}

type State struct {
	behaviour   []int //change to enum-ish?
	floor       []uint
	direction   []int
	cabRequests []bool
}

func Status(InternalUpdate chan<- UpdateMsg, ExternalUpdate <-chan UpdateMsg) {
	file, err := os.OpenFile("status.txt", os.O_RDWR, 0777)
	check(err)

	var status Status_Struct

	for {
		select {
		case message := <-ExternalUpdate:

		}
	}

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
