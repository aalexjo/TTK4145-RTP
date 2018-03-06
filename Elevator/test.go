package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"os"
)

func main() {
	cmd := exec.Command("cmd","/C","echo", "{\"Name\": \"Bob\", \"Age\": 32}")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	var person struct {
		Name string
		Age  int
	}

	var b []byte
	stdout.Read(b)
	fmt.Println(b)
	os.Stdout.Write(b)
	//if err := json.NewDecoder(stdout).Decode(&person); err != nil {
	//	log.Fatal(err)
	//}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s is %d years old\n", person.Name, person.Age)
}

/*

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	//"io"
)

type Status_Struct struct {
	HallRequests [][]bool `json:"hallRequests"`
	States       State `json:"states"`
}

type State struct{
	One State_Values
	Two State_Values
}

type State_Values struct {
	Behaviour   string `json:"behaviour"`
	Floor       uint `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequest"`
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
	c := exec.Command("cmd", "/K", "start", "C:\\Users\\Alexander\\Documents\\Skole\\6._Semseter\\TTK4145-RTP\\Resources\\cost_fns\\hall_request_assigner\\hall_request_assigner.exe","--i",string(b))
	fmt.Println(string(b))
	
	stdout, err := c.StdoutPipe()
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println("here1")
	//stderr, _ := c.StderrPipe()
	if err := c.Start(); err != nil { 
        fmt.Println("Error: ", err)
    }  	

	var new_status Status_Struct
	fmt.Println(*(stdout))
    json.NewDecoder(stdout).Decode(&new_status)
    fmt.Println("here3")
	b, err = json.Marshal(new_status)
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)


}
*/