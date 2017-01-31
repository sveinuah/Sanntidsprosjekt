package elevdriver

import ."elevio"

var state int 
const IDLE = 0
const GOING_UP = 1
const GOING_DOWN = -1

type status : struct {
currentFloor int
direction int
running bool
buttons bool [][]
door open bool
}


func driverInit() {
	elevInit()
	elevMotorDirection(0)
	state = idle
}

func setLight(Button) {
	elevButtonLight(Button., floor int, val bool)
}

func run(dir) {

}

func setIdle() {
	
}

func getStatus(statusChan) {

}

func stopRoutine() {

}


