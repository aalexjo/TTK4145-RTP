package fsm

import (
  "../Driver/Elevio"
  //"fmt"
  //"os/exec"
  //"bytes"
  //"log"
  "../Status"
)

var FLOORS int
var ELEVATORS int
var BUTTONS int

type State struct {
  //behaviour   []int //change to enum-ish?
  floor       int
  //orders      [FLOORS][BUTTONS]int //[floor][btn] for all floors
  orders      [][]bool
  direction   elevio.MotorDirection
}

//var elev_state State;
//var elev_state status.State

//#TODO: allow communication between modules,



func Fsm(NetworkUpdate chan<- status.UpdateMsg, ElevStatus <-chan status.Status_struct){
  var updateMessage status.UpdateMsg
  numFloors := 4

  elevio.Init("localhost:15657", numFloors)

  in_buttons := make(chan elevio.ButtonEvent)
  in_floors  := make(chan int)
  

  go elevio.PollButtons(in_buttons)
  go elevio.PollFloorSensor(in_floors)


  for {
      select {
      case a := <- in_buttons:
        if a.Button < 2 { // If hall request
          updateMessage.MsgType = 0
          updateMessage.Button = a.Button;
          updateMessage.Floor = a.Floor;
        }
        else {
          updateMessage.MsgType = 4 //Cab request

          //TODO: Hvordan legge til hvilken elevator det er snakk om??
          updateMessage.Floor = a.Floor;
        }
        updateMessage.update[0] = //Floor og knapp
        NetworkUpdate <- updateMessage;

      case a := <- in_floors:
          updateMessage.MsgType = 2; //Arrived at floor
          updateMessage.Floor = a;

          //TODO: HVordan legge til hvilken elevator det er snakk om??
          NetworkUpdate <- updateMessage;
      }
  }

}



func requestsAbove(elev_state State) bool {
  for floor := elev_state.floor+1; floor < FLOORS; floor++ {
    for button := 0; button < BUTTONS; button++ {
      if elev_state.orders[floor][button] {
        return true
      }
    }
  }
  return false
}

func requestsBelow(elev_state State) bool {
  for floor := 0; floor < elev_state.floor; floor++ {
    for button := 0; button < BUTTONS; button++ {
      if elev_state.orders[floor][button] {
        return true
      }
    }
  }
  return false
}

//Choose direction of travel
func chooseDirection(elev_state State) elevio.MotorDirection {
  switch elev_state.direction {
  case elevio.MD_Up:
    if requestsAbove(elev_state) {
      return elevio.MD_Up
    } else if requestsBelow(elev_state) {
      return elevio.MD_Down
    } else {
      return elevio.MD_Stop
    }

  case elevio.MD_Down:
  case elevio.MD_Stop:
    if requestsBelow(elev_state) {
      return elevio.MD_Down
    } else if requestsAbove(elev_state) {
      return elevio.MD_Up
    } else {
      return elevio.MD_Stop
    }

  default:
    return elevio.MD_Stop
  }
  return elevio.MD_Stop
}

//Called when elevator reaches new floor, returns 1 if it should stop
func shouldStop(elev_state State) bool {
  switch elev_state.direction {
  case elevio.MD_Down:
    return elev_state.orders[elev_state.floor][elevio.BT_HallDown] || elev_state.orders[elev_state.floor][elevio.BT_Cab] || !requestsBelow(elev_state)

  case elevio.MD_Up:
    return elev_state.orders[elev_state.floor][elevio.BT_HallUp] || elev_state.orders[elev_state.floor][elevio.BT_Cab] || !requestsAbove(elev_state)

  case elevio.MD_Stop:
  default:
    return true
  }
  return true
}

//Clear order only if elevator is travelling in the right direction. RETURNS updated state!
func clearAtCurrentFloor(elev_state State) State{
    elev_state.orders[elev_state.floor][elevio.BT_Cab] = false
    switch elev_state.direction {

    case elevio.MD_Up:
      elev_state.orders[elev_state.floor][elevio.BT_HallUp] = false
      if !requestsAbove(elev_state) {
        elev_state.orders[elev_state.floor][elevio.BT_HallDown] = false
      }

    case elevio.MD_Down:
      elev_state.orders[elev_state.floor][elevio.BT_HallDown] = false
      if !requestsBelow(elev_state) {
        elev_state.orders[elev_state.floor][elevio.BT_HallUp] = false
      }

    case elevio.MD_Stop:
    default:
      elev_state.orders[elev_state.floor][elevio.BT_HallDown] = false
      elev_state.orders[elev_state.floor][elevio.BT_HallUp] = false
    }
    return elev_state
}
