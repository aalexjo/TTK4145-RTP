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
  var elev_state cost.AssignedOrderInformation //Initialisere denne i en init-sak?


  in_buttons := make(chan elevio.ButtonEvent)
  in_floors  := make(chan int)

  door_timed_out := time.NewTimer(3 * time.Second)
  door_timed_out.Stop()

  go elevio.PollButtons(in_buttons)
  go elevio.PollFloorSensor(in_floors)
  fmt.Println("Fsm kjører nå.")

  if init {
    elevio.SetMotorDirection(elevio.MD_Down)
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
          fallthrough
        case "idle":
          newDirection := chooseDirection(elev_state, elevID)
          switch newDirection {
          case elevio.MD_Stop:
            elevio.SetDoorOpenLamp(true)
            door_timed_out.Reset(3 * time.Second)

            //Behaviour message
            updateMessage.MsgType = 1
            updateMessage.Behaviour = "doorOpen"
            updateMessage.Elevator = elevID
            //NetworkUpdate <- updateMessage

            clearAtCurrentFloor(elev_state, elevID, elev_state.States[elevID].Floor, NetworkUpdate)
            setAllLights(elev_state, elevID)

          case elevio.MD_Up:
            elevio.SetMotorDirection(newDirection)
            //Direction Message
            updateMessage.MsgType = 3
            updateMessage.Direction = "up"
            updateMessage.Elevator = elevID
            //NetworkUpdate <- updateMessage

            //Behaviour message
            updateMessage.MsgType = 1
            updateMessage.Behaviour = "moving"
            updateMessage.Elevator = elevID
            //NetworkUpdate <- updateMessage

          case elevio.MD_Down:
            elevio.SetMotorDirection(newDirection)
            //Direction Message
            updateMessage.MsgType = 3
            updateMessage.Direction = "down"
            updateMessage.Elevator = elevID
            //NetworkUpdate <- updateMessage

            //Behaviour message
            updateMessage.MsgType = 1
            updateMessage.Behaviour = "moving"
            updateMessage.Elevator = elevID
            //NetworkUpdate <- updateMessage

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
          updateMessage.ServedOrder = false //Nytt knappetrykk
          updateMessage.Elevator = elevID
        }
        //NetworkUpdate <- updateMessage;

        /*-------------Handling Elevator stuff---------------*/
        //TODO:Eventuelt STOPP-knapp behandling eller noe?
        //TODO:Eventuelt hvis vi skal akseptere cabRequests umiddelbart??

      case floor := <- in_floors:
        /*--------------Message to send--------------------*/
          updateMessage.MsgType = 2 //Arrived at floor
          updateMessage.Floor = floor
          updateMessage.Elevator = elevID
          //NetworkUpdate <- updateMessage;

          //fmt.Println("Reached floor: ", elev_state.Floor)

          if shouldStop(elev_state, elevID, floor) {
            elevio.SetMotorDirection(elevio.MD_Stop)

            //Stop message
            updateMessage.MsgType = 3
            updateMessage.Direction = "stop"
            updateMessage.Elevator = elevID
            //NetworkUpdate <- updateMessage

            elevio.SetDoorOpenLamp(true)
            door_timed_out.Reset(3 * time.Second)

            //Behaviour message
            updateMessage.MsgType = 1
            updateMessage.Behaviour = "doorOpen"
            updateMessage.Elevator = elevID
            //NetworkUpdate <- updateMessage

            clearAtCurrentFloor(elev_state, elevID, floor, NetworkUpdate)
            setAllLights(elev_state, elevID)
          }

      case <- door_timed_out.C:
        elevio.SetDoorOpenLamp(false);
        newDirection := chooseDirection(elev_state, elevID)
        switch newDirection {
        case elevio.MD_Stop:
          //Behaviour message
          updateMessage.MsgType = 1
          updateMessage.Behaviour = "idle"
          updateMessage.Elevator = elevID
          //NetworkUpdate <- updateMessage

        case elevio.MD_Up:
          elevio.SetMotorDirection(newDirection)
          //Direction Message
          updateMessage.MsgType = 3
          updateMessage.Direction = "up"
          updateMessage.Elevator = elevID
          //NetworkUpdate <- updateMessage

          //Behaviour message
          updateMessage.MsgType = 1
          updateMessage.Behaviour = "moving"
          updateMessage.Elevator = elevID
          //NetworkUpdate <- updateMessage

        case elevio.MD_Down:
          elevio.SetMotorDirection(newDirection)
          //Direction Message
          updateMessage.MsgType = 3
          updateMessage.Direction = "down"
          updateMessage.Elevator = elevID
          //NetworkUpdate <- updateMessage

          //Behaviour message
          updateMessage.MsgType = 1
          updateMessage.Behaviour = "moving"
          updateMessage.Elevator = elevID
          //NetworkUpdate <- updateMessage

        }
      }
  }
}


