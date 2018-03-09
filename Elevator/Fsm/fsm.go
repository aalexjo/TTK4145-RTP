package fsm

import (
  "../Driver/Elevio"
  "time"
  "fmt"
  "../Status"
  "../Cost"
)

const FLOORS = 4
const ELEVATORS = 1
const BUTTONS = 3

//#TODO: Uncomment network message sending

func Fsm(NetworkUpdate chan<- status.UpdateMsg, FSMinfo <-chan cost.AssignedOrderInformation, init bool, elevID string){
  var updateMessage status.UpdateMsg
  var elev_state cost.AssignedOrderInformation


  in_buttons := make(chan elevio.ButtonEvent)
  in_floors  := make(chan int)

  door_timed_out := time.NewTimer(3 * time.Second)
  door_timed_out.Stop()

  go elevio.PollButtons(in_buttons)
  go elevio.PollFloorSensor(in_floors)
  fmt.Println("Fsm kjører nå.")

  if init {
    elevio.SetMotorDirection(elevio.MD_Down)
    elevio.SetFloorIndicator(0)
    L:
    for {
      select {
      case floor := <- in_floors:
        if floor == 0 {
          elevio.SetMotorDirection(elevio.MD_Stop)
          break L
        }
      }
    }
  }

  for {
      select {

      case elev_state = <- FSMinfo: //New states from cost function
        setAllLights(elev_state, elevID)
        switch elev_state.States[elevID].Behaviour {
        case "doorOpen":
          continue
        case "idle":
          newDirection := chooseDirection(elev_state, elevID, elev_state.States[elevID].Floor)
          switch newDirection {
          case elevio.MD_Stop:
            if elev_state.AssignedOrders[elevID][elev_state.States[elevID].Floor][0] || elev_state.AssignedOrders[elevID][elev_state.States[elevID].Floor][1] || elev_state.States[elevID].CabRequests[elev_state.States[elevID].Floor] {
            elevio.SetDoorOpenLamp(true)
            door_timed_out.Reset(3 * time.Second)

            //Behaviour message
            updateMessage.MsgType = 1
            updateMessage.Behaviour = "doorOpen"
            updateMessage.Elevator = elevID
            NetworkUpdate <- updateMessage

            clearAtCurrentFloor(elev_state, elevID, elev_state.States[elevID].Floor, NetworkUpdate)
            setAllLights(elev_state, elevID)
            }

          case elevio.MD_Up:
            elevio.SetMotorDirection(newDirection)
            //Direction Message
            updateMessage.MsgType = 3
            updateMessage.Direction = "up"
            updateMessage.Elevator = elevID
            NetworkUpdate <- updateMessage

            //Behaviour message
            updateMessage.MsgType = 1
            updateMessage.Behaviour = "moving"
            updateMessage.Elevator = elevID
            NetworkUpdate <- updateMessage

          case elevio.MD_Down:
            elevio.SetMotorDirection(newDirection)
            //Direction Message
            updateMessage.MsgType = 3
            updateMessage.Direction = "down"
            updateMessage.Elevator = elevID
            NetworkUpdate <- updateMessage

            //Behaviour message
            updateMessage.MsgType = 1
            updateMessage.Behaviour = "moving"
            updateMessage.Elevator = elevID
            NetworkUpdate <- updateMessage

          }
        case "moving":

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
          updateMessage.Button = int(buttonEvent.Button)
          updateMessage.ServedOrder = false //Nytt knappetrykk
          updateMessage.Elevator = elevID
        }
        NetworkUpdate <- updateMessage;

        /*-------------Handling Elevator stuff---------------*/
        //TODO:Eventuelt STOPP-knapp behandling eller noe?
        //TODO:Eventuelt hvis vi skal akseptere cabRequests umiddelbart??

      case floor := <- in_floors:
        /*--------------Message to send--------------------*/
          updateMessage.MsgType = 2 //Arrived at floor
          updateMessage.Floor = floor
          updateMessage.Elevator = elevID
          NetworkUpdate <- updateMessage;

          fmt.Println("Reached floor: ", floor)
          elevio.SetFloorIndicator(floor)

          if shouldStop(elev_state, elevID, floor) {
            elevio.SetMotorDirection(elevio.MD_Stop)
             clearAtCurrentFloor(elev_state, elevID, floor, NetworkUpdate)

             /*
            //Stop message
            updateMessage.MsgType = 3
            updateMessage.Direction = "stop"
            updateMessage.Elevator = elevID
            NetworkUpdate <- updateMessage
            */

            elevio.SetDoorOpenLamp(true)
            door_timed_out.Reset(3 * time.Second)

            //Behaviour message
            updateMessage.MsgType = 1
            updateMessage.Behaviour = "doorOpen"
            updateMessage.Elevator = elevID
            NetworkUpdate <- updateMessage

            //clearAtCurrentFloor(elev_state, elevID, floor, NetworkUpdate)
            setAllLights(elev_state, elevID)
          }

      case <- door_timed_out.C:
        elevio.SetDoorOpenLamp(false)
        newDirection := chooseDirection(elev_state, elevID, elev_state.States[elevID].Floor)
        switch newDirection {
        case elevio.MD_Stop:
          //Behaviour message
          updateMessage.MsgType = 1
          updateMessage.Behaviour = "idle"
          updateMessage.Elevator = elevID
          NetworkUpdate <- updateMessage

        case elevio.MD_Up:
          elevio.SetMotorDirection(newDirection)
          //Direction Message
          updateMessage.MsgType = 3
          updateMessage.Direction = "up"
          updateMessage.Elevator = elevID
          NetworkUpdate <- updateMessage

          //Behaviour message
          updateMessage.MsgType = 1
          updateMessage.Behaviour = "moving"
          updateMessage.Elevator = elevID
          NetworkUpdate <- updateMessage

        case elevio.MD_Down:
          elevio.SetMotorDirection(newDirection)
          //Direction Message
          updateMessage.MsgType = 3
          updateMessage.Direction = "down"
          updateMessage.Elevator = elevID
          NetworkUpdate <- updateMessage

          //Behaviour message
          updateMessage.MsgType = 1
          updateMessage.Behaviour = "moving"
          updateMessage.Elevator = elevID
          NetworkUpdate <- updateMessage

        }
      }
  }
}


