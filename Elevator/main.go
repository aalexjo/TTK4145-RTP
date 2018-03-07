package main

import (
	"./Driver/Elevio"
	"./Status"
	//"./Network"
	"./Fsm"
	//"./cost"
)


var FLOORS = 4
var ELEVATORS = 3

func main() {

/*----TEST ELEVATOR STRUCT-------------*/
	testElevator := fsm.State {
	Behaviour: "idle",
	Floor: 1,
	Direction: elevio.MD_Stop,
	Orders: [][]bool{{true, false, false},{false, false, false},{false, true, true},{false, true, false}},
	}
	testNewOrderCh := make(chan fsm.NewOrder)
	/*----------------------------------*/
	AssignGlobals()

	//StatusUpdate := make(chan status.UpdateMsg) //sends updates that occured in the network to the status module
	NetworkUpdate := make(chan status.UpdateMsg)
	ElevStatus := make(chan status.Status_Struct)
	//HallRequests := make(chan cost.Order_Struct)
	elevio.Init("localhost:15657", FLOORS)

	//go network.Network(StatusUpdate, NetworkUpdate)
	//go status.Status(ElevStatus, StatusUpdate)
	//go fsm.Fsm(NetworkUpdate, StatusUpdate)
	go fsm.Fsm(NetworkUpdate, ElevStatus, testElevator, testNewOrderCh)
	//go cost.Cost(HallRequests, ElevStatus)
	for{
		
	}

}

func AssignGlobals(){
	status.FLOORS = FLOORS
	status.ELEVATORS = ELEVATORS

}
