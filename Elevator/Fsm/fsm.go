package fsm

//This package implements the basic operation of the elevator and builds on the elevator driver.
//The module does not retain any information (other than the name) about the state of the elevator,
//but expects the information recived from the cost module through the channel FSMinfo to be current.
//This information includes the elevators current state and the orders that have been assigned to it,
//all logic that is needed to execute these orders is contained within this module.
//All updates that occur to the elevator are transmitted to the network module and are further processed there.
//The different updatemessage types are described in the Status module.
import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"../Cost"
	"../Driver/Elevio"
	"../Status"
)

var FLOORS int
var ID string
var PORT string

//TODO: test motor failure. worries: behaviour=stop is not enough and elevator must disconnect from network
//TODO: @aalexjo clean up the motor failure part, not sure what needs to stay or not :)

func Fsm(NetworkUpdate chan<- status.UpdateMsg, FSMinfo <-chan cost.AssignedOrderInformation, init bool) {
	//Handling unexpected panic errors. Spawns a new process and initializing from a saved state.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r, " -FSM fatal panic, unable to recover. Rebooting...", "./main -init=false -port="+PORT+" -id="+ID)
			err := exec.Command("gnome-terminal", "-x", "sh", "-c", "./main -init=false -port="+PORT+" -id="+ID).Run() //if not running a compiled project add as the first paramters
			if err != nil {
				fmt.Println("Unable to reboot process, crashing...")
			}
		}
		os.Exit(0)
	}()

	//----initializing variables and subroutines ----//
	var updateMessage status.UpdateMsg
	var elevState cost.AssignedOrderInformation

	updateMessage.Elevator = ID
	elevState = <-FSMinfo //this variable contains all information the FSM needs to operate and is constantly refreshed
	buttonPress := make(chan elevio.ButtonEvent)
	newFloor := make(chan int)
	inFloor := make(chan int)

	doorTimedOut := time.NewTimer(3 * time.Second)
	doorTimedOut.Stop()
	motorTimedOut := time.NewTimer(4 * time.Second)
	motorTimedOut.Stop()

	go elevio.PollButtons(buttonPress)
	go elevio.PollFloorSensor(newFloor)
	go elevio.PollFloorSensorCont(inFloor)

	//----initializing elevator----//
	if init { //clean init
		elevio.SetMotorDirection(elevio.MD_Down)
		elevio.SetFloorIndicator(0)
		elevio.SetDoorOpenLamp(false)
		turnOffAllLights()
		updateMessage.Floor = 0
		updateMessage.Behaviour = "idle"
		updateMessage.Direction = "up"
	L: //label loop in order to break the for loop
		for {
			select {
			case floor := <-newFloor:
				if floor == 0 {
					elevio.SetMotorDirection(elevio.MD_Stop)
					break L
				}
			}
		}
	} else { //recovering from initialized system
		updateMessage.Floor = elevState.States[ID].Floor
		updateMessage.Behaviour = elevState.States[ID].Behaviour
		updateMessage.Direction = elevState.States[ID].Direction
		if elevState.States[ID].Behaviour == "doorOpen" {
			doorTimedOut.Reset(3 * time.Second)
		}
		if elevState.States[ID].Behaviour == "moving" {
			if elevState.States[ID].Direction == "up" {
				elevio.SetMotorDirection(elevio.MD_Up)
			} else {
				elevio.SetMotorDirection(elevio.MD_Down)
			}
		}
	}

	//Main loop in Fsm
	for {
		select {
		case elevState = <-FSMinfo: //New states from cost function
			setAllLights(elevState, ID)
			switch elevState.States[ID].Behaviour {
			case "doorOpen":
				if shouldStop(elevState, ID, elevState.States[ID].Floor) {
					if clearAtCurrentFloor(elevState, ID, elevState.States[ID].Floor, NetworkUpdate) {
						doorTimedOut.Reset(3 * time.Second)
					}
					setAllLights(elevState, ID)
				}
			case "idle":
				newDirection := chooseDirection(elevState, ID, elevState.States[ID].Floor)
				switch newDirection {
				case elevio.MD_Stop:
					if elevState.AssignedOrders[ID][elevState.States[ID].Floor][0] ||
						elevState.AssignedOrders[ID][elevState.States[ID].Floor][1] ||
						elevState.States[ID].CabRequests[elevState.States[ID].Floor] { //check if any button in current floor is pressed
						elevio.SetDoorOpenLamp(true)
						doorTimedOut.Reset(3 * time.Second)

						//Behaviour message
						updateMessage.MsgType = 1
						updateMessage.Behaviour = "doorOpen"
						updateMessage.Elevator = ID
						NetworkUpdate <- updateMessage

						clearAtCurrentFloor(elevState, ID, elevState.States[ID].Floor, NetworkUpdate)
						setAllLights(elevState, ID)
					}
				case elevio.MD_Up:
					elevio.SetMotorDirection(newDirection)
					motorTimedOut.Reset(4 * time.Second)

					//Direction Message
					updateMessage.MsgType = 3
					updateMessage.Direction = "up"
					updateMessage.Elevator = ID
					NetworkUpdate <- updateMessage

					//Behaviour message
					updateMessage.MsgType = 1
					updateMessage.Behaviour = "moving"
					updateMessage.Elevator = ID
					NetworkUpdate <- updateMessage
				case elevio.MD_Down:
					elevio.SetMotorDirection(newDirection)
					motorTimedOut.Reset(4 * time.Second)

					//Direction Message
					updateMessage.MsgType = 3
					updateMessage.Direction = "down"
					updateMessage.Elevator = ID
					NetworkUpdate <- updateMessage

					//Behaviour message
					updateMessage.MsgType = 1
					updateMessage.Behaviour = "moving"
					updateMessage.Elevator = ID
					NetworkUpdate <- updateMessage
				}
			case "moving":
			}
		case buttonEvent := <-buttonPress: //button press occurs, simply forward this to the network module
			/*------------Making update message ------------*/
			if buttonEvent.Button < 2 { // If hall request
				updateMessage.MsgType = 0
				updateMessage.Button = int(buttonEvent.Button)
				updateMessage.ServedOrder = false //Nytt knappetrykk
				updateMessage.Elevator = ID
				updateMessage.Behaviour = elevState.States[ID].Behaviour
				updateMessage.Floor = buttonEvent.Floor
			} else {
				updateMessage.MsgType = 4 //Cab request
				updateMessage.Floor = buttonEvent.Floor
				updateMessage.Button = int(buttonEvent.Button)
				updateMessage.Behaviour = elevState.States[ID].Behaviour
				updateMessage.ServedOrder = false //Nytt knappetrykk
				updateMessage.Elevator = ID
			}
			NetworkUpdate <- updateMessage
		case floor := <-newFloor: //We arrive at new floor, check if we should stop and clear orders
			/*--------------Message to send--------------------*/
			updateMessage.MsgType = 2 //Arrived at floor
			updateMessage.Floor = floor
			updateMessage.Behaviour = elevState.States[ID].Behaviour
			updateMessage.Elevator = ID
			NetworkUpdate <- updateMessage

			motorTimedOut.Stop()
			elevio.SetFloorIndicator(floor)
			if shouldStop(elevState, ID, floor) {
				elevio.SetMotorDirection(elevio.MD_Stop)
				clearAtCurrentFloor(elevState, ID, floor, NetworkUpdate)
				elevio.SetDoorOpenLamp(true)
				doorTimedOut.Reset(3 * time.Second)

				//Behaviour message
				updateMessage.MsgType = 1
				updateMessage.Behaviour = "doorOpen"
				updateMessage.Elevator = ID
				NetworkUpdate <- updateMessage

				setAllLights(elevState, ID)
			} else {
				motorTimedOut.Reset(4 * time.Second)
			}
		case <-doorTimedOut.C: //door closes and new direction is evaluated
			elevio.SetDoorOpenLamp(false)
			newDirection := chooseDirection(elevState, ID, elevState.States[ID].Floor)
			switch newDirection {
			case elevio.MD_Stop:
				//Behaviour message
				updateMessage.MsgType = 1
				updateMessage.Behaviour = "idle"
				updateMessage.Elevator = ID
				NetworkUpdate <- updateMessage

			case elevio.MD_Up:
				elevio.SetMotorDirection(newDirection)
				motorTimedOut.Reset(4 * time.Second)
				//Direction Message
				updateMessage.MsgType = 3
				updateMessage.Direction = "up"
				updateMessage.Elevator = ID
				NetworkUpdate <- updateMessage

				//Behaviour message
				updateMessage.MsgType = 1
				updateMessage.Behaviour = "moving"
				updateMessage.Elevator = ID
				NetworkUpdate <- updateMessage

			case elevio.MD_Down:
				elevio.SetMotorDirection(newDirection)
				motorTimedOut.Reset(4 * time.Second)
				//Direction Message
				updateMessage.MsgType = 3
				updateMessage.Direction = "down"
				updateMessage.Elevator = ID
				NetworkUpdate <- updateMessage

				//Behaviour message
				updateMessage.MsgType = 1
				updateMessage.Behaviour = "moving"
				updateMessage.Elevator = ID
				NetworkUpdate <- updateMessage

			}
		case <-motorTimedOut.C: //if the elevator does not detect a floor sensor within 4 seconds
			//all other operation is interrupted (this needs not be the case)
			currInfo := <-FSMinfo
			lastFloor := currInfo.States[ID].Floor
			updateMessage.MsgType = 8
			updateMessage.Elevator = ID
			fmt.Println("motor broke")
			NetworkUpdate <- updateMessage
			motorTimedOut.Reset(4 * time.Second)
			motorTimedOut.Stop()
			turnOffAllLights()

		F:
			for {
				select {
				case floor := <-inFloor:
					if floor != lastFloor {
						break F
					}
				}
			}

		}
	}
}

