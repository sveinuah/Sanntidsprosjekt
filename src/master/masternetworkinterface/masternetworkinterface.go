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

func Init(ID UnitID, masterBackupChan chan [][]masterOrder, unitUpdateChan chan UnitUpdate, quitChan chan bool) int {

	Name = ID

	statusChan := make(chan StatusType)
	statusRxChan := make(chan StatusType)
	statusReqChan := make(chan int)
	peerUpdateChan := make(chan peers.PeerUpdate)

	go peers.Transmitter(peersComPort, quitChan, string(Name)+":"+MASTER)
	go peers.Receiver(peersComPort, quitChan, peerUpdateChan)
	go bcast.Transmitter(TxPort, quitChan, statusReqChan, AckTxChan)
	go bcast.Receiver(RxPort, quitChan, statusRxChan, masterBackupChan, AckRxChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
	go requestAndReceiveStatus(statusChan, statusRxChan, statusReqChan, AckTxChan, quitChan)

	statusReqChan <- 1

	for {
		select {
		case <-quitChan:
			return nil
		case <-time.After(RESEND_TIME):
			statusReqChan <- 1
		case status := <-statusChan:
			return len(status.MyOrder)
		default:
		}
	}
}

func Active(unitUpdateChan chan UnitUpdate, orderTx chan OrderType, orderRx chan OrderType, masterBackupChan chan [][]masterorder, statusChan chan StatusType, statusReqChan chan bool, lightsChan chan [][]bool, quitChan chan bool) {

	statusRxChan := make(chan StatusType)
	statusReqChan

	AckTxChan := make(chan AckType)
	AckRxChan := make(chan AckType)

	transmitEnable := make(chan bool)
	peerUpdateChan := make(chan peers.PeerUpdate)

	newOrderTxChan := make(chan OrderType)
	newOrderAckRxChan := make(chan bool)

	extOrderRxChan := make(chan OrderType)

	lightsTxChan := make(chan [][]bool)

	go peers.Transmitter(peersComPort, string(Name)+":"+MASTER, quitChan)
	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)

	go bcast.Transmitter(TxPort, quitChan, newOrderTxChan, masterBackupChan, lightsChan)
	go bcast.Receiver(RxPort, quitChan, statusRxChan, extOrderRxChan, AckRxChan)

	go requestAndReceiveStatus(statusChan, statusRxChan, statusReqChan, AckTxChan, quitChan)
	go sendNewOrder(orderTx, newOrderTxChan, newOrderAckRxChan, quitChan)
	go receiveOrder(extOrderRxChan, orderRx, AckTxChan, quitChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
	go receiveAckHandler(AckRxChan, newOrderAckRxChan, quitChan)
}

func Passive(masterBackupChan chan [][]masterOrder, unitUpdateChan chan UnitUpdate, quitChan chan bool) {

	transmitEnableChan := make(chan bool)
	peerUpdateChan := make(chan peers.PeerUpdate)

	//Receive Backup
	go bcast.Receiver(RxPort, quitChan, quitChan, masterBackupChan)

	// Call and poll network for units
	go peers.Transmitter(peersComPort, quitChan, string(Name)+":"+MASTER)
	go peers.Receiver(peersComPort, quitChan, peerUpdateChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
}

func receiveAckHandler(AckRxChan chan AckType, newOrderAckRxChan chan bool, quitChan chan bool) {
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

func receiveOrder(extOrderRxChan chan OrderType, orderRx chan OrderType, AckTxChan chan AckType, quitChan chan bool) {
	var extOrder OrderType
	var AckRec AckType
	for {
		select {
		case <-quitChan:
			return
		case extOrder = <-extOrderRxChan:
			if extOrder.To != "" {
				break
			}

			if extOrder.New == 1 {

				AckRec.Type = "ButtonPress"
			} else if extOrder.New == 0 {
				AckRec.Type = "ExecOrder"
			}

			AckRec.To = extOrder.From
			AckTxChan <- AckRec

			orderRx <- extOrder
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
