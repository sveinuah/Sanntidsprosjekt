package masterNetworkInterface

import (
	"networkmodule/bcast"
	"networkmodule/peers"
	"strings"
	"time"
	. "typedef"
)

//abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan

//Variables
var name UnitID
var resendTime = 10 * time.Millisecond
var timeOutTime = 95 * time.Millisecond
var messageTimedOut = 0

var rxPort = 20014
var txPort = 30014
var peersComPort = 40014

// Init initializes the go routines needed during the initialization of the master.
func Init(unitUpdateChan chan UnitUpdate, newOrderChan chan OrderType, receivedOrdersChan chan OrderType, masterBackupChan chan [][]MasterOrder, statusChan chan StatusType, lightsChan chan [][]bool, quitChan chan bool) {

	numFloors := nil //skaffe numFloors

	statusRxChan := make(chan StatusType)
	statusReqChan := make(chan int)

	AckTxChan := make(chan AckType)
	AckRxChan := make(chan AckType)

	transmitEnable := make(chan bool)
	peerUpdateChan := make(chan peers.PeerUpdate)

	newOrderTxChan := make(chan OrderType)
	newOrderAckRxChan := make(chan bool)

	lightsTxChan := make(chan [][]bool)

	go peers.Transmitter(peersComPort, Name+":"+MASTER, transmitEnable)
	go peers.Receiver(peersComPort, peerUpdateChan)
	go bcast.Transmitter(txPort, newOrderTxChan, lightsTxChan)
	go bcast.Receiver(rxPort, statusRxChan)

	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
	go receiveAckHandler(AckRxChan, newOrderAckRxChan, quitChan)
	return numFloors
}

// Active starts the go-routines needed for an active master
func Active(unitUpdateChan chan UnitUpdate, orderTx chan OrderType, orderRx chan OrderType, masterBackupChan chan [][]MasterOrder, statusChan chan StatusType, lightsChan chan [][]bool, quitChan chan bool) {
	// initiate active routine
}

// Passive starts the go-routines needed for a passive master.
func Passive(masterBackupChan chan [][]MasterOrder, quit chan bool) {
	// initiate passive routine
}

func receiveAckHandler(AckRxChan chan AckType, newOrderAckRxChan chan AckType, quitChan chan bool) {
	var AckRec AckType
	for {
		select {
		case <-quitChan:
			return

		case AckRec = <-AckRxChan:
			if AckRec.Type == "Order Received" {
				newOrderAckRxChan <- true
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
					fallthrough

				case <-newOrderAckRxChan:
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

func broadcastExtLights(lightsChan chan [][]bool, lightsTxChan chan [][]bool, quitChan chan bool) {
	var lights [][]bool
	resendTimer := time.NewTimer(2 * resendTime)
	for {
		select {
		case <-quitChan:
			return
		case <-resendTimer.C:
			fallthrough
		case lights = <-lightsChan:
			if !resendTimer.Stop() {
				<-resendTimer.C
			}
			if lights {
				lightsTxChan <- lights
			}
			resendTimer.Reset(2 * resendTime)
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
