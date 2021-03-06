package main

/*This is the entry point for the elevator project in TTK4145 Real time programming.
The project consists of five modules tied together in this main package. The modules communicate through go channels according to the design
diagram found in the Design section of the project on github. https://github.com/TTK4145/project-merge-issues
The modules are: Cost, Fsm, Status, Network and Driver. Their communication and
further function description can be found in the README file.*/
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
	"./Network/network/acknowledge"
	"./Network/network/localip"

	"./Status"
)

//Change according to number of elevators. Could also be passed from command line.
const FLOORS = 4
const ELEVATORS = 3

var id string
var port string

func main() {
	time.Sleep(400 * time.Millisecond)
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var init bool
	flag.BoolVar(&init, "init", true, "false if elev is recovering")
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

	//Functionality for handling unexpected panic errors. Spawns another process and initializes the elevator from the previosly saved state. Also checks that elevator server is running
	defer func() {
		if r := recover(); r != nil {
			if r == "dial tcp 127.0.0.1:"+port+": getsockopt: connection refused" {
				err := exec.Command("gnome-terminal", "-x", "sh", "-c", "ElevatorServer").Run() //"../Simulator/SimElevatorServer --port="+port).Run()

				if err != nil {
					fmt.Println("Unable to reboot process, crashing...")
				}
			}
			fmt.Println(r, " MAIN fatal panic, unable to recover. Rebooting...", "./main -init=false -port="+port+" -id="+id)
			err := exec.Command("gnome-terminal", "-x", "sh", "-c", "./main -init=false -port="+port+" -id="+id).Run()
			if err != nil {
				fmt.Println("Unable to reboot process, crashing...")
			}
		}
		os.Exit(0)
	}()

	AssignGlobals(id, port)

	StatusUpdate := make(chan status.UpdateMsg) //sends updates that occured in the network to the status module
	NetworkUpdate := make(chan status.UpdateMsg)
	ElevStatus := make(chan status.StatusStruct)
	FSMinfo := make(chan cost.AssignedOrderInformation)
	StatusBroadcast := make(chan status.StatusStruct)
	StatusRefresh := make(chan status.StatusStruct)

	elevio.Init("localhost:"+port, FLOORS)

	//parameters on the form (output,output,...,input,input,...,other parameters)
	go atExit()
	go network.Network(StatusUpdate, StatusRefresh, StatusBroadcast, NetworkUpdate, id)
	go status.Status(ElevStatus, StatusBroadcast, StatusRefresh, StatusUpdate, init, id)
	go fsm.Fsm(NetworkUpdate, FSMinfo, init)
	go cost.Cost(FSMinfo, ElevStatus)

	select {}
}

//Ensures a smooth shutdown when program is killed from terminal. Currently it restarts the program 
func atExit() {
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	elevio.SetMotorDirection(elevio.MD_Stop)
	// do last actions and wait for all write operations to end
	log.Println("Rebooting", "sh", "-c", "./main -init=false -port="+port+" -id="+id)
	err := exec.Command("gnome-terminal", "-x", "sh", "-c", "./main -init=false -port="+port+" -id="+id).Run()
	if err != nil {
		fmt.Println("Unable to reboot process, crashing...")
	}
	log.Println("Program killed !")
	os.Exit(0)
}

func AssignGlobals(id string, port string) {
	status.FLOORS = FLOORS
	status.ELEVATORS = ELEVATORS
	fsm.FLOORS = FLOORS
	fsm.ID = id
	fsm.PORT = port
	acknowledge.ID = id
	acknowledge.PORT = port
}
