package elevdriver

import ."elev"

var state int 
const IDLE = 0
const GOING_UP = 1
const GOING_DOWN = -1

type ButtonFunction int

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

type Status struct {
	currentFloor int
	direction int
	running bool
	buttons bool [][]
	door open bool
}


func driverInit() {
	elevInit()
	elevMotorDirection(0)
	dire = idle
}

func run() {
	for {
		if dir == 1 {
			elevMotorDirection(1)
		}
	}
}

func setIdle() {
	elevMotorDirection(0)
	state = idle
}

func getStatus(statusChan) {

}

func stopRoutine() {

}


