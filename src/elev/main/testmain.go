package main

//For debugging
import (
	. "elev/elevio"
	"fmt"
	//. "typedef"
)

func main() {
	ElevInit()
	ElevMotorDirection(0)
	for {
		fmt.Println(ElevGetFloorSensorSignal())
	}
}

/*
Tested:
ElevButtonLight
motordir
floorindicator
floorsignal

Problem: Initiating [][]bool types: problem with matrix size.
*/
