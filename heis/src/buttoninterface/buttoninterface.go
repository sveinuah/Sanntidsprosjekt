package elevdriver

import "./src/elevio"

/*
1. check if any buttons are pressed
2. put new orders in out-channel
3. check if any set/clear commands are in in-channel
4. set/clear accordingly
*/
const (
	Up ButtonFunction = iota 
	Down ButtonFunction
	Command ButtonFunction
	Stop ButtonFunction
	Obstruction ButtonFunction
)

type Button struct {
	Floor int
	Type ButtonFunction
	Pushed bool
}

type Light struct {
	Floor 	int
	Type 	ButtonFunction
	On 		bool
}

var externalOrderList [][] Button

func init() {
	nFloors := elevio.ElevInit()


}

func ButtonInterface(downChan chan, upChan chan) {
	init()
	
}
