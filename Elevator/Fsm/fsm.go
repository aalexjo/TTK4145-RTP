package fsm

import (
  "../Driver/Elevio"
  //"fmt"
  //"os/exec"
  //"bytes"
  //"log"
  "../Status"
)


//#TODO: allow communication between modules,




func Fsm(NetworkUpdate chan<- status.Status_struct, ElevStatus <-chan UpdateMsg){
  var updateMessage status.UpdateMsg
  numFloors := 4

  elevio.Init("localhost:15657", numFloors)

  in_buttons := make(chan elevio.ButtonEvent)
  in_floors  := make(chan int)
  //out_msg    := make(chan status.UpdateMsg)

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
        //out_msg <- updateMessage;

      case a := <- in_floors:
          updateMessage.MsgType = 2; //Arrived at floor
          updateMessage.Floor = a;

          //TODO: HVordan legge til hvilken elevator det er snakk om??
          //out_msg <- updateMessage;
      }
  }

}
