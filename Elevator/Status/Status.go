package status

import (
	//"os"
	//"encoding/json"
	//"bufio"
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
	Elevator string //used in all other than 0
	Floor    uint //used in 0,2,4
	Button   int //used in 0
	Behaviour string //used in 1
	Direction string //used in 3
	ServedOrder bool //used in 0, 4 - true if the elevator har completed an order and wants to clear it
}

type StatusStruct struct {
	HallRequests [][]bool `json:"hallRequests"`
	States       map[string]StateValues `json:"states"` //key kan be changed to int if more practical but remember to cast to string before JSON encoding!
}

type StateValues struct {
	Behaviour   string `json:"behaviour"`
	Floor       uint `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequest"`
}

func Status(ElevStatus chan<- Status_Struct, StatusUpdate <-chan UpdateMsg, init bool) {
	/* ------------Commented out block until file is used-------------------
	file, err := os.OpenFile("status.txt", os.O_RDWR, 0777)
	check(err)
	reader := bufio.NewReader(f)
	------------------------------------------------------------------------*/

	if init{ //clean initialization
		
	status := new(StatusStruct) //todo: initilize with zero values and refresh file
	
	status.HallRequests = [][]bool{{false,false},{false,false},{false,false},{false,false}}
	status.States = map[string]StateValues{
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
		
	

	} else {//recover status from file
		status := StatusStruct
		e := json.NewDecoder(reader).Decode(&status)
		check(e)
	}

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

	for {
		select {
			case message := <-StatusUpdate:
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
						//	-check if elevator exists in status struct, handle if not

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
