package fsm

import (
  "./Driver/elevio"
  //"fmt"
  //"os/exec"
  //"bytes"
  //"log"
  "./Status/status"
)

// func OnInitBetweenFloors(){ //
//   elevio.SetMotorDirection(elevio.MD_Down)
//   //TODO: Update status w direction of travel
// }

//Denne funker ikke...
// func CalculateOptimalElevator(status string){
//   cmd := exec.Command("hall_request_assigner", "--input");
//   //cmd.Stdin = '{"hallRequests":[[false,false],[true,false],[false,false],[false,true]],"states":{"one":{"behaviour":"moving","floor":2,"direction":"up","cabRequests":[false,false,true,true]},"two":{"behaviour":"idle","floor":0,"direction":"stop","cabRequests":[false,false,false,false]}}}';
//   var out bytes.Buffer
// 	cmd.Stdout = &out
//   err := cmd.Run()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//   fmt.Printf("in all caps: %q\n", out.String())
// }

func main(){
  var updateMessage status.UpdateMsg
  numFloors := 4

  elevio.Init("localhost:15657", numFloors)

  in_buttons := make(chan elevio.ButtonEvent)
  in_floors  := make(chan int)
  out_msg    := make(chan status.UpdateMsg)

  go elevio.PollButtons(drv_buttons)
  go elevio.PollFloorSensor(drv_floors)


  for {
      select {
      case a := <- drv_buttons:
        if a.Button < 2 {
          updateMessage.msgType = 0 //Hall request
        }
        else {
          updateMessage.msgType = 4 //Cab request
        }
        updateMessage.update[0] = //FLoor og knapp
        out_msg <- updateMessage;

      case a := <- drv_floors:
          //Handle
      }
  }

}
