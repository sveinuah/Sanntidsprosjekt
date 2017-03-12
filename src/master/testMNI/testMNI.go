package testMNI

import (
	"networkmodule/bcast"
	"networkmodule/peers"
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

func Init_tmni(statusReqChan chan int, statusChan chan StatusType, unitUpdateChan chan UnitUpdate, quitChan chan bool) {

	ackTxChan := make(chan AckType)
	statusRxChan := make(chan StatusType)
	peerUpdateChan := make(chan peers.PeerUpdate)

	go peers.Receiver(peersComPort, peerUpdateChan, quitChan)
	go bcast.Transmitter(txPort, quitChan, ackTxChan)
	go bcast.Receiver(rxPort, quitChan, statusRxChan)

	go requestAndReceiveStatus(statusChan, statusRxChan, statusReqChan, ackTxChan, quitChan)
	go translatePeerUpdates(peerUpdateChan, unitUpdateChan, quitChan)
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