func requestsAbove(elev_state cost.AssignedOrderInformation, elevID string, reachedFloor int) bool {
  for floor := reachedFloor+1; floor < FLOORS; floor++ {
    if elev_state.States[elevID].CabRequests[floor]{
      return true
    }
    for button := 0; button < 2; button++ {
      if elev_state.AssignedOrders[elevID][floor][button] {
        return true
      }
    }
  }
  return false
}


func requestsBelow(elev_state cost.AssignedOrderInformation, elevID string, reachedFloor int) bool {
  for floor := 0; floor < reachedFloor; floor++ {
    if elev_state.States[elevID].CabRequests[floor]{
      return true
    }
    for button := 0; button < 2; button++ {
      if elev_state.AssignedOrders[elevID][floor][button]{
        return true
      }
    }
  }
  return false
}


//Choose direction of travel
func chooseDirection(elev_state cost.AssignedOrderInformation, elevID string, floor int) elevio.MotorDirection {
  switch elev_state.States[elevID].Direction {
  case "stop":
    fallthrough
  case "down":
    if requestsBelow(elev_state, elevID, floor) {
      return elevio.MD_Down
    } else if requestsAbove(elev_state, elevID, floor) {
      return elevio.MD_Up
    } else {
      return elevio.MD_Stop
    }

  case "up":
    if requestsAbove(elev_state, elevID, floor) {
      return elevio.MD_Up
    } else if requestsBelow(elev_state, elevID, floor) {
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
func shouldStop(elev_state cost.AssignedOrderInformation, elevID string, floor int) bool {
  switch elev_state.States[elevID].Direction {
  case "down":
    return (elev_state.AssignedOrders[elevID][floor][elevio.BT_HallDown] ||
    elev_state.States[elevID].CabRequests[floor] ||
    !requestsBelow(elev_state, elevID, floor))

  case "up":
    return (elev_state.AssignedOrders[elevID][floor][elevio.BT_HallUp] ||
    elev_state.States[elevID].CabRequests[floor] ||
    !requestsAbove(elev_state, elevID, floor))
  case "stop":
    fallthrough
  default:
    return true
  }
}


//Clear order only if elevator is travelling in the right direction.
func clearAtCurrentFloor(elev_state cost.AssignedOrderInformation, elevID string, floor int, NetworkUpdate chan<- status.UpdateMsg){
    //For cabRequests
    update := status.UpdateMsg {
      MsgType: 4,
      Floor: floor,
      Button: 2,
      Behaviour: elev_state.States[elevID].Behaviour,
      Direction: elev_state.States[elevID].Direction,
      ServedOrder: true,
      Elevator: elevID,
    }
    
    NetworkUpdate <- update

    //For hallRequests
    update.MsgType = 0
    switch elev_state.States[elevID].Direction{//chooseDirection(elev_state, elevID, floor) {

    case "up":// elevio.MD_Up:
      update.Button = int(elevio.BT_HallUp)
      NetworkUpdate <- update

      if !requestsAbove(elev_state, elevID, floor) {
        update.Button = int(elevio.BT_HallDown)
        NetworkUpdate <- update
      }

    case "down"://elevio.MD_Down:
      update.Button = int(elevio.BT_HallDown)
      NetworkUpdate <- update

      if !requestsBelow(elev_state, elevID, floor) {
        update.Button = int(elevio.BT_HallUp)
        NetworkUpdate <- update
      }

    case "stop"://elevio.MD_Stop:
      fallthrough
    default:
      update.Button = int(elevio.BT_HallDown)
      NetworkUpdate <- update
      update.Button = int(elevio.BT_HallUp)
      NetworkUpdate <- update
    }
}


func setAllLights(elev_state cost.AssignedOrderInformation, elevID string) {
  for floor := 0; floor < FLOORS; floor++ {
    elevio.SetButtonLamp(elevio.BT_Cab, floor, elev_state.States[elevID].CabRequests[floor])
    for button := elevio.BT_HallUp; button < elevio.BT_Cab; button++ {
      elevio.SetButtonLamp(button, floor, elev_state.HallRequests[floor][button])
    }
  }
}
