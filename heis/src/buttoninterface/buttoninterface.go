package elevdriver

import ."./src/elev"

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

func buttonInterface(downChan chan, upChan chan) {
	
}