package cost

import (

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


func Cost((HallRequests chan<- status.UpdateMsg, ElevStatus <-chan status.Status_struct)){
	for{
		select{
			case status:= <-ElevStatus:

		}
	}

}