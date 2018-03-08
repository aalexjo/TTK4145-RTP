package status

import (
	"os"
	"encoding/json"
	"bufio"
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

/*
format of status.txt
0010 	//bolean values for hall requests
idle 	//behaviour of elev 1
2		//floor of elev 1
stop	//direction of elev 1
0000	//cab requests of elev 1
moving	//behaviour of elev 2
2		//floor of elev 2
up 		//direction of elev 2
0010 	//cab Requests in elev 2
idle	//behaviour of elev 3
.
.
.
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
	// 5 = newElev
	// 6 = deleteElev
	Elevator string //used in all other than 0
	Floor    int //used in 0,2,4,5
	Button   int //used in 0
	Behaviour string //used in 1, 5
	Direction string //used in 3, 5
	ServedOrder bool //used in 0, 4 - true if the elevator har completed an order and wants to clear it
}

type StatusStruct struct {
	HallRequests [][]bool `json:"hallRequests"`
	States       map[string]StateValues `json:"states"` //key kan be changed to int if more practical but remember to cast to string before JSON encoding!
}

type StateValues struct {
	Behaviour   string `json:"behaviour"`
	Floor       int `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequest"`
}
/*-------------HOW TO INITIALIZE STATUS---------------
status := new(StatusStruct) //todo: initilize with correct values
status.HallRequests = [][]bool{{false,false},{false,false},{false,false},{false,false}}
status.States = map[string]State_Values{
	"One": {
		Behaviour: "moving",
		Floor: 2,
		Direction: "up",
		CabRequests: []bool{false,false,false,true},
		},
	"Two": {
		Behaviour: "moving",
		Floor: 2,
		Direction: "up",
		CabRequests: []bool{false,false,false,true},
		},
	}
------------------------------------------------------*/

func Status(ElevStatus chan<- StatusStruct, StatusUpdate <-chan UpdateMsg, init bool, id string) {
	// ------------Commented out block until file is used-------------------
	file, err := os.OpenFile("status.txt", os.O_RDWR, 0777)
	check(err)
	reader := bufio.NewReader(file)
	//------------------------------------------------------------------------*/
	status := new(StatusStruct) //HAd to move this up here??

	if init{ //clean initialization
		file, err = os.Create("status.txt")
		check(err)
		status.HallRequests = [][]bool{{false,false},{false,false},{false,false},{false,false}}
		status.States = map[string]State_Values{
			id: {
				Behaviour: "idle",
				Floor: 0,
				Direction: "stop",
				CabRequests: []bool{false,false,false,false},
				},
	} else {//recover status from file
		e := json.NewDecoder(reader).Decode(&status)
		check(e)
	}

	for {
		select {
			case message := <-StatusUpdate:
				//	-check if elevator exists in status struct, handle if not
				switch message.MsgType{
					case 0://hall request
						if message.ServedOrder{
							status.HallRequests[message.Floor][message.Button] = false
							//TODO write to file
						}else{
							status.HallRequests[message.Floor][message.Button] = true
							//TODO write to file
						}

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
					case 5: //init message
					/*
							new_status.States[message.Elevator] = StateValues{
							Behaviour: message.Behaviour,
							Floor: message.Floor,
							Direction: message.Direction,
							CabRequests: []bool{false,false,false,false},
						}*/
					case 6:
						delete(status.States, message.Elevator)
				}
			case ElevStatus <- status:
		}

	}

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
