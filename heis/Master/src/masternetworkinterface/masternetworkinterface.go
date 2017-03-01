package networkinterface

import (
	"fmt"
	"networkmodule/bcast"
	"time"
	. "typedef"
)

//abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan

//Variables
var Name UnitID
var RESEND_TIME = 10 * time.Millisecond
var TIMOUT_TIME = 95 * time.Millisecond

var RxPort = 20014
var TxPort = 30014

func MasterInit(newOrderChan chan OrderType, receivedOrdersChan chan OrderType, masterBackupChan chan [][]masterOrder, statusChan chan StatusType, quitChan chan bool) {

	statusRxChan := make(chan StatusType)
	statusReqChan := make(chan int)
	AckTxChan := make(chan AckType)

	go bcast.Receiver(RxPort, statusRxChan)
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

func sendNewOrder(newOrderChan chan OrderType)
