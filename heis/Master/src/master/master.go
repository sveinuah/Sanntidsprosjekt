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
var id UnitID
var unitList[] UnitType
var elevReports map[UnitID]StatusType

type Queue interface {
	Enqueue(interface{})
	Dequeue()
}

type timedOrder struct {
	Order OrderType
	timeStamp time.Time
}

type OrderQueue struct {
	OrderList []timedOrder
}

func (q *OrderQueue) Enqueue(o timedOrder) OrderQueue {
	return append(q,o)
}

func (q *OrderQueue) Dequeue() (OrderQueue, timedOrder, error) {
	l := len(q)
	if l == 0 {
		return q, nil, error{"Queue is empty"}
	}
	return q[1:], q[0], nil
}

func (q *OrderQueue) Find(o timedOrder) (timedOrder, error) {
	for _, order range(q) {
		if o.Floor == order.Floor && o.Dir == order.Dir {
			return order, nil
		}
	}
	return nil, error{"Could not find Order"}
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
	unitID = getUnitID()  //asks network interface for an ID

	checkIfActive()
}

func unitHandler(unit UnitType) {
	newUnit := true
		for _, u = range(unitList){
			if u.Port == unit.Port {
				newUnit = false
				break
			}
		}
		if(newUnit) {
			unitList.append(unit)
		}
}

func getElevStatus(elevStatusChan chan StatusType) {
	
	for i,unit = range(unitList){
		//getReport(unit)
		time.Sleep(time.MilliSecond*40) 
		select {
		case report := <- elevStatusChan:
			elevReports[report.ID] = report
		default:
			unitList = append(unitList[:i],unitList[i+1:]...) //deletes unit from list
		}
	}
} //Usikker på denne folkens :/

func handleOrders() {
	
}

func main() {
	orderChan := make(chan OrderPackage)
	unitChan := make(chan UnitType)
	elevStatusChan := make(chan StatusType, 1)
	masterSync := make(chan Queue)
	syncTimer := make(chan bool,1)

	go func() {
		for {
			time.Sleep(MASTER_SYNC_INTERVALL)
			syncTimer <- true
		}
	}

	masterQueue := new(OrderQueue)
	activeOrders := new(OrderQueue)


	init()

	for {
		switch active {
		case true:
			select {
			case unit := <- unitChan:
				unitHandler(unit)
			case order := <- orderChan:
				masterQueue.Enqueue(order)
			case <- syncTimer:
				masterSync <- orderQueue
			default:
				getElevStatus(elevStatusChan)
				handleOrders()
				checkIfActive()
			}

		case false:
			select {
			case update := <- masterSync:
				masterOrderList = update
			default:
				checkIfActive()				
			}
		}
		/*
			- Hvem er på nettverket?
			- Lag lister over heiser og mastere
			- Sjekk om jeg er sjef? Hvis ikke, hopp fram til **
			- Be om status fra alle heiser
			**
			


		*/
	}

}
