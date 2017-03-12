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
	unitChan := make(chan UnitUpdate, 1)

	orderTx := make(chan OrderType, 10)
	orderRx := make(chan OrderType, 10)

	statusReqChan := make(chan int, 1)
	statusChan := make(chan StatusType, 10)
	quit := make(chan bool)
	go testint(unitChan, orderTx, orderRx, quit)

	testMNI.Init_tmni(statusReqChan, statusChan, unitChan, orderTx, orderRx, quit)

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

func testing(unitChan <-chan UnitUpdate, orderTx chan<- OrderType, orderRx <-chan OrderType, quit <-chan bool)
{
	var units UnitUpdate
	orderTime := time.Tick(3000 * time.Millisecond)
	orders := make(chan OrderType, 15)



	for {
		select {
		case units = <- unitChan:
			fmt.Println("Got units:", units)
		case order := <- orderRx:
			fmt.Println("Got order:", order)
		case <- orderTime:
			orderTx <- OrderType{To: "Jarvis", From: "Carl", Floor: 2, Dir: 0, New: true}
		case <- quit:
			fmt.Println("quitting units")
			return
		}
	}
}