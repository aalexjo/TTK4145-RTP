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


	init := os.Args[1]

	/*----TEST ELEVATOR STRUCT-------------*/
	testElevator := fsm.State {
	Behaviour: "idle",
	Floor: 1,
	Direction: elevio.MD_Stop,
	Orders: [][]bool{{true, false, false},{false, false, false},{false, true, true},{false, true, false}},
	}
	testNewOrderCh := make(chan fsm.NewOrder)
	/*----------------------------------*/
	
	ElevID, err := localIP()[12:] //denne funker obviously ikke? kaller dypt nde i network mappestrukturen

	AssignGlobals()//TODO: assign more globals

	StatusUpdate := make(chan Status.UpdateMsg) //sends updates that occured in the network to the status module
	NetworkUpdate := make(chan Status.UpdateMsg)
	ElevStatus := make(chan Status.StatusStruct)
	FSMinfo := make(chan cost.AssignedOrderInformation)

	elevio.Init("localhost:15657", FLOORS)

	go network.Network(StatusUpdate, NetworkUpdate)
	go status.Status(ElevStatus, StatusUpdate, init)
	go fsm.Fsm(NetworkUpdate, FSMinfo, init, elevID)
	go cost.Cost(FSMinfo, ElevStatus)
	
	for{	
	}

}

func AssignGlobals(){
	status.FLOORS = FLOORS
	status.ELEVATORS = ELEVATORS

}
