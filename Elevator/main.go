package main

import (
	"./Driver/Elevio"
	"./Status"
	"./Network"


)


FLOORS := 4
ELEVATORS := 3

func main() {
	AssignGlobals()

	StatusUpdate := make(chan Status.UpdateMsg) //sends updates that occured in the network to the status module
	NetworkUpdate := make(chan Status.UpdateMsg)
	ElevStatus := make(chan Status.StatusStruct)
	HallRequests := make(chan cost.OrderStruct)
	elevio.Init("localhost:15657", FLOORS)

	go network.Network(StatusUpdate, NetworkUpdate)
	go status.Status(ElevStatus, StatusUpdate)
	go fsm.Fsm(NetworkUpdate, StatusUpdate)
	go cost.Cost(HallRequests, ElevStatus)
}

func AssignGlobals(){
	Status.FLOORS = FLOORS
	Status.FLOORS = FLOORS

}
