package masterNetworkInterface

import (
	"../networkmodule/bcast"
	"../networkmodule/peers"
	. "../typedef"
	"fmt"
	"strings"
	"time"
)

//abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan

//Variables
var Name UnitID
var RESEND_TIME = 10 * time.Millisecond
var TIMOUT_TIME = 95 * time.Millisecond
var messageTimedOut = 0

var RxPort = 20014
var TxPort = 30014
var peersComPort = 40014

func MasterInit(unitUpdateChan chan UnitUpdate, newOrderChan chan OrderType, receivedOrdersChan chan OrderType, masterBackupChan chan [][]masterOrder, statusChan chan StatusType, quitChan chan bool) {

	statusRxChan := make(chan StatusType)
	statusReqChan := make(chan int)

	AckTxChan := make(chan AckType)
	AckRxChan := make(chan AckType)

	transmitEnable := make(chan bool)
	peerUpdateChan := make(chan peers.PeerUpdate)

	newOrderTxChan := make(chan OrderType)
	newOrderAckRxChan := make(chan bool)

	go peers.Transmitter(peersComPort, Name+":"+MASTER, transmitEnable)
	go peers.Receiver(peersComPort, peerUpdateChan)
	go bcast.Transmitter(TxPort)
	go bcast.Receiver(RxPort, statusRxChan)

	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
	go receiveAckHandler(AckRxChan, newOrderAckRxChan, quitChan)

}

func receiveAckHandler(AckRxChan chan AckType, newOrderAckRxChan chan AckType, quitChan chan bool) {
	var AckRec AckType
	for {
		select {
		case <-quitChan:
			return
		case AckRec = <-AckRxChan:
			if AckRec.Type == "Order Received" {

			}
		}
	}
}

func requestAndReceiveStatus(statusChan chan StatusType, statusRxChan chan StatusType, statusReqChan chan bool, AckTxChan chan AckType, quitChan chan bool) {
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
		case reqStat.ID <- statusReqChan:
			AckTxChan <- reqStat

		//Handling the received statuses.
		case recStatus <- statusRxChan:

			ackStat.To = recStatus.From
			ackStat.ID = recStatus.ID
			AckTxChan <- ackStat

			//Discards old messages if they arrive after the next status update
			if recStatus.ID != reqStat.ID {
				break
			}

			statusChan <- recStatus

		default:
		}
	}
}

func sendNewOrder(newOrderChan chan OrderType, newOrderTxChan chan OrderType, newOrderAckRxChan chan bool, quitChan chan bool) {
	var newOrder OrderType
	var orderAck AckType
	for {
		select {
		case <-quitChan:
		case newOrder = <-newOrderChan:

			sending = true

			// Move new order into transmit channel
			newOrderTxChan <- newOrder

			// While we wait for acknowledge from Master:
			timeoutTimer := time.NewTimer(TIMOUT_TIME)
			resendTimer := time.NewTimer(RESEND_TIME)

			for sending {

				select {

				case <-timeoutTimer.C:
					messageTimedOut++
					fallthrough

				case <-newOrderAckRxChan:
					timeoutTimer.Stop()
					resendTimer.Stop()
					sending = false

				case <-resendTimer.C:
					newOrderTxChan <- newOrder
					resendTimer.Reset(RESEND_TIME)

				default:
				}
			}
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

			for _, val := range newPeerUpdate.Peers {
				val = strings.Split(val, ":")
				newUnit.Type = val[1]
				newUnit.ID = val[0]
				newUnitUpdate.Peers = append(newUnitUpdate.Peers, newUnit)
			}

			newUnit.Type = strings.Split(newPeerUpdate.New, ":")[1]
			newUnit.ID = strings.Split(newPeerUpdate.New, ":")[0]
			newUnitUpdate.New = newUnit

			for _, val := range newPeerUpdate.Lost {
				val = strings.Split(val, ":")
				newUnit.Type = val[1]
				newUnit.ID = val[0]
				newUnitUpdate.Lost = append(newUnitUpdate.Lost, newUnit)
			}

			unitUpdateChan <- newUnitUpdate
		default:
		}
	}
}
