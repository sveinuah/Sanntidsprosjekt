package main

import (
	"../orderqueue"
	"../typedef"
	"log"
	"time"
)

const (
	MASTER_SYNC_INTERVALL   int = 100 // in milliseconds
	INITIALIATION_WAIT_TIME int = 3   // in seconds
	REPORT_INTERVALL        int = 2   //in seconds

	STATE_GO_ACTIVE  int = 0
	STATE_ACTIVE     int = 1
	STATE_GO_PASSIVE int = 2
	STATE_PASSIVE    int = 3
	STATE_QUIT       int = 4

	COST_STOP		int = 2
	COST_FLOOR_CHANGE int = 2
	COST_DIR_CHANGE int = 6
	ESTIMATE_BUFFER int = 2
)

var id UnitID
var numFloors int
var units networkmodule.UnitUpdate
var orderList [][]masterOrder //numFloors+2directions

type masterOrder struct {
	o OrderType
	delegated time.Time
	estimated time.Time
}

func main() {

	syncChan := make(chan [][]masterOrder)
	quit := make(chan bool)

	syncTimer := time.Tick(MASTER_SYNC_INTERVALL * time.Millisecond)

	lastState := init(syncChan)
	state := lastState

	for {
		state = getState(lastState)
		switch state {
		case STATE_GO_PASSIVE:
			close(quit)
			fmt.Println("Going passive")
			fallthrough
		case STATE_PASSIVE:
			select {
			case orderList = <-syncChan:
			default:
			}
		case STATE_GO_ACTIVE:
			quit := make(chan bool)
			go active(quit)
			fallthrough
		case STATE_ACTIVE:
			select {
			case <-syncTimer:
				syncChan <- orderList
			default:
			}
		case STATE_QUIT:
			fmt.Println("Quitting")
			close(quit)
			return
		default:
		}
	}
}

func active(quit chan bool) { // not finished
	reportNum := 1

	elevReports := make(map[UnitID]StatusType)

	statusReqChan := make(chan int) //to request reports with id
	statusChan := make(chan StatusType)
	unitChan := make(chan networkinterface.UnitUpdate)

	//initialize network for active NYI

	go orders(elevReports, quit)

	reportTime := time.Tick(REPORT_INTERVALL * time.Second)

	for {
		select {
		case  units := <-unitChan:
			// addUnit(unit) antagelig unødvendig ved bruk av peers
		case report <- statusChan:
			elevReports[report.From] = report
		case <-reportTime:
			// handleDeadUnits(elevReports, reportNum) unødvendig ved bruk av peers
			reportNum++
			statusReqChan <- reportNum
		case err <- errChan:
			//handleErr
		case <-quit:
			fmt.Println("aborting active routine")
			return
		default:
		}
	}
}

