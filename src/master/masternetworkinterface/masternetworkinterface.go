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

/*func Testfunction() {
	quitChan := make(chan bool)
	dataChan := make(chan [4][2]bool)
	go bcast.Transmitter(txPort, quitChan, dataChan)

	var order [4][2]bool
	order = [4][2]bool{{false}}

	for {
		dataChan <- order
		time.Sleep(4 * time.Second)
	}
}*/

func Init_MNI(ID string, masterBackupChan chan [][]MasterOrder, unitUpdateChan chan UnitUpdate, quitChan chan bool) int {

	name = ID

	statusChan := make(chan StatusType, 10)
	statusRxChan := make(chan StatusType, 10)

	ackTxChan := make(chan AckType, 10)
	ackRxChan := make(chan AckType, 100)
	statusReqChan := make(chan int, 10)
	var statusAckTx AckType
	statusAckTx.Type = "Status"
	statusAckTx.ID = 1

	peerUpdateChan := make(chan peers.PeerUpdate, 100)

	go peers.Transmitter(peersComPort, name+":"+MASTER, quitChan)
	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)
	go bcast.Transmitter(txPort, quitChan, ackTxChan)
	go bcast.Receiver(rxPort, quitChan, statusRxChan, masterBackupChan, ackRxChan)

	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
	go requestAndReceiveStatus(statusChan, statusRxChan, statusReqChan, ackTxChan, quitChan)

	fmt.Println("All goroutines are go!")
	statusReqChan <- 1

	for {
		select {
		case status := <-statusChan:
			for len(statusChan) > 0 {
				<-statusChan
			}
			return len(status.MyOrders)
		}
	}
}

func Active(unitUpdateChan chan UnitUpdate, orderTx chan OrderType, orderRx chan OrderType, masterBackupChan chan [][]MasterOrder, statusChan chan StatusType, statusReqChan chan int, lightsChan chan [][]bool, quitChan chan bool) {

	statusRxChan := make(chan StatusType, 100)

	ackTxChan := make(chan AckType, 10)
	ackRxChan := make(chan AckType, 100)

	peerUpdateChan := make(chan peers.PeerUpdate, 100)

	newOrderTxChan := make(chan OrderType, 100)
	newOrderackRxChan := make(chan bool, 10)

	extOrderRxChan := make(chan OrderType, 100)

	lightsTxChan := make(chan [][]bool, 10)

	go peers.Transmitter(peersComPort, name+":"+MASTER, quitChan)
	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)

	go bcast.Transmitter(txPort, quitChan, newOrderTxChan, masterBackupChan, lightsTxChan, ackTxChan)
	go bcast.Receiver(rxPort, quitChan, statusRxChan, extOrderRxChan, ackRxChan)

	go requestAndReceiveStatus(statusChan, statusRxChan, statusReqChan, ackTxChan, quitChan)
	go sendNewOrder(orderTx, newOrderTxChan, newOrderackRxChan, quitChan)
	go receiveOrder(extOrderRxChan, orderRx, ackTxChan, quitChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
	go receiveAckHandler(ackRxChan, newOrderackRxChan, quitChan)
	go broadcastExtLights(lightsChan, lightsTxChan, quitChan)

}

func Passive(masterBackupChan chan [][]MasterOrder, unitUpdateChan chan UnitUpdate, quitChan chan bool) {

	peerUpdateChan := make(chan peers.PeerUpdate, 100)

	//Receive Backup
	go bcast.Receiver(rxPort, quitChan, masterBackupChan)

	// Call and poll network for units
	go peers.Transmitter(peersComPort, name+":"+MASTER, quitChan)
	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
}

func receiveAckHandler(ackRxChan chan AckType, newOrderackRxChan chan bool, quitChan chan bool) {
	fmt.Println("Starting receiveAckHandler!")
	var AckRec AckType
	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting recAckHandler!")
			return

		case AckRec = <-ackRxChan:
			if AckRec.Type == "Order Received" {
				newOrderackRxChan <- true
			}
		}
	}
}

func requestAndReceiveStatus(statusChan chan StatusType, statusRxChan chan StatusType, statusReqChan chan int, ackTxChan chan AckType, quitChan <-chan bool) {
	fmt.Println("Starting Req and Rec Status!")
	var reqStat AckType
	var ackStat AckType
	var recStatus StatusType

	reqStat.Type = "Status"
	ackStat.Type = "Status"

	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting Req and Rec Status")
			return

		//Master requests statusupdates.
		case reqStat.ID = <-statusReqChan:
			//fmt.Println("Sending Status Request")
			ackTxChan <- reqStat

		//Handling the received statuses.
		case recStatus = <-statusRxChan:
			//fmt.Println("Got Status from:", recStatus.From)

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
	fmt.Println("Starting translatePeerUpdates!")
	var newPeerUpdate peers.PeerUpdate
	var newUnitUpdate UnitUpdate
	var newUnit UnitType
	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting translatePeerUpdates!")
			return
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
	fmt.Println("Starting sendNewOrder!")
	var newOrder OrderType
	var sending bool
	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting sendNewOrder!")
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

func receiveOrder(extOrderRxChan chan OrderType, orderRxChan chan OrderType, ackTxChan chan AckType, quitChan chan bool) {
	fmt.Println("Starting receiveOrder!")
	var extOrder OrderType
	var AckRec AckType
	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting receiveOrder!")
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

			orderRxChan <- extOrder
		default:
		}
	}
}

func broadcastExtLights(lightsChan chan [][]bool, lightsTxChan chan [][]bool, quitChan chan bool) {
	fmt.Println("Starting broadcastExtLights!")
	var lights [][]bool
	t := time.Tick(20 * resendTime)

	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting broadCastExtLigths!")
			return

		case <-t:
			for len(lights) > 0 {
				lightsTxChan <- lights
			}

		case lights = <-lightsChan:
			fmt.Println("Got LightUpdate", lights)
		default:
		}
	}
}
