package main

//This is the entry point for the elevator project in TTK4145 Real time programming, made by Alexander Johansen and Bendik Standal.
//The project consists of five modules tied together in this main package. The modules communicate through go channels according to the design
//diagram found in the Design section of the project on github. The modules are: Cost, Fsm, Status, Network and Driver.
//The Driver module contains all functions necessary to run/drive the physical elevator. The Cost module assigns all hall request orders based on a cost function
//which utilises the information about every elevator, e.g. state, baheviour, cab orders, and determines which elevator should execute every order.
//The Cost module communicates with the Status and Fsm modules, by using the state information from the Status module and passing the resulting order
//assignments to the Fsm. The Fsm module uses the Driver to run the elevator based on the information it receives from the Cost module,
//sets lights, reads sensors and ultimately sends everything to the Network module. The Status module contains the data for all peers in the Network
//and sends this information to the Cost module. The Status module also receives all information from the Network module. The Network module consists of one information
//and several submodules. The main module is the top layer for communicating between this peer itself in addition to other peers. The acknowledge submodule contains
//the ack-logic for this system. It gives all messages sent a sequence number and sends acks for all messages received.
import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

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
	time.Sleep(400 * time.Millisecond)
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	var init bool
	var port string
	flag.BoolVar(&init, "init", true, "false if elev is recovering")
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&port, "port", "15657", "set port to connect to elevator")
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r, " MAIN fatal panic, unable to recover. Rebooting...", "go run main.go -init=false -port="+port, " -id=", id)
			err := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run main.go -init=false -port="+port, " -id="+id).Run()
			if err != nil {
				fmt.Println("Unable to reboot process, crashing...")
			}
		}
		os.Exit(0)
	}()

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
	go fsm.Fsm(NetworkUpdate, FSMinfo, init, id, port)
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
