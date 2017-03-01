//Dummy networkmodule to test elevator module
package dummynetworkinterface

import (
	//"../../networkmodule/bcast"
	//"fmt"
	. "typedef"
)

func DummyNetworkinterface(quitChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, elevStatusChan chan StatusType) {
	/*
		- absorb Status messages
		- pick up executed orders, if timeout, bounce back to BI
		- pick up button presses, if timeout, bounce back to setLights and allocateOrders
		- make extLights matrix and pass along
	*/
	//Init

	//statusTxChan := make(chan StatusType, 1000)
	//buttonTxChan := make(chan )

	//go bcast.Transmitter()

	for {
		select {
		case <-elevStatusChan: //status := <-elevStatusChan:
			//statusTxChan <- status
			//fmt.Println(status)
		case buttonPress := <-buttonPressesChan:
			//Normally send to master.

			//If timeout: bounce back to elevator
			allocateOrdersChan <- buttonPress
			setLightsChan <- buttonPress

		case <-executedOrdersChan: //executedOrder := <-executedOrdersChan:
			//Normally send to master.
			//fmt.Println(executedOrder)
			//If timeout:
		case <-quitChan:
			return

		default:
		}
	}
}
