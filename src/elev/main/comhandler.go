package main

import (
	"elev/elevdriver"
	//"elev/dummyheis"
	. "typedef"
	//"log"
	"elev/elevnetworkinterface"
	"flag"
	"fmt"
	//"time"
)

var id string

func main() {
	allocateOrdersChan := make(chan OrderType, 1000) //NetworkInterface/ButtonInterface -> Drive: Deliver orders for Drive to execute
	executedOrdersChan := make(chan OrderType, 1000) //Drive -> NetworkInterface/ButtonInterface: Puts executed orders here, picked up by NetworkInterface and sent to Master
	buttonPressesChan := make(chan OrderType, 1000)  //ButtonInterface -> NetworkInterface: New external button presses reported to master
	setLightsChan := make(chan OrderType, 1000)      //Drive -> ButtonInterface: Clear lights after executed orders
	extLightsChan := make(chan [][]bool, 10)         //NetworkInterface -> ButtonInterface: Update external lights according to Master list
	elevStatusChan := make(chan StatusType, 10)      //Drive -> NetworkInterface: Report elevator status to Master
	initChan := make(chan bool, 1)                   //Drive -> ButtonInterface: ButtonInterface waits for Drive to run ElevInit() and pass number of floors
	quitChan := make(chan bool)                      //All -> All: If value is true, all channels abort

	flag.StringVar(&id, "id", "", "The slave's ID")
	flag.Parse()

	go elevdriver.Drive(quitChan, allocateOrdersChan, executedOrdersChan, elevStatusChan, setLightsChan, initChan)
	go elevdriver.ButtonInterface(quitChan, extLightsChan, setLightsChan, buttonPressesChan, allocateOrdersChan, initChan)
	//go dummyheis.DummyHeis(quitChan, allocateOrdersChan, executedOrdersChan, extLightsChan, setLightsChan, buttonPressesChan, elevStatusChan)
	elevnetworkinterface.Start(id, quitChan, allocateOrdersChan, executedOrdersChan, extLightsChan, setLightsChan, buttonPressesChan, elevStatusChan)

	fmt.Println("Evig comhandler")
	<-quitChan
}