//Checks for any orders above current floor
func requestsAbove(elevState cost.AssignedOrderInformation, ID string, reachedFloor int) bool {
	for floor := reachedFloor + 1; floor < FLOORS; floor++ {
		if elevState.States[ID].CabRequests[floor] {
			return true
		}
		for button := 0; button < 2; button++ {
			if elevState.AssignedOrders[ID][floor][button] {
				return true
			}
		}
	}
	return false
}

//Checks any orders below current floor
func requestsBelow(elevState cost.AssignedOrderInformation, ID string, reachedFloor int) bool {
	for floor := 0; floor < reachedFloor; floor++ {
		if elevState.States[ID].CabRequests[floor] {
			return true
		}
		for button := 0; button < 2; button++ {
			if elevState.AssignedOrders[ID][floor][button] {
				return true
			}
		}
	}
	return false
}

//Choose direction of travel
func chooseDirection(elevState cost.AssignedOrderInformation, ID string, floor int) elevio.MotorDirection {
	switch elevState.States[ID].Direction {
	case "stop":
		fallthrough
	case "down":
		if requestsBelow(elevState, ID, floor) {
			return elevio.MD_Down
		} else if requestsAbove(elevState, ID, floor) {
			return elevio.MD_Up
		} else {
			return elevio.MD_Stop
		}
	case "up":
		if requestsAbove(elevState, ID, floor) {
			return elevio.MD_Up
		} else if requestsBelow(elevState, ID, floor) {
			return elevio.MD_Down
		} else {
			return elevio.MD_Stop
		}
	default:
		return elevio.MD_Stop
	}
	return elevio.MD_Stop
}

