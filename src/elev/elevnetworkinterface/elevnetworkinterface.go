package elevnetworkinterface

import (
	"fmt"
	"networkmodule/bcast"
	"networkmodule/peers"
	"sync"
	"time"
	. "typedef"
)

var id string

const STATUS_RESEND_TIME = 500 * time.Millisecond
const RESEND_TIME = 20 * time.Millisecond
const TIMOUT_TIME = 200 * time.Millisecond

const TxPort = 20014
const RxPort = 30014
const peersComPort = 40014

var cMutex = &sync.Mutex{}

// Start ititializes the network module and distributes the channels acordingly.
func Start(ID string, quitChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, elevStatusChan chan StatusType) {
	fmt.Println("Starting Elevator network interface!")

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
	go transmitOrder(buttonPressesChan, executedOrdersChan, setLightsChan, orderTx, ackRx, quitChan)
	go receiveOrder(allocateOrdersChan, setLightsChan, orderRx, ackTx, quitChan)
	go receiveExtLights(extLightsRx, extLightsChan, quitChan)
}

// transmitStatus sends the status report from the elevator.
// It should be run as a go-routine.
// It terminates when something is received from the uit channel.
func transmitStatus(statusTx chan StatusType, elevStatusChan chan StatusType, quitChan chan bool) {
	var status StatusType
	var counter int = 0
	t := time.Tick(STATUS_RESEND_TIME)
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

// transmitOrder sends gatheres orders from two channels and transmits them on one.
// if no ack is received within timeOut, if the order was an executed order. The order light is cleared.
// It should be run as a go-routine.
// it terminates if something is received on the quit channel.
func transmitOrder(buttonPressChan chan OrderType, executedOrdersChan chan OrderType, setLightsChan chan OrderType, orderTx chan OrderType, ackRx chan AckType, quitChan chan bool) {
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
				if order.New == false {
					setLightsChan <- order
				}
			case ack := <-ackRx:
				if ack.To == id {
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
	}
}

// receiveOrder sends incomming orders to the elevator and sends an Acknowledge to the master that sent the order.
// It should be run to as a go-routine.
// It terminates if something is received on the
func receiveOrder(allocateOrdersChan chan OrderType, setLightsChan chan OrderType, orderRx chan OrderType, ackTx chan AckType, quitChan chan bool) {
	var ack AckType
	ack.From = id

	for {
		select {
		case <-quitChan:
			return
		case order := <-orderRx:
			if order.To == id {
				fmt.Println(id, "received order", order)
				ack.To = order.From
				ackTx <- ack
				allocateOrdersChan <- order

				if order.Dir == DIR_NODIR {
					setLightsChan <- order
				}
			}
		}
	}
}

// receiveExtLights should be run as a go-routine
// it terminates when something is received on the quit channel.
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
