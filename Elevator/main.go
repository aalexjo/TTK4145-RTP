package main

import (
	"flag"
	"./Driver/Elevio"
	"./Status"
	"./Network"
	"./Network/network/localip"
	//"./Fsm"
	"os"
	//"./cost"
)


var FLOORS = 4
var ELEVATORS = 3

func main() {

	/*----TEST ELEVATOR STRUCT-------------
	testElevator := fsm.State {
	Behaviour: "idle",
	Floor: 1,
	Direction: elevio.MD_Stop,
	Orders: [][]bool{{true, false, false},{false, false, false},{false, true, true},{false, true, false}},
	}
	//testNewOrderCh := make(chan fsm.NewOrder)
	/*----------------------------------*/

	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	var init bool
	flag.BoolVar(&init, "init", false, "true if elev is starting for first time")
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	AssignGlobals()//TODO: assign more globals

	StatusUpdate := make(chan Status.UpdateMsg) //sends updates that occured in the network to the status module
	NetworkUpdate := make(chan Status.UpdateMsg)
	ElevStatus := make(chan Status.StatusStruct)
	FSMinfo := make(chan cost.AssignedOrderInformation)

	elevio.Init("localhost:15657", FLOORS)

	go network.Network(StatusUpdate, NetworkUpdate, id)
	go status.Status(ElevStatus, StatusUpdate, init, id)
	go fsm.Fsm(NetworkUpdate, FSMinfo, init, id)
	go cost.Cost(FSMinfo, ElevStatus)

	for{
	}

}

func AssignGlobals(){
	status.FLOORS = FLOORS
	status.ELEVATORS = ELEVATORS

}
