package main

import (
	"elevdriver"
	. "typedef"
	//"log"
	//"networkinterface"
	"time"
)

func makeReport() {

}

func main() {
	allocateOrdersChan := make(chan OrderType, 100) //NetworkInterface/ButtonInterface -> Drive: Deliver orders for Drive to execute
	executedOrdersChan := make(chan OrderType, 100) //Drive -> NetworkInterface/ButtonInterface: Puts executed orders here, picked up by NetworkInterface and sent to Master
	buttonPressesChan := make(chan OrderType, 100)  //ButtonInterface -> NetworkInterface: New external button presses reported to master
	setLightsChan := make(chan OrderType, 100)      //Drive -> ButtonInterface: Clear lights after executed orders
	extLightsChan := make(chan [][]bool, 1)         //NetworkInterface -> ButtonInterface: Update external lights according to Master list
	elevStatusChan := make(chan StatusType, 1)      //Drive -> NetworkInterface: Report elevator status to Master
	initChan := make(chan int, 1)                   //Drive -> ButtonInterface: ButtonInterface waits for Drive to run ElevInit() and pass number of floors
	abortChan := make(chan bool, 1)                 //All -> All: If value is true, all channels abort

	abortFlag := false
	abortChan <- abortFlag
	//Pass all as pointers?
	go elevdriver.Drive(abortChan, allocateOrdersChan, executedOrdersChan, elevStatusChan, setLightsChan, initChan)
	go elevdriver.ButtonInterface(abortChan, extLightsChan, setLightsChan, buttonPressesChan, allocateOrdersChan, initChan)
	//go Networkinterface(abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, setLightsChan, elevStatusChan)

	for abortFlag != true {
		abortFlag = CheckAbortFlag(abortChan)
		time.Sleep(time.Second)
	}
}
