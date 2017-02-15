package main

import (
	"../typedef"
	"log"
	"time"
)

/* Har vi lyst til å definere dette i networkInterface?

const (
	MIN_MASTER_PORT = 20000
	MAX_MASTER_PORT = 30000

	MASTER_LISTEN_PORT = 40100
	ORDER_COMPLETE_PORT = 40200
	MASTER_SYNC_PORT = 40300
)

type elevatorReport struct {
	error int 
	currentFloor int 
	direction int
	running bool
	intOrderList [] Order
	newExtOrders [] Order
}*/

const (
	MASTER_SYNC_INTERVALL = (time.Second * 1)
	INITIALIATION_WAIT_TIME = (time.Second * 3)
)

var active bool
var unitID int
var unitList[] UnitType

type Queue interface {
	Enqueue()
	Dequeue()
}

type OrderQueue struct {
	OrderList []OrderType
}

type QueueError struct {
	err string
}

func (err *QueueError) Error() {
	return err.err
}

func (q *OrderQueue) Enqueue(o OrderType) OrderQueue {
	return append(q,o)
}

func (q *OrderQueue) Dequeue() (OrderQueue, OrderType, error) {
	l := len(q)
	if l == 0 {
		return q, nil, QueueError{"Queue is Empty"}
	}
	return q[1:], q[0], nil
}

func checkIfActive() {
	active = true
	for _, unit := range(unitList) {
		if unit.Type == TYPE_MASTER {
			if unitID > unit.Port {
				active = false
			}
		}
	}
}

func init(unitStatusChan chan UnitType, masterSync chan Queue) {
	// broadcast "I'm here" NYI
	//start network interface w/channels NYI

	timeOut := make(chan bool, 1)
	go func {
		time.Sleep(INITIALIATION_WAIT_TIME)
		timeOut <- true
	}

	done := false
	for done != true {
		select {
		case unit := <- unitStatusChan:
			unitHandler(unit)
		case orderList := <- masterSync:
			copy(masterOrderList, orderList)
		case done <- timeOut
		}
	}
	unitID = getUnitID()  //asks network interface for an I
	checkIfActive()
}

func unitHandler(unit UnitType) {
	newUnit := true
		for i = range(unitList){
			if i.Port == unit.Port {
				newUnit = false
				break
			}
		}
		if(newUnit) {
			unitList.append(unit)
		}
}

func masterSyncTimer(syncTimer chan bool)

func main() {
	orderChan := make(chan OrderPackage)
	unitStatusChan := make(chan UnitType)
	reportChan := make(chan elevatorReport)
	masterSync := make(chan Queue)
	syncTimer := make(chan bool,1)

	go func {
		for {
			time.Sleep(MASTER_SYNC_INTERVALL)
			syncTimer <- true
		}
	}

	orderQueue := new(orderQueue)

	init()

	for {
		switch active {
		case true:
			select {
			case unit := <- unitStatusChan:
				unitHandler(unit)
			case order := <- orderChan:
				orderQueue.Enqueue(order)
			case <- syncTimer:
				masterSync <- orderQueue
			default:
			//reports
			//handleorders
			}

		case false:
			select {
			case update := <- masterSync:
				masterOrderList = update
			default:
				
			}
		}
		checkIfActive()
		/*
			- Hvem er på nettverket?
			- Lag lister over heiser og mastere
			- Sjekk om jeg er sjef? Hvis ikke, hopp fram til **
			- Be om status fra alle heiser
			**
			


		*/
	}

}
