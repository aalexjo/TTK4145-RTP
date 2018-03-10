package cost

import (
	//"io"
	"encoding/json"
	"fmt"
	"os/exec"

	"../Status"
)

var FLOORS int
var ELEVATORS int

type AssignedOrderInformation struct {
	AssignedOrders map[string][][]bool
	HallRequests   [][2]bool
	States         map[string]*status.StateValues
}

func Cost(FSMinfo chan<- AssignedOrderInformation, ElevStatus <-chan status.StatusStruct) {
	for {
		select {
		case status := <-ElevStatus:
			//fmt.Println("Status inn :", status)

			arg, err := json.Marshal(status)
			if err != nil {
				fmt.Println("error:", err)
			}
			fmt.Println("Marshaled: ", string(arg))

			result, err := exec.Command("sh", "-c", "./hall_request_assigner --input '"+string(arg)+"'").Output()

			//fmt.Println("Result fra command: ", string(result))

			if err != nil {
				fmt.Println("error:", err)
			}
			orders := new(map[string][][]bool)
			json.Unmarshal(result, orders)

			output := AssignedOrderInformation{
				AssignedOrders: *orders,
				HallRequests:   status.HallRequests,
				States:         status.States,
			}
			//fmt.Println("output fra cost: ", output)
			FSMinfo <- output
		}
	}

}
