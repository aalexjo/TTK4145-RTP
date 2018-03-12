package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"./Cost"
	"./Driver/Elevio"
	"./Fsm"
	"./Network"
	"./Network/network/localip"
	"./Status"
)

const FLOORS = 8
const ELEVATORS = 3

func main() {

	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	var init bool
	var port string
	flag.BoolVar(&init, "init", false, "true if elev is starting for first time")
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&port, "port", "15657", "set port to connect to elevator")
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

	AssignGlobals() //TODO: assign more globals

	StatusUpdate := make(chan status.UpdateMsg) //sends updates that occured in the network to the status module
	NetworkUpdate := make(chan status.UpdateMsg)
	ElevStatus := make(chan status.StatusStruct)
	FSMinfo := make(chan cost.AssignedOrderInformation)
	StatusBroadcast := make(chan status.StatusStruct)
	StatusRefresh := make(chan status.StatusStruct)

	elevio.Init("localhost:"+port, FLOORS)

	//parameters on the form (output,output,...,input,input,...,string,int,...)
	go atExit()
	go network.Network(StatusUpdate, StatusRefresh, StatusBroadcast, NetworkUpdate, id)
	go status.Status(ElevStatus, StatusBroadcast, StatusRefresh, StatusUpdate, init, id)
	go fsm.Fsm(NetworkUpdate, FSMinfo, init, id)
	go cost.Cost(FSMinfo, ElevStatus)

	select {}

}

func atExit() {
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	elevio.SetMotorDirection(elevio.MD_Stop)
	// do last actions and wait for all write operations to end

	log.Println("Program killed !")
	os.Exit(0)
}

func AssignGlobals() {
	status.FLOORS = FLOORS
	status.ELEVATORS = ELEVATORS
	fsm.FLOORS = FLOORS
}
