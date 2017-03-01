package main

import (
	"../orderqueue"
	"../typedef"
	"log"
	"time"
)

/* Har vi lyst til å definere dette i networkInterface?

type elevatorReport struct {
	error int
	currentFloor int
	direction int
	running bool
	intOrderList [] Order
	newExtOrders [] Order
}*/

const (
	MASTER_SYNC_INTERVALL   int = 100 // in milliseconds
	INITIALIATION_WAIT_TIME int = 3   // in seconds
	REPORT_INTERVALL        int = 2   //in seconds

	STATE_GO_ACTIVE  int = 0
	STATE_ACTIVE     int = 1
	STATE_GO_PASSIVE int = 2
	STATE_PASSIVE    int = 3
	STATE_QUIT       int = 4
)

var id UnitID
var numFloors int
var unitList []UnitType
var orderList [][]masterOrder //numFloors+2directions

type masterOrder struct {
	OrderType
	Delegated time.Time
	Estimated time.Time
}

func main() {

	syncChan := make(chan []TimedOrder)
	interCom := make(chan []TimedOrder)
	quit := make(chan bool)

	syncTimer := time.Tick(MASTER_SYNC_INTERVALL * time.Millisecond)

	lastState := init(syncChan)
	state := lastState

	for {
		state = getState(lastState)
		switch state {
		case STATE_GO_PASSIVE:
			//transition from active to passive
			fallthrough
		case STATE_PASSIVE:
			select {
			case orderList = <-syncChan:
			default:
			}
		case STATE_GO_ACTIVE:
			//transition from passive to active
			go active(quit)
			fallthrough
		case STATE_ACTIVE:
			//handle sync
			select {
			case <-syncTimer:
				syncChan <- orderList
			}
		case STATE_QUIT:
			//do quit stuff
			close(quit)
		default:
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

func active(quit chan bool) { // not finished
	reportNum := 1
	orderNum := 1 //move to go-routine giving orders

	elevReports := make(map[UnitID]StatusType)

	statusReqChan := make(chan int) //to request reports with id
	statusChan := make(chan StatusType)
	unitChan := make(chan UnitType)

	go orders(elevReports, quit)

	reportTime := time.Tick(REPORT_INTERVALL * time.Second)

	for {
		select {
		case unit := <-unitChan:
			addUnit(unit)
		case report <- statusChan:
			elevReports[report.From] = report
		case <-reportTime:
			handleDeadUnits(elevReports, reportNum)
			reportNum++
			statusReqChan <- reportNum
		case err <- errChan:
			//handleErr
		case <-quit:
			//do quit stuff
		default:
		}
	}
}

func orders(reports *map[UnitID]StatusType, quit chan bool) {
	orderRx := make(chan OrderType)
	orderTx := make(chan OrderType)

	for {
		select {
		case order := <-orderRx:
			orderCapsule := masterOrder{order}
			handleOrder(orderCapsule)
		case <-quit:
			//quit
		default:
			//delegate orders
		}
	}
}

func init(syncChan chan [][]masterOrder) int {
	id, numFloors = networkinterface.masterInit()
	orderList = [numFloors][2]masterOrder{}

	timeOut := time.After(INITIALIATION_WAIT_TIME * time.Second)

	done := false
	for done != true {
		select {
		case orderList = <-sync:
		case <-timeOut:
			done = true
		}
	}

	if ckeckIfActive() {
		go active(orderChan, unitChan, elevStatusChan, masterSync, quit)
		return STATE_ACTIVE
	} else {
		return STATE_PASSIVE
	}
}

func getState(lastState int) int {
	// get and return STATE_QUIT when quit flag is raised NYI
	if checkIfActive() {
		if lastState == STATE_ACTIVE {
			return STATE_ACTIVE
		}
		return STATE_GO_ACTIVE
	} else {
		if lastState == STATE_PASSIVE {
			return STATE_PASSIVE
		}
		return STATE_GO_PASSIVE
	}
}

func checkIfActive() bool {
	for _, unit := range unitList {
		if unit.Type == TYPE_MASTER {
			if id > unit.ID {
				return false
			}
		}
	}
	return true
}

func addUnit(unit UnitType) {
	newUnit := true
	for _, u = range unitList {
		if u.ID == unit.ID {
			newUnit = false
			break
		}
	}
	if newUnit {
		unitList = append(unitList, unit)
	}
}

func handleDeadUnits(list map[UnitID]StatusType, num int) {
	deadUnits := make([]UnitID, 0, len(unitList))
	for id, report := range list {
		if report.ID != num {
			deadUnits = append(deadUnits, id)
		}
	}
	for i, unit := range unitList {
		for _, id := range deadUnits {
			if unit.ID == id {
				unitList = append(unitList[:i], unitList[i+1:]...)
			}
		}
	}
}

func handleOrder(o masterOrder) {
	if o.New {
		if orderList[o.Floor][o.Dir] == nil {
			orderList[o.Floor][o.Dir] = o
		}
		return
	}
	orderList[o.Floor][o.Dir] = masterOrder{}
}
