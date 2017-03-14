package masternetworkinterface

import (
	"fmt"
	"networkmodule/bcast"
	"networkmodule/peers"
	"strings"
	"time"
	. "typedef"
)

var id string

const (
	resendTime      = 20 * time.Millisecond
	lightResendTime = 200 * time.Millisecond
	timeOutTime     = 200 * time.Millisecond

	rxPort       = 20014
	txPort       = 30014
	peersComPort = 40014
)

func Peers(identity string, unitUpdateChan chan UnitUpdate, quitChan chan bool) {

	id = identity

	peerUpdateChan := make(chan peers.PeerUpdate, 100)

	go peers.Transmitter(peersComPort, id+":"+MASTER, quitChan)
	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
}

func Init_MNI(masterBackupChan chan [][]MasterOrder, quitChan chan bool) int {
	fmt.Println("Network interface initializing..")

	statusRx := make(chan StatusType, 10)

	go bcast.Receiver(rxPort, quitChan, statusRx)

	status := <-statusRx
	return len(status.MyOrders)
}

func Active(masterOrderTx chan OrderType, masterOrderRx chan OrderType, masterBackupChan chan [][]MasterOrder, masterStatusRx chan StatusType, masterLightsTx chan [][]bool, quitChan chan bool) {

	statusRx := make(chan StatusType, 100)
	lightsTx := make(chan [][]bool, 100)

	orderRx := make(chan OrderType, 100)
	orderTx := make(chan OrderType, 100)

	ackTx := make(chan AckType, 100)
	ackRx := make(chan AckType, 100)

	go bcast.Transmitter(txPort, quitChan, orderTx, masterBackupChan, lightsTx, ackTx)
	go bcast.Receiver(rxPort, quitChan, statusRx, orderRx, ackRx)

	go receiveStatus(masterStatusRx, statusRx, quitChan)
	go sendOrder(masterOrderTx, masterOrderRx, orderTx, ackRx, quitChan)
	go receiveOrder(orderRx, masterOrderRx, ackTx, quitChan)
	go broadcastExtLights(masterLightsTx, lightsTx, quitChan)

}

func Passive(masterBackupChan chan [][]MasterOrder, quitChan chan bool) {

	go bcast.Receiver(txPort, quitChan, masterBackupChan)

}

func receiveStatus(masterStatusRx chan StatusType, statusRx chan StatusType, quitChan <-chan bool) {
	var lastID int

	for {
		select {
		case <-quitChan:
			return

		case status := <-statusRx:
			if status.ID != lastID {
				masterStatusRx <- status
				lastID = status.ID
			}
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
	var order OrderType
	var sending bool

	for {
		select {
		case <-quitChan:
			return

		case order = <-masterOrderTx:

			sending = true

			for len(ackRx) > 0 {
				<-ackRx
			}
			orderTx <- order

			timeout := time.After(timeOutTime)
			resend := time.NewTicker(resendTime)

			for sending {

				select {
				case <-timeout:
					fmt.Println("Sending Order Timed Out..")
					sending = false
					order.To = ""
					masterOrderRx <- order
				case ack := <-ackRx:
					if ack.To == id && ack.From == order.To {
						sending = false
					}

					for len(ackRx) > 0 {
						ack = <-ackRx
						if ack.To == id && ack.From == order.To {
							sending = false
						}
					}

				case <-resend.C:
					orderTx <- order
				}
			}
			resend.Stop()
		}
	}
}

func receiveOrder(orderRx chan OrderType, masterOrderRx chan OrderType, ackTx chan AckType, quitChan chan bool) {
	var order OrderType
	var ack AckType
	for {
		select {
		case <-quitChan:
			return
		case order = <-orderRx:

			if order.To != "" {
				break
			}

			ack.To = order.From
			ack.From = id
			ackTx <- ack

			masterOrderRx <- order
		}
	}
}

func broadcastExtLights(masterLightsTx chan [][]bool, lightsTx chan [][]bool, quitChan chan bool) {
	var lights [][]bool
	t := time.NewTicker(lightResendTime)

	for {
		select {
		case <-quitChan:
			return
		case <-t.C:
			if len(lights) > 1 {
				lightsTx <- lights

			}
		case lights = <-masterLightsTx:
		}
	}
}
