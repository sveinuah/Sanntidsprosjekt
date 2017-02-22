package networkinterface

import (
	"../networkmodule/bcast"
	. "../typedef"
	"fmt"
	"time"
)

//abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan

var TIMEOUT = false
var RESEND_TIME = 10 * time.Millisecond
var TIMOUT_TIME = 100 * time.Millisecond

func ElevNetworkinterface(quitChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, elevStatusChan chan StatusType) {
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

	go transmitStatus(statusTxChan, statusReqRxChan, elevStatusChan, statusAckRxChan, quitChan)
	go transmitButtonPress(buttonPressesChan, buttonPressTxChan, buttonAckRxChan, allocateOrders, setLightsChan, quitChan)
	go transmitExecOrders(executedOrdersChan, executedOrdersTxChan, executedOrdersAckRxChan, quitChan)
}

func transmitStatus(statusTxChan chan StatusType, statusReqRxChan chan bool, elevStatusChan chan StatusType, statusAckRxChan chan bool, quitChan chan bool) {

	var status StatusType
	var statusReq bool
	var success bool
	var quit bool

	for !quit {

		select {

		case quit = <-quitChan:

		default:

			success = false

			// Wait for status request
			statusReq = <-statusReqRxChan

			// Get current status
			status = <-elevStatusChan

			// Move current status into transmit channel
			statusTxChan <- status

			// While we wait for acknowledge from Master:

			timeoutTimer := time.NewTimer(TIMOUT_TIME)
			resendTimer := time.NewTimer(RESEND_TIME)

			for !success {

				select {

				case success = <-statusAckRxChan:
					timeoutTimer.Stop()
					resendTimer.Stop()
					break

				case <-timeoutTimer.C:
					success = true
					TIMEOUT = true
					break

				case <-resendTimer.C:
					statusTxChan <- status
					resendTimer.Reset(RESEND_TIME)

				default:
				}
			}
		}
	}
}

func transmitButtonPress(buttonPressChan chan OrderType, buttonPressTxChan chan OrderType, buttonAckRxChan chan bool, allocateOrdersChan chan OrderType, setLightsChan chan OrderType, quitChan chan bool) {

	var buttonPress OrderType
	var success bool
	var quit bool

	for !quit {

		select {

		case quit = <-quitChan:

		default:

			success = false

			// Wait for button press
			buttonPress = <-buttonPressChan

			// Move current button press into transmit channe
			buttonPressTxChan <- buttonPress

			// While we wait for acknowledge from Master:
			timeoutTimer := time.NewTimer(TIMOUT_TIME)
			resendTimer := time.NewTimer(RESEND_TIME)

			for !success {

				select {

				case success = <-buttonAckRxChan:
					timeoutTimer.Stop()
					resendTimer.Stop()
					break

				case <-timeoutTimer.C:
					success = true
					TIMEOUT = true

					allocateOrdersChan <- buttonPress
					setLightsChan <- buttonPress

					break

				case <-resendTimer.C:
					buttonPressTxChan <- buttonPress
					resendTimer.Reset(RESEND_TIME)

				default:
				}
			}
		}
	}
}

func transmitExecOrders(executedOrdersChan chan OrderType, executedOrdersTxChan chan OrderType, executedOrdersAckRxChan chan bool, quitChan chan bool) {

	var executedOrder OrderType
	var success bool
	var quit bool

	for !quit {

		select {

		case quit = <-quitChan:

		default:

			success = false

			// Wait for button press
			executedOrder = <-executedOrdersChan

			// Move current button press into transmit channe
			executedOrdersTxChan <- executedOrder

			// While we wait for acknowledge from Master:
			timeoutTimer := time.NewTimer(TIMOUT_TIME)
			resendTimer := time.NewTimer(RESEND_TIME)

			for !success {

				select {

				case success = <-executedOrdersAckRxChan:
					timeoutTimer.Stop()
					resendTimer.Stop()
					break

				case <-timeoutTimer.C:
					success = true
					TIMEOUT = true
					break

				case <-resendTimer.C:
					executedOrdersTxChan <- executedOrder
					resendTimer.Reset(RESEND_TIME)

				default:
				}
			}
		}
	}
}
