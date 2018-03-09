package cost

import (
	//"io"
	"os/exec"
	"fmt"
	"encoding/json"
	"../Status"
)

var FLOORS int
var ELEVATORS int


type AssignedOrderInformation struct{
	AssignedOrders map[string][][]bool
	HallRequests [][]bool
	States map[string]*status.StateValues
}



func Cost(FSMinfo chan<- AssignedOrderInformation, ElevStatus <-chan status.StatusStruct){
	for{
		select{
			case status:= <-ElevStatus:
					
					arg, err := json.Marshal(status)
					if err != nil {
						fmt.Println("error:", err)
					}

					result, err := exec.Command("sh", "-c", "./hall_request_assigner --input '"+string(arg)+"'").Output()

					if err != nil {
						fmt.Println("error:", err)
					}
					orders := new(map[string][][]bool)
					json.Unmarshal(result, orders)

					output := AssignedOrderInformation{
						AssignedOrders: *orders,
						HallRequests: status.HallRequests,
						States: status.States,
					}



					FSMinfo<-output
		}
	}

}
