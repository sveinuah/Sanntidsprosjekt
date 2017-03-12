package masternetworkinterface

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

func Testfunction() {
	quitChan := make(chan bool)
	dataChan := make(chan [4][2]bool)
	go bcast.Transmitter(txPort, quitChan, dataChan)

	var order [4][2]bool
	order = [4][2]bool{{false}}

	for {
		dataChan <- order
		time.Sleep(4 * time.Second)
	}
}

func Init(ID string, masterBackupChan chan [][]MasterOrder, unitUpdateChan chan UnitUpdate, numFloorsChan chan int, quitChan chan bool) {

	name = ID

	statusChan := make(chan StatusType)
	statusRxChan := make(chan StatusType)

	ackTxChan := make(chan AckType)
	ackRxChan := make(chan AckType)
	statusReqChan := make(chan AckType)
	var statusAckTx AckType
	statusAckTx.Type = "Status"
	statusAckTx.ID = 1

	peerUpdateChan := make(chan peers.PeerUpdate)

	go peers.Transmitter(peersComPort, name+":"+MASTER, quitChan)
	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)
	go bcast.Transmitter(txPort, quitChan, statusReqChan, ackTxChan)
	go bcast.Receiver(rxPort, quitChan, statusRxChan, masterBackupChan, ackRxChan)

	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
	go requestAndReceiveStatus(statusChan, statusRxChan, statusReqChan, ackTxChan, quitChan)
	fmt.Println("All goroutines are go!")
	statusReqChan <- statusAckTx

	for {
		select {
		case <-quitChan:
			return
		case <-time.After(resendTime):
			fmt.Println("Sending Status Req")
			statusReqChan <- statusAckTx
		case status := <-statusChan:
			numFloorsChan <- len(status.MyOrders)
		default:
		}
	}
}

func Active(unitUpdateChan chan UnitUpdate, orderTx chan OrderType, orderRx chan OrderType, masterBackupChan chan [][]MasterOrder, statusChan chan StatusType, statusReqChan chan int, lightsChan chan [][]bool, quitChan chan bool) {

	statusRxChan := make(chan StatusType)

	ackTxChan := make(chan AckType)
	ackRxChan := make(chan AckType)

	peerUpdateChan := make(chan peers.PeerUpdate)

	newOrderTxChan := make(chan OrderType)
	newOrderackRxChan := make(chan bool)

	extOrderRxChan := make(chan OrderType)

	lightsTxChan := make(chan [][]bool)

	go peers.Transmitter(peersComPort, string(name)+":"+MASTER, quitChan)
	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)

	go bcast.Transmitter(txPort, quitChan, newOrderTxChan, masterBackupChan, lightsTxChan)
	go bcast.Receiver(rxPort, quitChan, statusRxChan, extOrderRxChan, ackRxChan)

	go requestAndReceiveStatus(statusChan, statusRxChan, statusReqChan, ackTxChan, quitChan)
	go sendNewOrder(orderTx, newOrderTxChan, newOrderackRxChan, quitChan)
	go receiveOrder(extOrderRxChan, orderRx, ackTxChan, quitChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
	go receiveAckHandler(ackRxChan, newOrderackRxChan, quitChan)

}

func Passive(masterBackupChan chan [][]MasterOrder, unitUpdateChan chan UnitUpdate, quitChan chan bool) {

	peerUpdateChan := make(chan peers.PeerUpdate)

	//Receive Backup
	go bcast.Receiver(rxPort, quitChan, masterBackupChan)

	// Call and poll network for units
	go peers.Transmitter(peersComPort, string(name)+":"+MASTER, quitChan)
	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)

	<-quitChan
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
			if extOrder.To != "" {
				break
			}

			if extOrder.New == true {

				AckRec.Type = "ButtonPress"
			} else if extOrder.New == false {
				AckRec.Type = "ExecOrder"
			}

			AckRec.To = extOrder.From
			ackTxChan <- AckRec

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
			resendTimer.Stop()
			return

		case <-resendTimer.C:
			lightsTxChan <- lights
			resendTimer.Reset(resendTime)

		case lights = <-lightsChan:
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
				strArr := strings.Split(val, ":")
				newUnit.Type = strArr[1]
				newUnit.ID = string(strArr[0])
				newUnitUpdate.Peers = append(newUnitUpdate.Peers, newUnit)
			}

			newUnit.Type = strings.Split(newPeerUpdate.New, ":")[1]
			newUnit.ID = strings.Split(newPeerUpdate.New, ":")[0]
			newUnitUpdate.New = newUnit

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
