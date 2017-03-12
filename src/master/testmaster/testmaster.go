package main

import (
	"fmt"
	"master/testMNI"
	"time"
	. "typedef"
)

func main() {
	fmt.Println("Running tests!")

	reportNum := 0
	unitChan := make(chan UnitUpdate)

	statusReqChan := make(chan int, 1)
	statusChan := make(chan StatusType, 10)
	quit := make(chan bool)
	go testUnits(unitChan, quit)

	testMNI.Init_tmni(statusReqChan, statusChan, unitChan, quit)

	for {
		time.Sleep(1 * time.Second)
		reportNum++
		statusReqChan <- reportNum
		time.Sleep(300 * time.Millisecond)
		for len(statusChan) > 0 {
			fmt.Println("Got report:", <-statusChan)
		}
		if reportNum > 142 {
			close(quit)
			time.Sleep(1 * time.Second)
			return
		}
	}
}

func testUnits(unitChan <-chan UnitUpdate, quit <-chan bool)
{
	var units UnitUpdate
	for {
		select {
		case units = <- unitChan:
			fmt.Println("Got units:", units)
		case <- quit:
			fmt.Println("quitting units")
			return
		}
	}
}