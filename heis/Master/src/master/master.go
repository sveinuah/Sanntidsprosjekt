package main

import (
	"../typedef"
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


var active bool
var unitID int
var unitList [] UnitType
var masterOrderList [] Order

func checkIfActive() {
	if unitList[0].Port == unitID {
		active = true
	} else {
		active = false
	}
}

func init() {
	// Find other units
	// If other masters:
		// update from other masters
		// ask for free port
	// else
		// unitID = MIN_MASTER_PORT
	// check if active
	// initialize variables
	checkIfActive()
}

func main() {
	//Init
	orderChan = make(chan OrderPackage)
	unitStatusChan = make(chan UnitType)
	reportChan = make(chan elevatorReport)
	masterSync = make(chan [] Order)

	for {
		
		switch active {
		case true:
			select {
			case unit := <- unitStatusChan:
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
