package fsm

import (
  "../Driver/Elevio"
  "time"
  "fmt"
  //"os/exec"
  //"bytes"
  //"log"
  "../Status"
)

const FLOORS = 4
const ELEVATORS = 1
const BUTTONS = 3

type State struct {
  Behaviour   string
  Floor       int
  //Orders      [FLOORS][BUTTONS]int //[floor][btn] for all floors
  Orders      [][]bool
  Direction   elevio.MotorDirection
}

type NewOrder struct {
  Floor int
  Button elevio.ButtonType
}

//var elev_state State;
//var elev_state status.State

//#TODO: allow communication between modules, Add state functionality some
//places and change to Status' state struct?


func Fsm(NetworkUpdate chan<- status.UpdateMsg, ElevStatus <-chan status.Status_Struct, elev_state State, newOrderChannel chan NewOrder){ // Remove elev_state and change to the channel?? (From status)
  var updateMessage status.UpdateMsg

  in_buttons := make(chan elevio.ButtonEvent)
  in_floors  := make(chan int)

   //Channel for door timer. Bruk door_timed_out.reset(3 * time.Second) når døren åpnes
  door_timed_out := time.NewTimer(3 * time.Second)
  door_timed_out.Stop()

  go elevio.PollButtons(in_buttons)
  go elevio.PollFloorSensor(in_floors)

  //TESTGREIE:--------
  //elevio.SetMotorDirection(elevio.MD_Up)
  //TESTGREIE SLUTT---------


  for {
      select {
        /*----------CASE for handling new orders. using a newOrder struct temporarily until
        final behaviour is decided.------------*/
      case newOrder := <- newOrderChannel:
        elev_state.Orders[newOrder.Floor][int(newOrder.Button)] = true
        elevio.SetButtonLamp(newOrder.Button, newOrder.Floor, true)
        switch elev_state.Behaviour {

        case "idle": //Heisen står i ro
          elev_state.Direction = chooseDirection(elev_state)
          elevio.SetMotorDirection(elev_state.Direction)
          if elev_state.Direction == elevio.MD_Stop { //Kan dette muligens skje mellom stasjoner???
            elevio.SetDoorOpenLamp(true)
            door_timed_out.Reset(3 * time.Second)
            elev_state.Behaviour = "doorOpen"
            elev_state = clearAtCurrentFloor(elev_state)
            setAllLights(elev_state)
          } else {
            elev_state.Behaviour = "moving"
          }
        case "moving":
        case "doorOpen":
          if elev_state.Floor == newOrder.Floor {
            door_timed_out.Reset(3 * time.Second)
            elev_state = clearAtCurrentFloor(elev_state)
            setAllLights(elev_state)
          }

        }

      case a := <- in_buttons:
        /*----Making update message ----*/
        if a.Button < 2 { // If hall request
          updateMessage.MsgType = 0
          updateMessage.Button = int(a.Button);
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

          fmt.Println("Reached floor: ", elev_state.Floor)
          //TODO: HVordan legge til hvilken elevator det er snakk om?? eeeller trengs det nå?
          //NetworkUpdate <- updateMessage;

          //Handling elevator stuff
          if shouldStop(elev_state) {
            elevio.SetMotorDirection(elevio.MD_Stop)
            elevio.SetDoorOpenLamp(true)
            door_timed_out.Reset(3 * time.Second)
            elev_state.Behaviour = "doorOpen"
            elev_state.Direction = chooseDirection(elev_state)
            elev_state = clearAtCurrentFloor(elev_state)
            setAllLights(elev_state)
            //^clears orders in the loca state variable. Needs to send an update message??
            //TODO: Send order cleared message to other peers
          }


      case <- door_timed_out.C:
        elevio.SetDoorOpenLamp(false);
        /*fmt.Println("Old direction is: ", elev_state.Direction)
        elev_state.Direction = chooseDirection(elev_state)
        fmt.Println("New direction is: ", elev_state.Direction)
        */
        if(elev_state.Direction == elevio.MD_Stop) {
          elev_state.Behaviour = "idle"
        } else{
          elev_state.Behaviour = "moving"
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
  for floor := 0; floor < elev_state.Floor; floor++ {
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

  case elevio.MD_Stop:
    fallthrough
  case elevio.MD_Down:
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
    fallthrough
  default:
    return true
  }
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
      fallthrough
    default:
      elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] = false
      elev_state.Orders[elev_state.Floor][elevio.BT_HallUp] = false
    }
    return elev_state
}


func setAllLights(elev_state State) {
  for floor := 0; floor < FLOORS; floor++ {
    for button := elevio.BT_HallUp; button <= elevio.BT_Cab; button++ {
      elevio.SetButtonLamp(button, floor, elev_state.Orders[floor][button])
    }
  }
}
