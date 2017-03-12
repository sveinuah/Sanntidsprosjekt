//Dummy networkmodule to test elevator module
package dummynetworkinterface

import (
	//"../../networkmodule/bcast"
	"fmt"
	"time"
	. "typedef"
)

func DummyNetworkinterface(abortChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, elevStatusChan chan StatusType) {
	/*
		- absorb Status messages
		- pick up executed orders, if timeout, do nothing
		- pick up button presses, if timeout, bounce back to setLights and allocateOrders
		- make extLights matrix and pass along
	*/
	//Init

	//statusTxChan := make(chan StatusType, 1000)
	//buttonTxChan := make(chan )

	//go bcast.Transmitter()

	abortFlag := false
	for abortFlag != true {
		select {
		case status := <-elevStatusChan:
			//statusTxChan <- status
			fmt.Println(status)
		case buttonPress := <-buttonPressesChan:
			//Normally send to master.

			//If timeout: bounce back to elevator
			allocateOrdersChan <- buttonPress
			setLightsChan <- buttonPress
		case executedOrder <- executedOrdersChan:
			//Normally pass to master.
			//If timeout: Do nothing
		default:
		}
		//Do stuff

		abortFlag = CheckAbortFlag(abortChan)
	}
}
