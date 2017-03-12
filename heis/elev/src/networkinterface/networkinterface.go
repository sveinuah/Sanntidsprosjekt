package networkinterface

import (
	"../networkmodule/bcast"
	. "../typedef"
	"fmt"
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

	ID := "Jarvis"
	TxPort := 20014
	RxPort := 30014

	statusTxChan := make(chan StatusType)
	statusReqRxChan := make(chan bool)

	buttonPressTxChan := make(chan OrderType)
	extLightsTxChan := make(chan [][]bool)

	executedOrdersTxChan := make(chan OrderType)
	newOrdersRxChan := make(chan OrderType)

	ackTxChan := make(chan bool)
	ackRxChan := make(chan bool)

	go bcast.Transmitter(TxPort, statusTxChan, buttonPressTxChan, executedOrdersTxChan, ackTxChan)
	go bcast.Receiver(RxPort, statusReqRxChan, extLightsTxChan, newOrdersRxChan, ackRxChan)

	timeOutIter := 0
	timeOut := false

	abortFlag := false
	for abortFlag != true {
		select {
		case <-statusReqRxChan:

			status := <-elevStatusChan
			//Setting the ID before sending, may put this somewhere else
			status.ID = ID

			//Sendng the status
			statusTxChan <- status
			acked := false

			//Waiting for acknowledgement from master
			for acked == false && timeOut != true {
				select {
				//If we receive acknowledgement, set acked = true and return ack.
				case acked = <-ackRxChan:
					ackTxChan <- true

				// If we have tried ten times and not gotten an ack, set timeout and quit.
				default:
					if timeOutIter > 10 {
						timeOut = true
					}

					// Iterate and try to send the information 10 times.
					timeOutIter++
					time.Sleep(10 * time.Millisecond)
					statusTxChan <- status
				}
			}

		case status := <-elevStatusChan:

			statusTxChan <- status

		case buttonPress := <-buttonPressesChan:
			//Normally send to master.

			//If timeout: bounce back to elevator
			/*
				allocateOrdersChan <- buttonPress
				setLightsChan <- buttonPress
			*/
		case executedOrder <- executedOrdersChan:
			//Normally pass to master.
			//If timeout:

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
