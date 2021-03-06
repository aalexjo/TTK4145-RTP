package cost

/*The Cost module utilizes the hall_request_assigner made by @klasbo for TTK4145 to calculate cost for each elevator
and then assigning all orders to the elevators accordingly. This module communicates with the Status and Fsm modules with which it respectively
receives and transmits its information. The status from the Status module is converted to JSON-format and the executable hall_request_assigner is run.
The result is then converted back and sent to the Fsm.*/
import (
	"encoding/json"
	"fmt"
	"os/exec"

	"../Status"
)

var FLOORS int
var ELEVATORS int
//var Mtx sync.Mutex = sync.Mutex{}

type AssignedOrderInformation struct {
	AssignedOrders map[string][][]bool
	HallRequests   [][2]bool
	States         map[string]*status.StateValues
}

func Cost(FSMinfo chan<- AssignedOrderInformation, ElevStatus <-chan status.StatusStruct) {
	for {
		select {
		case state := <-ElevStatus:
			status.Mtx.Lock()
			arg, err := json.Marshal(state)
			status.Mtx.Unlock()
			if err != nil {
				fmt.Println("error:", err)
			}
			result, err := exec.Command("sh", "-c", "./hall_request_assigner --input '"+string(arg)+"'").Output()
			if err != nil {
				fmt.Println("error:", err, "cost function")
				fmt.Println("recived:", string(arg))
				continue
			}
			orders := new(map[string][][]bool)
			json.Unmarshal(result, orders)

			status.Mtx.Lock()
			output := AssignedOrderInformation{
				AssignedOrders: *orders,
				HallRequests:   state.HallRequests,
				States:         state.States,
			}
			status.Mtx.Unlock()
			FSMinfo <- output
		}
	}

}
