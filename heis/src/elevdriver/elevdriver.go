package elevdriver

import ."elevio/elevio"

var state int 
const (
	DIR_STOP = 0
	DIR_UP = 1
	DIR_DOWN = -1
)

type ButtonFunction int

const (
	Up ButtonFunction = iota 
	Down
	Command
	Stop
	Obstruction
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

type Status struct {
	CurrentFloor int
	Dir int
	Running	bool
	Door bool
}



func driverInit() {
	elevInit()
	elevMotorDirection(0)
	state = IDLE
}

func setLights(buttons) {

}

func run(dir) {

}

func setIdle() {
	
}

func getStatus(statusChan) {
	

}

func stopRoutine() {

}


