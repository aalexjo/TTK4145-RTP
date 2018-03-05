package runElevator
//Module for running the individual elevators
//Maybe find a better name?
//Calculates what to do based on the status register for the individual elevator

import (
  "../Driver/elevio"
  //"../Status"
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
