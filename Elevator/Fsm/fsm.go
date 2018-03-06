package fsm

import (
  "../Driver/Elevio"
  "time"
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
  Behaviour   []int //change to enum-ish?
  Floor       int
  //Orders      [FLOORS][BUTTONS]int //[floor][btn] for all floors
  Orders      [][]bool
  Direction   elevio.MotorDirection
}

//var elev_state State;
//var elev_state status.State

//#TODO: allow communication between modules,


func Fsm(NetworkUpdate chan<- status.UpdateMsg, ElevStatus <-chan status.Status_struct, elev_state State){ // Remove elev_state and change to the channel?? (From status)
  var updateMessage status.UpdateMsg
  numFloors := 4

  //elevio.Init("localhost:15657", numFloors)

  in_buttons := make(chan elevio.ButtonEvent)
  in_floors  := make(chan int)

   //Channel for door timer. Bruk door_timed_out.reset(3 * time.Second) når døren åpnes
  door_timed_out := time.NewTimer(3 * time.Second)
  door_timed_out.stop()

  go elevio.PollButtons(in_buttons)
  go elevio.PollFloorSensor(in_floors)


  for {
      select {
        /*----------CASE for handling new orders. using a newOrder struct temporarily until
        final behaviour is decided.------------*/
      case newOrder := <- newOrderChannel:
        elev_state.Orders[newOrder.Floor][newOrder.button] = true
        elevio.SetButtonLamp(newOrder.button, newOrder.Floor, true)
        switch elev_state.Behaviour {

        case Idle: //Heisen står i ro
          elev_state.Direction = chooseDirection(elev_state)
          elevio.SetMotorDirection(elev_state.Direction)
          if elev_state.Direction == elevio.MD_stop { //Kan dette muligens skje mellom stasjoner???
            elevio.SetDoorOpenLamp(true)
            door_timed_out.reset(3 * time.Second)
            //elev_state.state = DOOR_OPEN???
            elev_state = clearAtCurrentFloor(elev_state)
            setAllLights(elev_state)
          } else {
            //elev_state.Behaviour = Moving
          }
        case Moving:
        case DoorOpen:
          if elevator.Floor == newOrder.Floor {
            door_timed_out.reset(3 * time.Second)
            elev_state = clearAtCurrentFloor(elev_state)
            setAllLights(elev_state)
          }

        }

      case a := <- in_buttons:
        /*----Making update message ----*/
        if a.Button < 2 { // If hall request
          updateMessage.MsgType = 0
          updateMessage.Button = a.Button;
          updateMessage.Floor = a.Floor;
        }else {
          updateMessage.MsgType = 4 //Cab request

          //TODO: Hvordan legge til hvilken elevator det er snakk om??
          updateMessage.Floor = a.Floor;
        }
        //updateMessage.update[0] = //Floor og knapp
        NetworkUpdate <- updateMessage;

        /*-----Handling Elevator stuff--------*/


      case elev_state.Floor = <- in_floors:
          updateMessage.MsgType = 2; //Arrived at floor
          updateMessage.Floor = elev_state.Floor;

          //TODO: HVordan legge til hvilken elevator det er snakk om?? eeeller trengs det nå?
          NetworkUpdate <- updateMessage;

          //Handling elevator stuff
          if shouldStop(elev_state) {
            elevio.SetMotorDirection(elevio.MD_Stop)
            elevio.SetDoorOpenLamp(true)
            door_timed_out.reset(3 * time.Second)
            //elev_state.state = DOOR_OPEN???
            elev_state = clearAtCurrentFloor(elev_state)
            setAllLights(elev_state)
            //^clears orders in the loca state variable. Needs to send an update message??
            //TODO: Send order cleared message to other peers
          }


      case a:= <- door_timed_out:
        elevio.SetDoorOpenLamp(false);
        elev_state.Direction = chooseDirection(elev_state)
        //TODO: Handle state/behaviour update??
        if(elev_state.Direction == elevio.MD_Stop) {
          //elev_state.state = IDLE
        } else{
          //elev_state.state = MOVING
          elevio.SetMotorDirection(elev_state.Direction)
        }
      }
  }

}



func requestsAbove(elev_state State) bool {
  for floor := elev_state.Floor+1; floor < FLOORS; floor++ {
    for button := 0; button < BUTTONS; button++ {
      if elev_state.Orders[floor][button] {
        return true
      }
    }
  }
  return false
}


func requestsBelow(elev_state State) bool {
  for floor := 0; floor < elev_state.floor; floor++ {
    for button := 0; button < BUTTONS; button++ {
      if elev_state.Orders[floor][button] {
        return true
      }
    }
  }
  return false
}


//Choose direction of travel
func chooseDirection(elev_state State) elevio.MotorDirection {
  switch elev_state.Direction {
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
  switch elev_state.Direction {
  case elevio.MD_Down:
    return elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] || elev_state.Orders[elev_state.Floor][elevio.BT_Cab] || !requestsBelow(elev_state)

  case elevio.MD_Up:
    return elev_state.Orders[elev_state.Floor][elevio.BT_HallUp] || elev_state.Orders[elev_state.Floor][elevio.BT_Cab] || !requestsAbove(elev_state)

  case elevio.MD_Stop:
  default:
    return true
  }
  return true
}


//Clear order only if elevator is travelling in the right direction. RETURNS updated state!
func clearAtCurrentFloor(elev_state State) State{
    elev_state.Orders[elev_state.Floor][elevio.BT_Cab] = false
    switch elev_state.Direction {

    case elevio.MD_Up:
      elev_state.Orders[elev_state.Floor][elevio.BT_HallUp] = false
      if !requestsAbove(elev_state) {
        elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] = false
      }

    case elevio.MD_Down:
      elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] = false
      if !requestsBelow(elev_state) {
        elev_state.Orders[elev_state.Floor][elevio.BT_HallUp] = false
      }

    case elevio.MD_Stop:
    default:
      elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] = false
      elev_state.Orders[elev_state.floor][elevio.BT_HallUp] = false
    }
    return elev_state
}


func setAllLights(elev_state State) {
  for floor := 0; floor < FLOORS; floor++ {
    for button := 0; button < BUTTONS; button++ {
      elevio.SetButtonLamp(button, floor, elev_state[floor][button])
    }
}
