package networkinterface

import (
	"../networkmodule/bcast/"
	"../typedef"
	"fmt"
	"strconv"
	"time"
)

//abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan

func ElevNetworkinterface(abortChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, extLightsChan chan [][]bool, setLightsChan chan OrderType, elevStatusChan chan StatusType) {
	/*
		- absorb Status messages
		- pick up executed orders, if timeout, bounce back to BI
		- pick up button presses, if timeout, bounce back to setLights and allocateOrders
		- make extLights matrix and pass along
	*/
	//Init

	statusTxChan := make(chan StatusType,1)
	buttonTxChan := make(chan )

	go bcast.Transmitter()

	abortFlag := false
	for abortFlag != true {
		select {
		case status := <-elevStatusChan:

			statusTxChan <- status

		case buttonPress := <-buttonPressesChan:
			//Normally send to master.

			//If timeout: bounce back to elevator
			allocateOrdersChan <- buttonPress
			setLightsChan <- buttonPress

		case executedOrder <- executedOrdersChan:
			//Normally send to master.
			//If timeout: 

		default:
		}
		//Do stuff

		abortFlag = CheckAbortFlag()
	}
}