func requestsAbove(elev_state cost.AssignedOrderInformation, elevID string) bool {
  for floor := elev_state.States[elevID].Floor+1; floor < FLOORS; floor++ {
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


func requestsBelow(elev_state cost.AssignedOrderInformation, elevID string) bool {
  for floor := 0; floor < elev_state.States[elevID].Floor; floor++ {
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


//Choose direction of travel
func chooseDirection(elev_state cost.AssignedOrderInformation, elevID string) elevio.MotorDirection {
  switch elev_state.States[elevID].Direction {
  case "up":
    if requestsAbove(elev_state, elevID) {
      return elevio.MD_Up
    } else if requestsBelow(elev_state, elevID) {
      return elevio.MD_Down
    } else {
      return elevio.MD_Stop
    }

  case "stop":
    fallthrough
  case "down":
    if requestsBelow(elev_state, elevID) {
      return elevio.MD_Down
    } else if requestsAbove(elev_state, elevID) {
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
func shouldStop(elev_state cost.AssignedOrderInformation, elevID string, floor int) bool {
  switch elev_state.States[elevID].Direction {
  case "down":
    return (elev_state.AssignedOrders[elevID][elev_state.States[elevID].Floor][elevio.BT_HallDown] ||
    elev_state.States[elevID].CabRequests[floor] ||
    !requestsBelow(elev_state, elevID))

  case "up":
    return (elev_state.AssignedOrders[elevID][elev_state.States[elevID].Floor][elevio.BT_HallUp] ||
    elev_state.States[elevID].CabRequests[floor] ||
    !requestsAbove(elev_state, elevID))
  case "stop":
    fallthrough
  default:
    return true
  }
}


//Clear order only if elevator is travelling in the right direction. RETURNS updated state!
func clearAtCurrentFloor(elev_state cost.AssignedOrderInformation, elevID string, floor int, NetworkUpdate chan<- status.UpdateMsg){
    //For cabRequests
    update := status.UpdateMsg {
      MsgType: 4,
      Floor: floor,
      ServedOrder: true,
      Elevator: elevID,
    }
    //elev_state.Orders[elev_state.Floor][elevio.BT_Cab] = false
    //NetworkUpdate <- update

    //For hallRequests
    update.MsgType = 0
    switch chooseDirection(elev_state, elevID) {

    case elevio.MD_Up:
      //elev_state.Orders[elev_state.Floor][elevio.BT_HallUp] = false
      update.Button = int(elevio.BT_HallUp)
      //NetworkUpdate <- update

      if !requestsAbove(elev_state, elevID) {
        //elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] = false
        update.Button = int(elevio.BT_HallDown)
        //NetworkUpdate <- update
      }

    case elevio.MD_Down:
      //elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] = false
      update.Button = int(elevio.BT_HallDown)
      //NetworkUpdate <- update

      if !requestsBelow(elev_state, elevID) {
        //elev_state.Orders[elev_state.Floor][elevio.BT_HallUp] = false
        update.Button = int(elevio.BT_HallUp)
        //NetworkUpdate <- update
      }

    case elevio.MD_Stop:
      fallthrough
    default:
      //elev_state.Orders[elev_state.Floor][elevio.BT_HallDown] = false
      update.Button = int(elevio.BT_HallDown)
      //NetworkUpdate <- update
      //elev_state.Orders[elev_state.Floor][elevio.BT_HallUp] = false
      update.Button = int(elevio.BT_HallUp)
      //NetworkUpdate <- update
    }
    //return elev_state
}


func setAllLights(elev_state cost.AssignedOrderInformation, elevID string) {
  for floor := 0; floor < FLOORS; floor++ {
    elevio.SetButtonLamp(elevio.BT_Cab, floor, elev_state.States[elevID].CabRequests[floor])
    for button := elevio.BT_HallUp; button < elevio.BT_Cab; button++ {
      elevio.SetButtonLamp(button, floor, elev_state.HallRequests[floor][button])
    }
  }
}
