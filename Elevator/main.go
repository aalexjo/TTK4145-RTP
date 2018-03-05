package main

import (
<<<<<<< HEAD
	"./Driver/Elevio"
	//"./Status"
	"./Network"
=======
	//"./Driver/Elevio"
	"./Fsm"
>>>>>>> 3cfb9cd8b8d0cdca86521ef3e52f4d33a61b1e74
)

type UpdateMsg struct{

}

type somthing struct{
updateType

 floor int
 buttonType int
	arrivedAtfloor int

}
FLOORS := 4
ELEVATORS := 3

func main() {
<<<<<<< HEAD


	InternalUpdate := make(chan UpdateMsg) //sends updates that occured in this node to the network module
	ExternalUpdate := make(chan UpdateMsg) //sends updates that occured in the network to the status module
	elevio.Init("localhost:15657", FLOORS)

	go network.Network(InternalUpdate, ExternalUpdate)
	go status.Status(InternalUpdate, ExternalUpdate)


=======
	//numFloors := 4

	//elevio.Init("localhost:15657", numFloors)
	fsm.CalculateOptimalElevator("test")

>>>>>>> 3cfb9cd8b8d0cdca86521ef3e52f4d33a61b1e74
}
