package elevnetworkinterface

import (
	"fmt"
	"networkmodule/bcast"
	"networkmodule/peers"
	"time"
	. "typedef"
)

//abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan

//Variables

var Name UnitID

const messageTimedOut = 0

const RESEND_TIME = 500 * time.Millisecond
const TIMOUT_TIME = 5000 * time.Millisecond
const INDEPENDENT = true //If independent the elevator will handle its own external orders when disconnected.

const TxPort = 20014
const RxPort = 30014
const peersComPort = 40014

var CONNECTED = true

func Init(quitChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, elevStatusChan chan StatusType) {
	fmt.Println("elevnet Init!")
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

	buttonAckRxChan := make(chan bool)

	extLightsRxChan := make(chan [][]bool)

	ordersTxChan := make(chan OrderType)
	executedOrdersAckRxChan := make(chan bool)

	newOrdersRxChan := make(chan OrderType)
	ackRxChan := make(chan AckType)
	ackTxChan := make(chan AckType)
	transmitEnable := make(chan bool)

	go peers.Transmitter(peersComPort, string(Name)+":"+SLAVE, transmitEnable)

	go bcast.Transmitter(TxPort, statusTxChan, ordersTxChan, ackTxChan)
	go bcast.Receiver(RxPort, statusReqRxChan, extLightsRxChan, newOrdersRxChan, ackRxChan)

	go receiveAck(ackRxChan, statusReqRxChan, statusAckRxChan, buttonAckRxChan, executedOrdersAckRxChan, quitChan)
	go answerStatusCall(statusTxChan, statusReqRxChan, elevStatusChan, statusAckRxChan, quitChan)
	go transmitButtonPress(buttonPressesChan, ordersTxChan, buttonAckRxChan, allocateOrdersChan, setLightsChan, quitChan)
	go transmitExecOrders(executedOrdersChan, ordersTxChan, executedOrdersAckRxChan, quitChan)
	go receiveNewOrder(allocateOrdersChan, newOrdersRxChan, ackTxChan, quitChan)
	go receiveExtLights(extLightsRxChan, extLightsChan, quitChan)
}

func disConnect() {
	fmt.Println("Disconnected")
	CONNECTED = false
}

func reConnect() {
	fmt.Println("Connected")
	CONNECTED = true
}

func resetTimers(timeoutTimer *time.Timer, resendTimer *time.Timer) {
	if !timeoutTimer.Stop() {
		<-timeoutTimer.C
		fmt.Println("Reset Timeout")
	}
	timeoutTimer.Reset(TIMOUT_TIME)

	if !resendTimer.Stop() {
		<-resendTimer.C
		fmt.Println("Reset Resend")
	}
	resendTimer.Reset(RESEND_TIME)
}

func receiveAck(AckRxChan chan AckType, statusReqRxChan chan int, statusAckRxChan chan int, buttonAckRxChan chan bool, executedOrdersAckRxChan chan bool, quitChan chan bool) {
	var AckRec AckType
	for {
		select {
		case <-quitChan:
			return
		case AckRec = <-AckRxChan:
			if AckRec.Type == "Status" && AckRec.ID > 0 {
				statusReqRxChan <- AckRec.ID
				reConnect()
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

func answerStatusCall(statusTxChan chan StatusType, statusReqRxChan chan int, elevStatusChan chan StatusType, statusAckRxChan chan int, quitChan chan bool) {

	var status StatusType
	var statusReq int
	var sending bool

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
					disConnect()
					timeoutTimer.Stop()
					resendTimer.Stop()
					sending = false

				case <-statusAckRxChan:
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

			if !CONNECTED {
				if INDEPENDENT {
					allocateOrdersChan <- buttonPress
					setLightsChan <- buttonPress
				}
				sending = false
				break
			}

			// Move current button press into transmit channe
			buttonPressTxChan <- buttonPress

			// While we wait for acknowledge from Master:
			timeoutTimer := time.NewTimer(TIMOUT_TIME)
			resendTimer := time.NewTimer(RESEND_TIME)

			resetTimers(timeoutTimer, resendTimer)

			for sending {

				select {

				case <-timeoutTimer.C:
					if INDEPENDENT {
						allocateOrdersChan <- buttonPress
						setLightsChan <- buttonPress
					}

					disConnect()
					timeoutTimer.Stop()
					resendTimer.Stop()
					sending = false

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

			if !CONNECTED {
				sending = false
				break
			}

			// Move current button press into transmit channe
			executedOrdersTxChan <- executedOrder

			// While we wait for acknowledge from Master:
			timeoutTimer := time.NewTimer(TIMOUT_TIME)
			resendTimer := time.NewTimer(RESEND_TIME)

			resetTimers(timeoutTimer, resendTimer)

			for sending {

				select {

				case <-timeoutTimer.C:
					timeoutTimer.Stop()
					resendTimer.Stop()
					sending = false

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
	var extLights [][]bool
	for {
		select {
		case <-quitChan:
			return
		case extLights = <-extLightsRxChan:
			extLightsChan <- extLights
		default:
		}
	}
}
