package fsm

import (
  "../Driver/Elevio"
  "time"
  "fmt"
  "../Status"
)

const FLOORS = 4
const ELEVATORS = 1
const BUTTONS = 3

type State struct {
  Behaviour   string
  Floor       int
  Orders      [][]bool
  Direction   elevio.MotorDirection
}

type NewOrder struct {
  Floor int
  Button elevio.ButtonType
}

//#TODO: allow communication between modules, add information about which elevator i am?
//#TODO:Add state functionality some
//#TODO:places and change to Status' state struct?


func Fsm(NetworkUpdate chan<- status.UpdateMsg, ElevStatus <-chan status.Status_Struct, elev_state State, newOrderChannel chan NewOrder){ // Remove elev_state and change to the channel?? (From status)
  var updateMessage status.UpdateMsg

  in_buttons := make(chan elevio.ButtonEvent)
  in_floors  := make(chan int)

   //Channel for door timer. Bruk door_timed_out.reset(3 * time.Second) når døren åpnes
  door_timed_out := time.NewTimer(3 * time.Second)
  door_timed_out.Stop()

  go elevio.PollButtons(in_buttons)
  go elevio.PollFloorSensor(in_floors)


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
            elev_state = clearAtCurrentFloor(elev_state, NetworkUpdate)
            setAllLights(elev_state)
          } else {
            elev_state.Behaviour = "moving"
          }
        case "moving":
        case "doorOpen":
          if elev_state.Floor == newOrder.Floor {
            door_timed_out.Reset(3 * time.Second)
            elev_state = clearAtCurrentFloor(elev_state, NetworkUpdate)
            setAllLights(elev_state)
          }

        }

      case buttonEvent := <- in_buttons:
        /*------------Making update message ------------*/
        if buttonEvent.Button < 2 { // If hall request
          updateMessage.MsgType = 0
          updateMessage.Button = int(buttonEvent.Button)
          updateMessage.Floor = buttonEvent.Floor
        }else {
          updateMessage.MsgType = 4 //Cab request
          updateMessage.Floor = buttonEvent.Floor
          updateMessage.ServedOrder = false //Nytt knappetrykk
          //TODO: Hvordan legge til hvilken elevator det er snakk om??
        }
        //NetworkUpdate <- updateMessage;

        /*-------------Handling Elevator stuff---------------*/
        //TODO:Eventuelt STOPP-knapp behandling eller noe?
        //TODO:Eventuelt hvis vi skal akseptere cabRequests umiddelbart??

      case elev_state.Floor = <- in_floors:
        /*--------------Message to send--------------------*/
          updateMessage.MsgType = 2; //Arrived at floor
          updateMessage.Floor = elev_state.Floor;
          //updateMessage.Elevator = ?????

          fmt.Println("Reached floor: ", elev_state.Floor)
          //TODO: HVordan legge til hvilken elevator det er snakk om?? eeeller trengs det nå?
          //NetworkUpdate <- updateMessage;

          if shouldStop(elev_state) {
            elevio.SetMotorDirection(elevio.MD_Stop)

            //Stop message
            updateMessage.MsgType = 3
            updateMessage.Direction = "stop"
            //updateMessage.Elevator = ???
            //NetworkUpdate <- updateMessage

            elevio.SetDoorOpenLamp(true)
            door_timed_out.Reset(3 * time.Second)

            //Behaviour message
            updateMessage.MsgType = 1
            updateMessage.Behaviour = "doorOpen"
            //updateMessage.Elevator = ???
            //NetworkUpdate <- updateMessage

            elev_state.Behaviour = "doorOpen"
            elev_state = clearAtCurrentFloor(elev_state, NetworkUpdate)
            setAllLights(elev_state)
          }

      case <- door_timed_out.C:
        elevio.SetDoorOpenLamp(false);
        elev_state.Direction = chooseDirection(elev_state)
        switch elev_state.Direction {
        case elevio.MD_Stop:
          elev_state.Behaviour = "idle"

          //Behaviour message
          updateMessage.MsgType = 1
          updateMessage.Behaviour = "idle"
          //updateMessage.Elevator = ???
          //NetworkUpdate <- updateMessage

        case elevio.MD_Up:
          elev_state.Behaviour = "moving"
          elevio.SetMotorDirection(elev_state.Direction)

          //Behaviour message
          updateMessage.MsgType = 1
          updateMessage.Behaviour = "moving"
          //updateMessage.Elevator = ???
          //NetworkUpdate <- updateMessage

        case elevio.MD_Down:
          elev_state.Behaviour = "moving"
          elevio.SetMotorDirection(elev_state.Direction)

          //Behaviour message
          updateMessage.MsgType = 1
          updateMessage.Behaviour = "moving"
          //updateMessage.Elevator = ???
          //NetworkUpdate <- updateMessage

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
func clearAtCurrentFloor(elev_state State, NetworkUpdate chan<- status.UpdateMsg) State{
    //For cabRequests
    update := status.UpdateMsg {
      MsgType: 4,
      Floor: elev_state.Floor,
      ServedOrder: true,
      //Elevator: ???
    }
    elev_state.Orders[elev_state.Floor][elevio.BT_Cab] = false
    //NetworkUpdate <- update

    //For hallRequests
    update.MsgType = 0
    switch chooseDirection(elev_state) {

    case elevio.MD_Up:
      elev_state.Orders[elev_state.Floor][elevio.BT_HallUp] = false
      update.Button = int(elevio.BT_HallUp)
      //NetworkUpdate <- update

      if !requestsAbove(elev_state) {
        elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] = false
        update.Button = int(elevio.BT_HallDown)
        //NetworkUpdate <- update
      }

    case elevio.MD_Down:
      elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] = false
      update.Button = int(elevio.BT_HallDown)
      //NetworkUpdate <- update

      if !requestsBelow(elev_state) {
        elev_state.Orders[elev_state.Floor][elevio.BT_HallUp] = false
        update.Button = int(elevio.BT_HallUp)
        //NetworkUpdate <- update
      }

    case elevio.MD_Stop:
      fallthrough
    default:
      elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] = false
      update.Button = int(elevio.BT_HallDown)
      //NetworkUpdate <- update
      elev_state.Orders[elev_state.Floor][elevio.BT_HallUp] = false
      update.Button = int(elevio.BT_HallUp)
      //NetworkUpdate <- update
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
