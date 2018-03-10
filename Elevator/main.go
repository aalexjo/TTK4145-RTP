package main

import (
	"fmt"
	"flag"
	"./Driver/Elevio"
	"./Status"
	"./Network"
	"./Network/network/localip"
	"./Fsm"
	"os"
	"./Cost"
)


var FLOORS = 4
var ELEVATORS = 3

func main() {

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

	StatusUpdate := make(chan status.UpdateMsg) //sends updates that occured in the network to the status module
	NetworkUpdate := make(chan status.UpdateMsg)
	ElevStatus := make(chan status.StatusStruct)
	FSMinfo := make(chan cost.AssignedOrderInformation)
	StatusBroadcast := make(chan status.StatusStruct)
	StatusRefresh := make(chan status.StatusStruct)

	elevio.Init("localhost:15657", FLOORS)

	//parameters on the form (output,output,...,input,input,...,string,int,...)
	go network.Network(StatusUpdate, StatusRefresh, StatusBroadcast, NetworkUpdate, id)
	go status.Status(ElevStatus, StatusBroadcast, StatusRefresh, StatusUpdate, init, id)
	go fsm.Fsm(NetworkUpdate, FSMinfo, init, id)
	go cost.Cost(FSMinfo, ElevStatus)

	select{
	}

}

func AssignGlobals(){
	status.FLOORS = FLOORS
	status.ELEVATORS = ELEVATORS

}
