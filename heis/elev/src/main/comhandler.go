package main

import (
	"./elevio/elevio"
	. "./typedef"
	"log"
	"netinterface"
	"time"
)

func makeReport() {

}

func main() {
	allocateOrdersChan := make(chan OrderType, 100)
	executedOrdersChan := make(chan OrderType, 100)
	buttonPressesChan := make(chan OrderType, 100)
	setLightsChan := make(chan OrderType, 100)
	extLightsChan := make(chan extLightsMatrix, 1)
	elevStatusChan := make(chan StatusType, 1)
	initChan := make(chan int, 1)
	abortChan := make(chan bool, 1)

	abortFlag := false
	abortChan <- abort
	//Pass all as pointers?
	go drive(abortChan, allocateOrdersChan, executedOrdersChan, elevStatusChan, setLightsChan, elevStatusChan, initChan)
	go buttonInterface(abortChan, extLightsChan, buttonPressesChan, allocateOrdersChan, initChan)
	go networkinterface(abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan)

	for abort != true {
		abort = checkAbortFlag(abortChan)
		//delay?
	}
}
