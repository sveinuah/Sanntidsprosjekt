package main

import (."elevdriver"
"networkinterface"
)

const(
UP = 1
DOWN = 0
)
type externalReport struct{
error int
currentFloor int
direction int
running bool
intOrderList bool [ ]
newExtOrderList bool [ ]
}

var (
statusChan chan struct
masterRecvChan chan int
reportChan chan struct
backupFile string
localOrderList bool [ ]
)

func initialize() {

}

func backupToFile(backupFile) {
	//Copy externalReport to file
}

func backupFromFile(backupFile) {

}

func operateElevator() {
	
}

func main() {
	initialize()
	for {
		go operateElevator()
		

	}
}