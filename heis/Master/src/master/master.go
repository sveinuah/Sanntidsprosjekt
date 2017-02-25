package main

import (
	"../typedef"
	"log"
	"time"
	"../orderqueue"
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
	MASTER_SYNC_INTERVALL int = 1000 // in milliseconds
	INITIALIATION_WAIT_TIME int = 3 // in seconds
	REPORT_INTERVALL int = 2 //in seconds

	STATE_GO_ACTIVE int = 0
	STATE_ACTIVE int = 1
	STATE_GO_PASSIVE int = 2
	STATE_PASSIVE int = 3
	STATE_QUIT int = 4
)

var id UnitID
var unitList[] UnitType

func main() {
	var orderBackup map[UnitID]TimedOrder
	var elevReports map[UnitID]StatusType

	orderChan := make(chan OrderPackage)
	unitChan := make(chan UnitType)
	sync := make(chan []TimedOrder)
	interCom := make(chan []TimedOrder)
	quit := make(chan bool)
	
	syncTimer := time.Tick(MASTER_SYNC_INTERVALL * time.Millisecond)

	lastState := init()
	state := lastState

	for {
		state = getState(lastState)
		switch state {
		case STATE_GO_PASSIVE:
			//transition from active to passive
			fallthrough
		case STATE_PASSIVE:
			select {
			case orderBackup = <- masterSync:
			default:			
			}
		case STATE_GO_ACTIVE:
			//transition from passive to active
			go active(orderChan, unitChan, sync, interCom, quit)
			fallthrough
		case STATE_ACTIVE:
			//handle sync
			masterSync <- orderBackup
		case STATE_QUIT:
			//do quit stuff
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

func active(orders chan OrderType, units chan UnitType, interCom chan map[UnitID]TimedOrder, quit chan bool) { // not finished
	elevReports := make(map[UnitID]StatusType)
	orderList := <- interCom // provided from backup

	reportTime := time.Tick(REPORT_INTERVALL * time.Second)

	for {
		select {
		case unit := <- unitChan:
			unitHandler(unit)
		case order := <- orderChan:
			timedOrder := TimedOrder{order}
			if order.New {
				orderList[timedOrder.ID] = timedOrder
			} else {
				delete(orderList,timedOrder.ID)
			}
			interCom <- orderList
		case <- reportTime:
			elevReports = getElevStatus()
		case <- quit:
			//do quit stuff
		default:
			for key, order := range(orderList) {
				// BLI ENIGE OM MASTER SKAL HOLDE TID. TRUR EGENTLIG DET ER NETWORK INTERFACE MAT.
		}
	}
}


func init(unitStatusChan chan UnitType, sync chan Queue) {
	// broadcast "I'm here" NYI
	//start network interface w/channels NYI

	timeOut := time.After(INITIALIATION_WAIT_TIME * time.Second)

	done := false
	for done != true {
		select {
		case unit := <- unitStatusChan:
			unitHandler(unit)
		case orderList := <- sync:
			copy(masterOrderList, orderList)
		case <- timeOut:
			done = true
		}
	}
	unitID = getUnitID()  //asks network interface for an ID 
	
	if ckeckIfActive() {
		go active(orderChan, unitChan, elevStatusChan, masterSync, quit)
		return STATE_ACTIVE
	} else {
		go passive(masterSync, quit)
		return STATE_PASSIVE
	}
}
func getState(lastState int) {
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
	for _, unit := range(unitList) {
		if unit.Type == TYPE_MASTER {
			if unitID > unit.Port {
				return false
			}
		}
	}
	return true
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
			unitList = append(unitlsit, unit)
		}
}

func getElevStatus() map[UnitID]StatusType {
	elevReports := map[UnitID]StatusType{}
	statusChan := make(chan StatusType)
	i := 0
	length := len(unitList)

	for i < length{
		//getReport(unit, statusChan) // Should send empty report if timeout
		report := <- statusChan
		if report.ID == "" {
			unitList = append(unitList[:i],unitList[i+1:]...) //deletes unit from list
			length--
		} else {
			elevReports[report.ID] = report
			i++
		}
	}
	return elevReports
}