//Called when elevator reaches new floor, returns true if it should stop
func shouldStop(elevState cost.AssignedOrderInformation, ID string, floor int) bool {
	switch elevState.States[ID].Direction {
	case "down":
		return (elevState.AssignedOrders[ID][floor][elevio.BT_HallDown] ||
			elevState.States[ID].CabRequests[floor] ||
			!requestsBelow(elevState, ID, floor))
	case "up":
		return (elevState.AssignedOrders[ID][floor][elevio.BT_HallUp] ||
			elevState.States[ID].CabRequests[floor] ||
			!requestsAbove(elevState, ID, floor))
	case "stop":
		fallthrough
	default:
		return true
	}
}

//Clear order only if elevator is travelling in the right direction. Returns true if an order has been cleared.
func clearAtCurrentFloor(elevState cost.AssignedOrderInformation, ID string, floor int, NetworkUpdate chan<- status.UpdateMsg) bool {
	//For cabRequests
	cleared := false
	update := status.UpdateMsg{
		MsgType:     4,
		Floor:       floor,
		Button:      2,
		Behaviour:   elevState.States[ID].Behaviour,
		Direction:   elevState.States[ID].Direction,
		ServedOrder: true,
		Elevator:    ID,
	}
	if elevState.States[ID].CabRequests[elevState.States[ID].Floor] {
		NetworkUpdate <- update
		cleared = true
	}

	//For hallRequests
	update.MsgType = 0
	switch elevState.States[ID].Direction {
	case "up":
		if elevState.HallRequests[elevState.States[ID].Floor][int(elevio.BT_HallUp)] {
			update.Button = int(elevio.BT_HallUp)
			NetworkUpdate <- update
			cleared = true
		}
		if !requestsAbove(elevState, ID, floor) &&
			elevState.HallRequests[elevState.States[ID].Floor][int(elevio.BT_HallDown)] {
			update.Button = int(elevio.BT_HallDown)
			NetworkUpdate <- update
			cleared = true
		}
	case "down":
		if elevState.HallRequests[elevState.States[ID].Floor][int(elevio.BT_HallDown)] {
			update.Button = int(elevio.BT_HallDown)
			NetworkUpdate <- update
			cleared = true
		}
		if !requestsBelow(elevState, ID, floor) &&
			elevState.HallRequests[elevState.States[ID].Floor][int(elevio.BT_HallUp)] {
			update.Button = int(elevio.BT_HallUp)
			NetworkUpdate <- update
			cleared = true//Attaching a timer to each ack message.
}
		}
	case "stop":
		update.Button = int(elevio.BT_HallDown)
		NetworkUpdate <- update
		update.Button = int(elevio.BT_HallUp)
		NetworkUpdate <- update
		cleared = true
	}
	return cleared

}

//Updates all lights based on all orders from Cost module
func setAllLights(elevState cost.AssignedOrderInformation, ID string) {
	for floor := 0; floor < FLOORS; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, elevState.States[ID].CabRequests[floor])
		for button := elevio.BT_HallUp; button < elevio.BT_Cab; button++ {
			elevio.SetButtonLamp(button, floor, elevState.HallRequests[floor][button])
		}
	}
}

func turnOffAllLights() {
	for floor := 0; floor < FLOORS; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
		for button := elevio.BT_HallUp; button < elevio.BT_Cab; button++ {
			elevio.SetButtonLamp(button, floor, false)
		}
	}
}
