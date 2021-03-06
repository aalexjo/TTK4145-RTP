package status

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

var FLOORS int
var ELEVATORS int
var Mtx sync.Mutex = sync.Mutex{}

/*
JSON format for saving the status
{
    "hallRequests" :
        [[Boolean, Boolean], ...],
    "states" :
        {
            "id_1" : {
                "behaviour"     : < "idle" | "moving" | "doorOpen" >
                "floor"         : NonNegativeInteger
                "direction"     : < "up" | "down" | "stop" >
                "cabRequests"   : [Boolean, ...]
            },
            "id_2" : {...}
        }
}
*/

type UpdateMsg struct {
	//encoding of relevant update information
	MsgType int
	// 0 = hallRequest
	// 1 = newBehaviour
	// 2 = arrivedAtfloor
	// 3 = newDirection
	// 4 = cabRequest
	// 5 = deleteElev
	// 8 = brokenMotor
	Elevator    string //used in all other than 0
	Floor       int    //used in 0, 2, 4
	Button      int    //used in 0, 4
	Behaviour   string //used in 1
	Direction   string //used in 3
	ServedOrder bool   //used in 0, 4, 6 - true if the elevator has completed an order and wants to clear it
}

type StatusStruct struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]*StateValues `json:"states"` //key kan be changed to int if more practical but remember to cast to string before JSON encoding!
}

type StateValues struct {
	Behaviour   string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

func Status(ElevStatus chan<- StatusStruct, StatusBroadcast chan<- StatusStruct, StatusRefresh <-chan StatusStruct, StatusUpdate <-chan UpdateMsg, init bool, id string) {

	file, err := os.OpenFile("status.txt", os.O_RDWR|os.O_CREATE, 0777)
	check(err)

	status := new(StatusStruct) //main status struct, continually updated
	Mtx.Lock()
	if init { //clean initialization
		file, err = os.Create("status.txt")
		check(err)

		status.HallRequests = make([][2]bool, FLOORS)
		status.States = make(map[string]*StateValues)
		initNewElevator(id, status, "idle", 0, "stop", make([]bool, FLOORS))
		file.Seek(0, 0)
		e := json.NewEncoder(file).Encode(status)
		check(e)

	} else { //recover status from file
		e := json.NewDecoder(file).Decode(status)
		check(e)
		fmt.Println("From file", status, "        ", status.States[id])
		arg, _ := json.Marshal(status)
		fmt.Println(string(arg))
	}
	Mtx.Unlock()
	for {
		select {
		case message := <-StatusUpdate:
			Mtx.Lock()
			if message.Elevator != "" {
				if _, ok := status.States[message.Elevator]; !ok && message.MsgType != 5 { //Elevator is not in status struct, initialized with best guess
					initNewElevator(message.Elevator, status, message.Behaviour, message.Floor, message.Direction, make([]bool, FLOORS))
				}
			}
			switch message.MsgType {
			case 0: //hall request
				if message.ServedOrder {
					status.HallRequests[message.Floor][message.Button] = false
				} else {
					status.HallRequests[message.Floor][message.Button] = true
				}
			case 1: //new Behaviour
				status.States[message.Elevator].Behaviour = message.Behaviour

			case 2: //arrived at floor
				status.States[message.Elevator].Floor = message.Floor

			case 3: //new direction
				status.States[message.Elevator].Direction = message.Direction

			case 4: //cab request
				if message.ServedOrder {
					status.States[message.Elevator].CabRequests[message.Floor] = false
				} else {
					status.States[message.Elevator].CabRequests[message.Floor] = true
				}
			case 5: //lost Elevator
				if message.Elevator != id { //dont delete ourselves
					delete(status.States, message.Elevator)
				}
			default:
				continue
			}
			Mtx.Unlock()
			file, err = os.Create("status.txt")
			check(err)
			e := json.NewEncoder(file).Encode(status)

			check(e)

		case inputState := <-StatusRefresh: //only add orders and update states
			//refresh hall requests
			Mtx.Lock()
			for floor := 0; floor < FLOORS; floor++ {
				for button := 0; button < 2; button++ {
					if inputState.HallRequests[floor][button] {
						status.HallRequests[floor][button] = true
					}
				}
			}

			for elev, estate := range status.States {
				if _, ok := status.States[elev]; !ok {
					initNewElevator(elev, status, estate.Behaviour, estate.Floor, estate.Direction, estate.CabRequests)
				} else {
					status.States[elev].Behaviour = estate.Behaviour
					status.States[elev].Floor = estate.Floor
					status.States[elev].Direction = estate.Direction
					for floor := 0; floor < FLOORS; floor++ {
						if estate.CabRequests[floor] {
							status.States[elev].CabRequests[floor] = true
						}
					}
				}
			}
			Mtx.Unlock()

		case ElevStatus <- *status:
		case StatusBroadcast <- *status:
		}
	}
}

//Used to initialize a new elevator in @param status with the gven paramters
func initNewElevator(elevName string, status *StatusStruct, Behaviour string, Floor int, Direction string, cabRequests []bool) {
	if elevName == "" {
		fmt.Println("invalid elevator name")
		return
	}
	if _, ok := status.States[elevName]; ok {
		fmt.Println("elevator ", elevName, " already initialized")
		return
	}
	status.States[elevName] = new(StateValues)
	if status.States[elevName].Behaviour == "" {
		status.States[elevName].Behaviour = "idle"
	} else {
		status.States[elevName].Behaviour = Behaviour
	}

	status.States[elevName].Floor = Floor
	if status.States[elevName].Direction == "" {
		status.States[elevName].Direction = "up"
	} else {
		status.States[elevName].Direction = Behaviour
	}
	if len(cabRequests) != FLOORS {
		status.States[elevName].CabRequests = make([]bool, FLOORS)
	} else {
		status.States[elevName].CabRequests = cabRequests
	}
	return
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
