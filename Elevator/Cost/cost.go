package cost

import (
	"io"
	"os"
	"../Status"
)

var FLOORS int
var ELEVATORS int

type Order_Struct struct{
	Elevator []HallReq_Struct
}

type HallReq_Struct struct{
	orders [][]bool
}

type Status_Struct struct {
	HallRequests [][]bool `json:"hallRequests"`
	States       State `json:"states"`
}

type State struct{
	One State_Values
	Two State_Values
}

type State_Values struct {
	Behaviour   string `json:"behaviour"`
	Floor       uint `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequest"`
}


func Cost((HallRequests chan<- status.UpdateMsg, ElevStatus <-chan status.Status_struct)){
	for{
		select{
			case status:= <-ElevStatus:
					b, err := json.Marshal(status)
					if err != nil {
						fmt.Println("error:", err)
					}

					c := exec.Command("cmd", "")
    				stdin, err := c.StdinPipe()
    				if err != nil { 
        				fmt.Println("Error: ", err)
    				}	 

					io.WriteString(stdin,b)


		}
	}

}