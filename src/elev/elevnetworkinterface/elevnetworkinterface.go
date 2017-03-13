package elevnetworkinterface

import (
	"fmt"
	"networkmodule/bcast"
	"networkmodule/peers"
	"sync"
	"time"
	. "typedef"
)

//abortChan, allocateOrdersChan, executedOrdersChan, extLightsChan, extReportChan, elevStatusChan

//Variables

var id string

const RESEND_TIME = 15 * time.Millisecond
const TIMOUT_TIME = 2000 * time.Millisecond
const INDEPENDENT = false //If independent the elevator will handle its own external orders when disconnected.

const TxPort = 20014
const RxPort = 30014
const peersComPort = 40014

var CONNECTED = true
var cMutex = &sync.Mutex{}

func Start(ID string, quitChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, elevStatusChan chan StatusType) {
	fmt.Println("elevnet Init!")
	/*
		- absorb Status messages
		- pick up executed orders, if timeout, bounce back to BI
		- pick up button presses, if timeout, bounce back to setLights and allocateOrders
		- make extLights matrix and pass along
	*/

	id = ID

	statusTx := make(chan StatusType, 8)
	extLightsRx := make(chan [][]bool, 100)

	orderTx := make(chan OrderType, 100)
	orderRx := make(chan OrderType, 100)

	ackRx := make(chan AckType, 100)
	ackTx := make(chan AckType, 100)

	go peers.Transmitter(peersComPort, id+":"+SLAVE, quitChan)
	go bcast.Transmitter(TxPort, quitChan, statusTx, orderTx, ackTx)
	go bcast.Receiver(RxPort, quitChan, extLightsRx, orderRx, ackRx)

	go transmitStatus(statusTx, elevStatusChan, quitChan)
	go transmitOrder(buttonPressesChan, executedOrdersChan, orderTx, ackRx, quitChan)
	go receiveOrder(allocateOrdersChan, orderRx, ackTx, quitChan)
	go receiveExtLights(extLightsRx, extLightsChan, quitChan)

}

func transmitStatus(statusTx chan StatusType, elevStatusChan chan StatusType, quitChan chan bool) {
	var status StatusType
	var counter int = 0
	t := time.Tick(250 * time.Millisecond)
	for {
		select {
		case <-quitChan:
			return
		case <-t:
			status = <-elevStatusChan
			status.From = id
			status.ID = counter
			counter++

			statusTx <- status
		}
	}
}

func transmitOrder(buttonPressChan chan OrderType, executedOrdersChan chan OrderType, orderTx chan OrderType, ackRx chan AckType, quitChan chan bool) {
	var order OrderType
	var sending bool

	resend := time.NewTicker(RESEND_TIME)

	for {
		select {
		case <-quitChan:
			resend.Stop()
			return
		case order = <-buttonPressChan:
			fmt.Println("Button Order:", order)
		case order = <-executedOrdersChan:
			fmt.Println("Exec Order:", order)
		}

		sending = true

		order.From = id
		orderTx <- order

		timeout := time.After(TIMOUT_TIME)

		for sending {
			select {
			case <-timeout:
				sending = false
				fmt.Println("Timed out..")
			case ack := <-ackRx:
				if ack.To == id {
					sending = false
				}
			case <-resend.C:
				orderTx <- order
			}
		}
	}
}

func receiveOrder(allocateOrdersChan chan OrderType, orderRx chan OrderType, ackTx chan AckType, quitChan chan bool) {
	var ack AckType
	ack.From = id

	for {
		select {
		case <-quitChan:
			return
		case order := <-orderRx:
			if order.To == id {
				fmt.Println(id, " received order", order)
				ackTx <- ack
				allocateOrdersChan <- order
			}
		}
	}
}

func receiveExtLights(extLightsRx chan [][]bool, extLightsChan chan [][]bool, quitChan chan bool) {
	var extLights [][]bool
	for {
		select {
		case <-quitChan:
			return
		case extLights = <-extLightsRx:
			extLightsChan <- extLights
		}
	}
}
