package elevnetworkinterface

import (
	"../networkmodule/bcast"
	"../networkmodule/peers"
	"fmt"
	"time"
	. "typedef"
)

//abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan

//Variables

var Name UnitID
var messageTimedOut = 0
var RESEND_TIME = 10 * time.Millisecond
var TIMOUT_TIME = 95 * time.Millisecond

var TxPort = 20014
var RxPort = 30014
var peersComPort = 40014

var TIMEOUT = false

func Init(quitChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, elevStatusChan chan StatusType) {
	/*
		- absorb Status messages
		- pick up executed orders, if timeout, bounce back to BI
		- pick up button presses, if timeout, bounce back to setLights and allocateOrders
		- make extLights matrix and pass along
	*/

	Name = "Jarvis"

	statusTxChan := make(chan StatusType)
	statusReqRxChan := make(chan int)
	statusAckRxChan := make(chan int)

	buttonPressTxChan := make(chan OrderType)
	buttonAckRxChan := make(chan bool)

	extLightsRxChan := make(chan [][]bool)

	executedOrdersTxChan := make(chan OrderType)
	executedOrdersAckRxChan := make(chan bool)

	newOrdersRxChan := make(chan OrderType)
	ackRxChan := make(chan AckType)
	ackTxChan := make(chan AckType)

	go peers.Transmitter(peersComPort, Name+":"+SLAVE, transmitEnable)

	go bcast.Transmitter(TxPort, statusTxChan, buttonPressTxChan, executedOrdersTxChan, ackTxChan)
	go bcast.Receiver(RxPort, statusReqRxChan, extLightsRxChan, newOrdersRxChan, ackRxChan)

	go timeoutHandle(quitChan)
	go receiveAck(ackRxChan, statusReqRxChan, statusAckRxChan, buttonAckRxChan, executedOrdersAckRxChan, quitChan)
	go answerStatusCall(statusTxChan, statusReqRxChan, elevStatusChan, statusAckRxChan, quitChan)
	go transmitButtonPress(buttonPressesChan, buttonPressTxChan, buttonAckRxChan, allocateOrders, setLightsChan, quitChan)
	go transmitExecOrders(executedOrdersChan, executedOrdersTxChan, executedOrdersAckRxChan, quitChan)
	go receiveNewOrder(allocateOrdersChan, newOrdersRxChan, newOrderAckChan, quitChan)
	go receiveExtLights(extLightsRxChan, extLightsChan, quitChan)
}

func timeoutHandle(quitChan chan bool) {
	for {
		select {
		case <-quitChan:
		default:
			if messageTimedOut >= 1 {
				TIMEOUT = true
				messageTimedOut = 0
			}
		}
	}
}

func receiveAck(AckRxChan chan AckType, statusReqRxChan chan int, statusAckRxChan chan bool, buttonAckRxChan chan bool, executedOrdersAckRxChan chan bool, quitChan chan bool) {
	var AckRec AckType
	for {
		select {
		case <-quitChan:
			return
		case AckRec = <-AckRxChan:
			if AckRec.Type == "Status" && AckRec.ID > 0 {
				statusReqRxChan <- AckRec.ID
				TIMEOUT = false
			}
			if AckRec.To == Name {

				switch AckRec.Type {
				case "Status":
					statusAckRxChan <- AckRec.ID
				case "ButtonPress":
					buttonAckRxChan <- true
				case "ExecOrder":
					executedOrdersAckRxChan <- true
				default:
				}
			}
		}
	}
}

