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
var id string
var messageTimedOut = 0

const (
	resendTime      = 50 * time.Millisecond
	lightResendTime = 100 * time.Millisecond
	timeOutTime     = 95 * time.Millisecond

	rxPort       = 20014
	txPort       = 30014
	peersComPort = 40014
)

func Init_MNI(ID string, masterBackupChan chan [][]MasterOrder, unitUpdateChan chan UnitUpdate, quitChan chan bool) int {

	id = ID

	statusRx := make(chan StatusType, 10)

	peerUpdateChan := make(chan peers.PeerUpdate, 100)

	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)
	go bcast.Receiver(rxPort, quitChan, statusRx)

	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)

	fmt.Println("All goroutines are go!")

	status := <-statusRx
	return len(status.MyOrders)
}

func Active(unitUpdateChan chan UnitUpdate, masterOrderTx chan OrderType, masterOrderRx chan OrderType, masterBackupChan chan [][]MasterOrder, masterStatusRx chan StatusType, masterLightsTx chan [][]bool, quitChan chan bool) {

	peerUpdateChan := make(chan peers.PeerUpdate, 100)

	statusRx := make(chan StatusType, 100)
	lightsTx := make(chan [][]bool, 10)

	orderRx := make(chan OrderType, 100)
	orderTx := make(chan OrderType, 100)

	ackTx := make(chan AckType, 10)
	ackRx := make(chan AckType, 100)

	go peers.Transmitter(peersComPort, id+":"+MASTER, quitChan)
	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)

	go bcast.Transmitter(txPort, quitChan, orderTx, masterBackupChan, lightsTx, ackTx)
	go bcast.Receiver(rxPort, quitChan, statusRx, orderRx, ackRx)

	go receiveStatus(masterStatusRx, statusRx, quitChan)
	go sendOrder(masterOrderTx, masterOrderRx, orderTx, ackRx, quitChan)
	go receiveOrder(orderRx, masterOrderRx, ackTx, quitChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
	go broadcastExtLights(masterLightsTx, lightsTx, quitChan)

}

func Passive(masterBackupChan chan [][]MasterOrder, unitUpdateChan chan UnitUpdate, quitChan chan bool) {

	peerUpdateChan := make(chan peers.PeerUpdate, 100)

	//Receive Backup
	go bcast.Receiver(rxPort, quitChan, masterBackupChan)

	// Call and poll network for units
	go peers.Transmitter(peersComPort, id+":"+MASTER, quitChan)
	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
}

func receiveStatus(masterStatusRx chan StatusType, statusRx chan StatusType, quitChan <-chan bool) {
	fmt.Println("Starting receiveStatus!")
	var lastID int

	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting receiveStatus")
			return

		//Handling the received statuses.
		case status := <-statusRx:
			if status.ID > lastID || status.ID == 0 {
				masterStatusRx <- status
				lastID = status.ID
			}
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

func sendOrder(masterOrderTx chan OrderType, masterOrderRx chan OrderType, orderTx chan OrderType, ackRx chan AckType, quitChan chan bool) {
	fmt.Println("Starting sendNewOrder!")
	var order OrderType
	var sending bool

	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting sendNewOrder!")
			return

		case order = <-masterOrderTx:

			sending = true

			// Move new order into transmit channel
			orderTx <- order

			// While we wait for acknowledge from Elevator:
			timeout := time.After(timeOutTime)
			resend := time.NewTicker(resendTime)

			for sending {

				select {
				case <-timeout:
					sending = false
					order.To = ""
					masterOrderRx <- order
				case ack := <-ackRx:
					if ack.To == id && ack.From == order.To {
						sending = false
					}
				case <-resend.C:
					orderTx <- order
				}
			}
		}
	}
}

func receiveOrder(orderRx chan OrderType, masterOrderRx chan OrderType, ackTx chan AckType, quitChan chan bool) {
	fmt.Println("Starting receiveOrder!")
	var order OrderType
	var ack AckType
	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting receiveOrder!")
			return
		case order = <-orderRx:
			fmt.Println("Got order:", order)

			if order.To != "" {
				break
			}

			ack.To = order.From
			ack.From = id
			ackTx <- ack

			masterOrderRx <- order
		default:
		}
	}
}

func broadcastExtLights(masterLightsTx chan [][]bool, lightsTx chan [][]bool, quitChan chan bool) {
	fmt.Println("Starting broadcastExtLights!")
	var lights [][]bool
	t := time.Tick(lightResendTime)

	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting broadCastExtLigths!")
			return

		case <-t:
			for len(lights) > 0 {
				lightsTx <- lights
			}

		case lights = <-masterLightsTx:
			fmt.Println("Got LightUpdate", lights)
		}
	}
}
