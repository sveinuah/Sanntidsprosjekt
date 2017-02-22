package networkinterface

import (
	"../networkmodule/bcast"
	. "../typedef"
	"fmt"
	"time"
)

//abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan


var TIMEOUT = false

func ElevNetworkinterface(abortChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, extLightsChan chan [][]bool, setLightsChan chan OrderType, elevStatusChan chan StatusType, quitChan chan bool) {
	/*
		- absorb Status messages
		- pick up executed orders, if timeout, bounce back to BI
		- pick up button presses, if timeout, bounce back to setLights and allocateOrders
		- make extLights matrix and pass along
	*/

	ID := "Jarvis"
	TxPort := 30014
	RxPort := 30014

	statusTxChan := make(chan StatusType)
	statusReqRxChan := make(chan bool)
	statusAckRxChan := make(chan bool)

	buttonPressTxChan := make(chan OrderType)
	buttonAckRxChan := make(chan bool)

	extLightsRxChan := make(chan [][]bool)

	executedOrdersTxChan := make(chan OrderType)
	executedOrdersAckRxChan := make(chan bool)

	newOrdersRxChan := make(chan OrderType)
	newOrderAckChan := make(chan bool)

	go bcast.Transmitter(TxPort, statusTxChan, buttonPressTxChan, executedOrdersTxChan, ackTxChan)
	go bcast.Receiver(RxPort, statusReqRxChan, extLightsRxChan, newOrdersRxChan, ackRxChan)

	go sendStatus(statusTxChan, statusReqRxChan, elevStatusChan, statusAckRxChan, quitChan)

}

func sendStatus(statusTxChan chan StatusType, statusReqRxChan chan bool, elevStatusChan chan StatusType, statusAckRxChan chan bool, quitChan chan bool){
	
	var status StatusType
	var success bool
	var quit bool
	timeout := 0
	
	for !quit {

		select{
		
		case quit = <- quitChan:
		
		default:

			success = false

			// Wait for status request
			statusReq := <- statusReqRxChan

			// Get current status
			status = <- elevStatusChan

			// Move current status into transmit channel
			statusTxChan <- status 

			// While we wait for acknowledge from Master:
			for !success {
		
				select{
		
				case success = <- statusAckRxChan:
					timeout = 0
					break

				case timeout == 10:
					TIMEOUT = true
					break

				default:
					if timeout >= 1 {
						statusTxChan <- status	
					}
					
					go func(){
						for !success {
							time.Sleep(10*time.Millisecond)
							timeout++
						}
					}()	
				}
			}
		}
	}
}

/*
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