func answerStatusCall(statusTxChan chan StatusType, statusReqRxChan chan bool, elevStatusChan chan StatusType, statusAckRxChan chan bool, quitChan chan bool) {

	var status StatusType
	var statusReq int
	var sending bool
	var ackStat AckType

	for {

		select {

		case <-quitChan:
			return
		case statusReq = <-statusReqRxChan:

			sending = true

			// Get current status
			status = <-elevStatusChan

			//Add name to status
			status.From = Name

			//Add status itteration ID
			status.ID = statusReq

			// Move current status into transmit channel
			statusTxChan <- status

			// While we wait for acknowledge from Master:

			timeoutTimer := time.NewTimer(TIMOUT_TIME)
			resendTimer := time.NewTimer(RESEND_TIME)

			for sending {

				select {

				case <-timeoutTimer.C:
					messageTimedOut++
					fallthrough

				case ackStat = <-statusAckRxChan:
					timeoutTimer.Stop()
					resendTimer.Stop()
					sending = false

				case <-resendTimer.C:
					statusTxChan <- status
					resendTimer.Reset(RESEND_TIME)

				default:
				}
			}
		default:
		}
	}
}

func transmitButtonPress(buttonPressChan chan OrderType, buttonPressTxChan chan OrderType, buttonAckRxChan chan bool, allocateOrdersChan chan OrderType, setLightsChan chan OrderType, quitChan chan bool) {

	var buttonPress OrderType
	var sending bool

	for {

		select {

		case <-quitChan:
			return
		case buttonPress = <-buttonPressChan:

			sending = true

			if TIMEOUT {
				allocateOrdersChan <- buttonPress
				setLightsChan <- buttonPress
				sending = false
				break
			}

			// Move current button press into transmit channe
			buttonPressTxChan <- buttonPress

			// While we wait for acknowledge from Master:
			timeoutTimer := time.NewTimer(TIMOUT_TIME)
			resendTimer := time.NewTimer(RESEND_TIME)

			for sending {

				select {

				case <-timeoutTimer.C:
					allocateOrdersChan <- buttonPress
					setLightsChan <- buttonPress
					messageTimedOut++
					fallthrough

				case <-buttonAckRxChan:
					timeoutTimer.Stop()
					resendTimer.Stop()
					sending = false

				case <-resendTimer.C:
					buttonPressTxChan <- buttonPress
					resendTimer.Reset(RESEND_TIME)

				default:
				}
			}
		default:
		}
	}
}

func transmitExecOrders(executedOrdersChan chan OrderType, executedOrdersTxChan chan OrderType, executedOrdersAckRxChan chan bool, quitChan chan bool) {

	var executedOrder OrderType
	var sending bool

	for {

		select {

		case <-quitChan:
			return
		case executedOrder = <-executedOrdersChan:

			sending = true

			// Move current button press into transmit channe
			executedOrdersTxChan <- executedOrder

			// While we wait for acknowledge from Master:
			timeoutTimer := time.NewTimer(TIMOUT_TIME)
			resendTimer := time.NewTimer(RESEND_TIME)

			for sending {

				select {

				case <-timeoutTimer.C:
					messageTimedOut++
					fallthrough

				case <-executedOrdersAckRxChan:
					timeoutTimer.Stop()
					resendTimer.Stop()
					sending = false

				case <-resendTimer.C:
					executedOrdersTxChan <- executedOrder
					resendTimer.Reset(RESEND_TIME)

				default:
				}
			}
		}
	}
}

func receiveNewOrder(allocateOrdersChan chan OrderType, newOrdersRxChan chan OrderType, ackTxChan chan AckType, quitChan chan bool) {
	var newOrderAck AckType
	newOrderAck.From = Name
	newOrderAck.Type = "Order Received"

	for {
		select {
		case <-quitChan:
			return

		case newOrder := <-newOrdersRxChan:
			if newOrder.To != Name {
				break
			}

			ackTxChan <- newOrderAck
			allocateOrdersChan <- newOrder

		default:
		}
	}
}

func receiveExtLights(extLightsRxChan chan [][]bool, extLightsChan chan [][]bool, quitChan chan bool) {

	for {
		select {
		case <-quitChan:
			return
		default:
			extLightsChan <- extLightsRxChan
		}
	}
}