package main

//For debugging
import (
	. "elevio"
	"fmt"
	"time"
	//. "typedef"
)

func main() {
	ElevInit()
	testvar := -1
	for {
		time.Sleep(time.Millisecond * 5)
		testvar = ElevGetFloorSensorSignal()
		fmt.Println(testvar)
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
