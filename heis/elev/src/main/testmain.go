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
	testvar := 2
	for {
		ElevMotorDirection(testvar)
		time.Sleep(2 * time.Second)
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
