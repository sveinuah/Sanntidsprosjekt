package testMNI

import (
	"fmt"
	"networkmodule/bcast"
	"networkmodule/peers"
	"strings"
	"time"
	. "typedef"
)

//abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan

//Variables
var name string
var messageTimedOut = 0

const (
	resendTime  = 10 * time.Millisecond
	timeOutTime = 95 * time.Millisecond

	rxPort       = 20014
	txPort       = 30014
	peersComPort = 40014
)

func Init_tmni(statusReqChan chan int, statusChan chan StatusType, unitUpdateChan chan UnitUpdate, orderTxChan chan OrderType, orderRxChan chan OrderType, lightsChan chan [][]bool, quitChan chan bool) {
	fmt.Println("testMNI initializing!")
	ackTxChan := make(chan AckType, 10)
	statusRxChan := make(chan StatusType, 10)
	peerUpdateChan := make(chan peers.PeerUpdate, 100)
	newOrderackRxChan := make(chan bool, 1)
	newOrderTxChan := make(chan OrderType, 100)
	ackRxChan := make(chan AckType, 100)
	extOrderRxChan := make(chan OrderType, 100)
	lightsTxChan := make(chan [][]bool, 10)

	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)
	go bcast.Transmitter(txPort, quitChan, ackTxChan, newOrderTxChan, lightsTxChan)
	go bcast.Receiver(rxPort, quitChan, statusRxChan, extOrderRxChan, ackRxChan)

	go requestAndReceiveStatus(statusChan, statusRxChan, statusReqChan, ackTxChan, quitChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
	go sendNewOrder(orderTxChan, newOrderTxChan, newOrderackRxChan, quitChan)
	go receiveAckHandler(ackRxChan, newOrderackRxChan, quitChan)
	go receiveOrder(extOrderRxChan, orderRxChan, ackTxChan, quitChan)
	go broadcastExtLights(lightsChan, lightsTxChan, quitChan)
}

func receiveAckHandler(ackRxChan chan AckType, newOrderackRxChan chan bool, quitChan chan bool) {
	var AckRec AckType
	for {
		select {
		case <-quitChan:
			return

		case AckRec = <-ackRxChan:
			if AckRec.Type == "Order Received" {
				newOrderackRxChan <- true
			}
		}
	}
}

func requestAndReceiveStatus(statusChan chan StatusType, statusRxChan chan StatusType, statusReqChan chan int, ackTxChan chan AckType, quitChan chan bool) {
	var reqStat AckType
	var ackStat AckType
	var recStatus StatusType

	reqStat.Type = "Status"
	ackStat.Type = "Status"

	for {
		select {
		case <-quitChan:
			return

		//Master requests statusupdates.
		case reqStat.ID = <-statusReqChan:
			ackTxChan <- reqStat

		//Handling the received statuses.
		case recStatus = <-statusRxChan:

			ackStat.To = recStatus.From
			ackStat.ID = recStatus.ID
			ackTxChan <- ackStat

			//Discards old messages if they arrive after the next status update
			if recStatus.ID != reqStat.ID {
				break
			}

			statusChan <- recStatus

		default:
		}
	}
}

func translatePeerUpdates(peerUpdateChan chan peers.PeerUpdate, unitUpdateChan chan UnitUpdate, quitChan chan bool) {
	var newPeerUpdate peers.PeerUpdate
	var newUnitUpdate UnitUpdate
	var newUnit UnitType
	for {
		select {
		case <-quitChan:
		case newPeerUpdate = <-peerUpdateChan:
			newUnitUpdate = UnitUpdate{}
			for _, val := range newPeerUpdate.Peers {
				strArr := strings.Split(val, ":")
				newUnit.Type = strArr[1]
				newUnit.ID = string(strArr[0])
				newUnitUpdate.Peers = append(newUnitUpdate.Peers, newUnit)
			}

			if newPeerUpdate.New != "" {
				newUnit.Type = strings.Split(newPeerUpdate.New, ":")[1]
				newUnit.ID = strings.Split(newPeerUpdate.New, ":")[0]
				newUnitUpdate.New = newUnit
			}

			for _, val := range newPeerUpdate.Lost {
				strArr := strings.Split(val, ":")
				newUnit.Type = strArr[1]
				newUnit.ID = string(strArr[0])
				newUnitUpdate.Lost = append(newUnitUpdate.Lost, newUnit)
			}

			unitUpdateChan <- newUnitUpdate
		default:
		}
	}
}

func sendNewOrder(newOrderChan chan OrderType, newOrderTxChan chan OrderType, newOrderackRxChan chan bool, quitChan chan bool) {
	var newOrder OrderType
	var sending bool
	for {
		select {
		case <-quitChan:
			return

		case newOrder = <-newOrderChan:

			sending = true

			// Move new order into transmit channel
			newOrderTxChan <- newOrder

			// While we wait for acknowledge from Master:
			timeoutTimer := time.NewTimer(timeOutTime)
			resendTimer := time.NewTimer(resendTime)

			for sending {

				select {

				case <-timeoutTimer.C:
					messageTimedOut++
					timeoutTimer.Stop()
					resendTimer.Stop()
					sending = false

				case <-newOrderackRxChan:
					timeoutTimer.Stop()
					resendTimer.Stop()
					sending = false

				case <-resendTimer.C:
					newOrderTxChan <- newOrder
					resendTimer.Reset(resendTime)

				default:
				}
			}
		default:
		}
	}
}

func receiveOrder(extOrderRxChan chan OrderType, orderRx chan OrderType, ackTxChan chan AckType, quitChan chan bool) {
	var extOrder OrderType
	var AckRec AckType
	for {
		select {
		case <-quitChan:
			return
		case extOrder = <-extOrderRxChan:
			//fmt.Println("Got Order:", extOrder)

			if extOrder.To != "" {
				break
			}

			if extOrder.New == true {

				AckRec.Type = "ButtonPress"
			} else if extOrder.New == false {
				AckRec.Type = "ExecOrder"
			}

			AckRec.To = extOrder.From

			//fmt.Println("Ack Order", AckRec)
			ackTxChan <- AckRec

			orderRx <- extOrder
		default:
		}
	}
}

func broadcastExtLights(lightsChan chan [][]bool, lightsTxChan chan [][]bool, quitChan chan bool) {

	var lights [][]bool
	t := time.Tick(10 * resendTime)

	for {
		select {
		case <-quitChan:
			return

		case <-t:
			if len(lights) > 0 {
				lightsTxChan <- lights
			}

		case lights = <-lightsChan:
		default:
		}
	}
}
