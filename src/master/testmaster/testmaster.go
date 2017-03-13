package main

import (
	"fmt"
	"master/testMNI"
	"runtime"
	"time"
	. "typedef"
)

func main() {
	fmt.Println("Running tests!")

	runtime.GOMAXPROCS(runtime.NumCPU())

	reportNum := 0

	unitChan := make(chan UnitUpdate, 5)

	orderTx := make(chan OrderType, 10)
	orderRx := make(chan OrderType, 10)
	lights := make(chan [][]bool, 10)

	statusReqChan := make(chan int, 1)
	statusChan := make(chan StatusType, 10)
	quit := make(chan bool)
	go testing(unitChan, orderTx, orderRx, lights, quit)

	testMNI.Init_tmni(statusReqChan, statusChan, unitChan, orderTx, orderRx, lights, quit)

	for {
		time.Sleep(1 * time.Second)
		reportNum++
		statusReqChan <- reportNum
		time.Sleep(300 * time.Millisecond)
		for len(statusChan) > 0 {
			/*status := */ <-statusChan
			//fmt.Println("Got report:", status)
		}
		if reportNum > 42 {
			close(quit)
			time.Sleep(1 * time.Second)
			return
		}
	}
}

func testing(unitChan <-chan UnitUpdate, orderTx chan<- OrderType, orderRx <-chan OrderType, lightChan chan<- [][]bool, quit <-chan bool) {
	var units UnitUpdate
	var order OrderType
	orderTime := time.Tick(3000 * time.Millisecond)
	lightMat := make([][]bool, 4, 4)
	for i := range lightMat {
		lightMat[i] = make([]bool, 2, 2)
	}
	//orders := make(chan OrderType, 15)

	for {
		select {
		case units = <-unitChan:
			fmt.Println("Got units:", units)
		case order = <-orderRx:
			if !order.New {
				lightMat[order.Floor][order.Dir] = false
				lightChan <- lightMat
			}
			//fmt.Println("Got order:", order)
		case <-orderTime:
			orderTx <- OrderType{To: "Jarvis", From: "Carl", Floor: 2, Dir: 0, New: true}
			lightMat[2][0] = true
			lightChan <- lightMat
		case <-quit:
			fmt.Println("quitting units")
			return
		}
	}
}
