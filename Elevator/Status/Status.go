package status

import (
	//"os"
	//"encoding/json"
	//"bufio"
	//"fmt"
)

var FLOORS int
var ELEVATORS int
/*
JSON format for saving the status
{
    "hallRequests" :
        [[Boolean, Boolean], ...],
    "States" :
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
	//copy from network module
	MsgType int
	// 0 = hallRequest
	// 1 = newBehaviour
	// 2 = arrivedAtfloor
	// 3 = newDirection
	// 4 = cabRequest
	// 5 = deleteElev
	Elevator string //used in all other than 0
	Floor    int //used in 0, 2, 4
	Button   int //used in 0, 4
	Behaviour string //used in 1
	Direction string //used in 3
	ServedOrder bool //used in 0, 4 - true if the elevator har completed an order and wants to clear it
}

type StatusStruct struct {
	HallRequests [][]bool `json:"hallRequests"`
	States       map[string]*StateValues `json:"states"` //key kan be changed to int if more practical but remember to cast to string before JSON encoding!
}

type StateValues struct {
	Behaviour   string `json:"behaviour"`
	Floor       int `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

func Status(ElevStatus chan<- StatusStruct, StatusBroadcast chan<- StatusStruct, StatusRefresh <-chan status.StatusStruct, StatusUpdate <-chan UpdateMsg, init bool, id string) {
	// ------------Commented out block until file is used-------------------
	//file, err := os.OpenFile("status.txt", os.O_RDWR, 0777)
	//check(err)
	//reader := bufio.NewReader(file)
	//------------------------------------------------------------------------*/
	status := new(StatusStruct)

	if init{ //clean initialization
		//file, err = os.Create("status.txt")
		//check(err)
		//reader := bufio.NewReader(file)

		status.HallRequests = [][]bool{{false,false},{false,false},{false,true},{false,false}}
		status.States = make(map[string]*StateValues)
		
		status.States[id] = new(StateValues)
		status.States[id].Behaviour = "idle"
		status.States[id].Floor = 0
		status.States[id].Direction = "stop"
		status.States[id].CabRequests = []bool{false,false,false,false}
			
	} else {//recover status from file
		//e := json.NewDecoder(reader).Decode(&status)
		//check(e)
	}

	for {
		select {
			case message := <-StatusUpdate:
				if message.Elevator != ""{
					if _, ok := status.States[message.Elevator]; !ok{//Elevator is not i status struct, initialized with best guess
						status.States[message.Elevator] = new(StateValues)
						status.States[message.Elevator].Behaviour = message.Behaviour
						status.States[message.Elevator].Floor = message.Floor
						status.States[message.Elevator].Direction = message.Direction
						status.States[message.Elevator].CabRequests = []bool{false,false,false,false}
					}
				}
				switch message.MsgType{
					case 0://hall request
						if message.ServedOrder{
							status.HallRequests[message.Floor][message.Button] = false
							//TODO write to file
						}else{
							status.HallRequests[message.Floor][message.Button] = true
							//TODO write to file
						}
						//fmt.Println(status)
					case 1://new Behaviour
						status.States[message.Elevator].Behaviour = message.Behaviour
						//TODO: write to file

					case 2://arrived at floor
						status.States[message.Elevator].Floor = message.Floor
						//TODO: write to file

					case 3://new direction
						status.States[message.Elevator].Direction = message.Direction
						//TODO: write to file

					case 4://cab request
						if message.ServedOrder {
							status.States[message.Elevator].CabRequests[message.Floor] = false
							//TODO write to file
						}else{
							status.States[message.Elevator].CabRequests[message.Floor] = true
							//TODO write to file
						}
					case 5:
						delete(status.States, message.Elevator)
				}
			case inputState <- StatusRefresh:



			case ElevStatus <- *status:
			case StatusBroadcast <- *status:

		}

	}

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
