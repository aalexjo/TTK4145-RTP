package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"io"

)
type Status_struct struct {
	HallRequests [][]bool `json:"hallRequests"`
	States       map[string]State_Values `json:"states"`
}

type State_Values struct {
	Behaviour   string `json:"behaviour"`
	Floor       uint `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequest"`
}

type Assigned_Requests struct {
	One [][]bool
	Two [][]bool

}


func main() {
	
	status := Status_Struct{
		HallRequests: [][]bool{{false,false},{false,false},{false,false},{false,false}},
		States: State{
					One: State_Values{
						Behaviour: "moving",
						Floor: 2,
						Direction: "up",
						CabRequests: []bool{false,false,false,true},
						},
					Two: State_Values{
						Behaviour: "moving",
						Floor: 2,
						Direction: "up",
						CabRequests: []bool{false,false,false,true},
						},
				 	},
			}
	b, err := json.Marshal(status)
	if err != nil {
		fmt.Println("error:", err)
	}
	//os.Stdout.Write(b)



	c:= exec.Command("./hall_request_assigner")//"./hall_request_assigner -i '" + string(b[:]) + "'").Output()				//"gnome-terminal","-x", "sh", "-c", 
	stdout, err := c.StdoutPipe()
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println("here1")

	stdin, err := c.StdinPipe()
	if err != nil {
		fmt.Println("error:", err)
	}

	if err := c.Start(); err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println("here2")
	io.WriteString(stdin, string(b[:]))
	fmt.Println("here3")
	/*
	requests := Assigned_Requests{
		One: [][]bool{{false,false},{false,false},{false,false},{false,false}},
		Two: [][]bool{{false,false},{false,false},{false,false},{false,false}},
	}*/
	var requests Assigned_Requests
	stdout.Read(b)
	fmt.Println(string(b))
    json.NewDecoder(stdout).Decode(&requests)
    fmt.Println("here4")
	b, err = json.Marshal(requests)
	if err != nil {
		fmt.Println("error:", err)
	}

	if err := c.Wait(); err != nil {
		fmt.Println("Error: ", err)
	}

	os.Stdout.Write(b)
	fmt.Println("")
}


//*/