func orders(reports *map[UnitID]StatusType, quit chan bool) {
	var externalLights [numFloors][3] bool = {}

	orderRx := make(chan OrderType)
	orderTx := make(chan OrderType)
	lightChan := make(chan [][]bool)

	//initialize network for orders 

	for {
		select {
		case order := <-orderRx:
			orderCapsule := masterOrder{order}
			handleNewOrder(orderCapsule)
			lightChan <- lights
		case <-quit:
			fmt.Println("go orders aborting")
			return
		default:
			for i := range(orderList) {
				for j, order := range(orderList[i]) {
					diff := time.Now().Sub(order.Estimated)

					if order.To == nil || diff > 0 {
						to, estim := findAppropriate(order) // 2 sek pr etasje + 2 sek pr stop + leeway
						orderList[i][j].To = to
						orderList[i][j].From = id
						orderLIst[i][j].Delegated = time.Now()
						orderList[i][j].Estimated = estim

						txOrder := OrderType{orderList[i][j].o}
						orderTx <- txOrder
					}
				}
			}
			
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

/*
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
}*/

func handleNewOrder(o masterOrder) {
	if o.New {
		if orderList[o.Floor][o.Dir] == nil {
			orderList[o.Floor][o.Dir] = o
			lights[o.Floor][o.Dir] = true
		}
		return
	}
	if o.To == 
	orderList[o.Floor][o.Dir] = masterOrder{} //clear order
	lights[o.Floor][o.Dir] = false
}

func findAppropriate(o masterOrder) {
	/*
	cost := 10000 //high number
	chosenUnit := id
	*/
	for i, unit := range units.Peers {
		//For testing purposes--------------------------
		if unit.Type == TYPE_SLAVE {
			return unit.ID, time.Now().Add(10*time.Second)
		}
		//----------------------------------------------
		/* ACTUAL ALGORITHM
		if unit.Type == TYPE_SLAVE {
			tempCost := 0
			floorChanges := 0
			stops := 0
			dirChanges := 0
			report := elevReports[unit.ID]

			if report.Running && report.Floor == o.Floor {
				if report.Dir == DIR_UP {
					if report.Floor < numFloors {
						report.Floor++
					}
				}
				if report.Floor > 0 {
					report.Floor--
				}
			}

			if o.Floor > report.Floor && report.Dir == DIR_DOWN {
				dirChanges = 1
				if o.Dir == DIR_DOWN {
					dirChanges++
				}

				lowestFloor := report.Floor
				for i:= report.Floor; i > 0; i-- {
					switch {
						case report.MyOrders[i][DIR_DOWN] == true:
							fallthrough
						case report.Myorders[i][DIR_NODIR] == true:
							stops++
							lowestFloor = i
					}
				}

				for i := lowestFloor; i < o.Floor; i++ {
					switch {
						case report.MyOrders[i][DIR_UP] == true:
							fallthrough
						case report.MyOrders[i][DIR_NODIR] == true:
							stops++
					}
				}

				floorChanges = (report.Floor - lowerFloor) + (o.Floor - lowestFloor)

			} else if o.Floor > report.Floor && report.Dir == DIR_UP {
				if o.Dir == DIR_DOWN {
					dirChanges = 1
				}

				for i := report.Floor; i < o.Floor; i++ {
					switch {
						case report.MyOrders[i][DIR_UP] == true:
							fallthrough
						case report.MyOrders[i][DIR_NODIR] == true:
							stops++
					}
				}

				floorChanges = o.Floor - report.Floor	

			} else if o.Floor < report.Floor && report.Dir == DIR_DOWN {
				if o.Dir == DIR_UP {
					dirChanges = 1
				}

				for i := o.Floor; i < report.Floor; i++ {
					switch {
						case report.MyOrders[i][DIR_DOWN] == true:
							fallthrough
						case report.MyOrders[i][DIR_NODIR] == true:
							stops++
					}
				}

				floorChanges = report.Floor - o.Floor
				
			} else if o.Floor < report.Floor && report.Dir == DIR_UP {
				dirChanges = 1
				if o.Dir == DIR_UP {
					dirChanges++
				}

				highestFloor := report.Floor
				for i:= report.Floor; i < numFloors + 1; i++ {
					switch {
						case report.MyOrders[i][DIR_UP] == true:
							fallthrough
						case report.Myorders[i][DIR_NODIR] == true:
							stops++
							highestFloor = i
					}
				}

				for i := highestFloor; i > o.Floor; i-- {
					switch {
						case report.MyOrders[i][DIR_DOWN] == true:
							fallthrough
						case report.MyOrders[i][DIR_NODIR] == true:
							stops++
					}
				}

				floorChanges = (highestFloor - report.Floor) + (highestFloor - o.Floor)
	
			} else {
				if o.Floor > report.Floor {
					floorChanges = o.Floor - report.Floor

					if o.Dir == DIR_DOWN {
						dirChanges = 1
					}
				} else {
					floorChanges = report.Floor - o.Floor

					if o.Dir == DIR_UP {
						dirChanges = 1
					}
				}
			}
			tempCost += COST_FLOOR_CHANGE * floorChanges
			tempCost += COST_STOP * stops
			tempCost += COST_DIR_CHANGE * dirChanges

			if tempCost < cost {
				chosenUnit = unit.ID
				cost = tempCost
			}
		}




		*/
	}
	return id, time.Time{} //bounce back to myself
	//return chosenUnit, time.Now().Add((cost+ESTIMATE_BUFFER) * time.Second) //This should be returned
}