package testMNI

import (
	"networkmodule/bcast"
	//"networkmodule/peers"
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

func Init_tmni(statusReqChan chan int, statusChan chan StatusType, quitChan chan bool) {

	ackTxChan := make(chan AckType)
	statusRxChan := make(chan StatusType)

	go bcast.Transmitter(txPort, quitChan, ackTxChan)
	go bcast.Receiver(rxPort, quitChan, statusRxChan)

	go requestAndReceiveStatus(statusChan, statusRxChan, statusReqChan, ackTxChan, quitChan)

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
