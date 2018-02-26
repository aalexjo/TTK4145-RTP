package main

import (
	"./Driver/Elevio"
	//"./Status"
)

func main() {
	numFloors := 4

	elevio.Init("localhost:15657", numFloors)
}
