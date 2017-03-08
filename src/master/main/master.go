package main

import (
	"flag"
	"fmt"
	"master/masternetworkinterface"
	"time"
	"typedef"
)

const (
	masterSyncInterval = 100 // in milliseconds
	initWaitTime       = 3   // in seconds
	reportInterval     = 2   //in seconds

	stateGoActive  int = 0
	stateActive    int = 1
	stateGoPassive int = 2
	statePassive   int = 3
	stateQuit      int = 4

	stopCost        int = 2
	floorChangeCost int = 2
	dirChangeCost   int = 6
	estimateBuffer  int = 2
)

var id typedef.UnitID
var numFloors int
var units typedef.UnitUpdate
var orderList [][]typedef.MasterOrder //numFloors+2directions

func main() {

	syncChan := make(chan [][]typedef.MasterOrder)
	quit := make(chan bool)

	syncTimer := time.Tick(masterSyncInterval * time.Millisecond)

	init()

	lastState := -1
	var state int

	for {
		state = getState(lastState)
		switch state {
		case stateGoPassive:
			close(quit)
			quit := make(chan bool)
			go passive(syncChan, quit)
			fmt.Println("Going passive")
			fallthrough
		case statePassive:
			select {
			case orderList = <-syncChan:
			default:
			}
		case stateGoActive:
			close(quit)
			quit := make(chan bool)
			go active(sync, quit)
			fmt.Println("Going Active")
			fallthrough
		case stateActive:
			select {
			case <-syncTimer:
				syncChan <- orderList
			default:
			}
		case stateQuit:
			fmt.Println("Quitting")
			close(quit)
			return
		default:
		}
	}
}

func passive(sync chan [][]typedef.MasterOrder, quit chan bool) {
	unitChan := make(chan typedef.UnitUpdate)

	networkinterface.Passive(sync, unitChan, quit)

	for {
		select {
		case units = <-unitChan:
		case <-quit:
			fmt.Println("Aborting passive go-goutine")
			return
		}
	}
}

func active(sync chan [][]typedef.MasterOrder, quit chan bool) {
	reportNum := 1

	elevReports := make(map[typedef.UnitID]typedef.StatusType)

	statusReqChan := make(chan int) //to request reports with id
	statusChan := make(chan typedef.StatusType)
	unitChan := make(chan typedef.UnitUpdate)
	orderRx := make(chan typedef.OrderType)
	orderTx := make(chan typedef.OrderType)
	lightChan := make(chan [][]bool)

	networkinterface.Active(unitChan, orderTx, orderRx, sync, statusChan, statusReqChan, lightChan, quit)

	go orders(elevReports, orderRx, orderTx, lightChan, quit)

	reportTime := time.Tick(reportInterval * time.Second)

	for {
		select {
		case units := <-unitChan:
		case report <- statusChan:
			elevReports[report.From] = report
		case <-reportTime:
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

func orders(reports *map[typedef.UnitID]typedef.StatusType, orderRx chan typedef.OrderType, orderTx chan typedef.OrderType, lightChan chan [][]bool, quit chan bool) {
	var externalLights [numFloors][3]bool

	for {
		select {
		case order := <-orderRx:
			orderCapsule := typedef.MasterOrder{order}
			handleNewOrder(orderCapsule, externalLights)
			lightChan <- externalLights
		case <-quit:
			fmt.Println("go orders aborting")
			return
		default:
			for i := range orderList {
				for j, order := range orderList[i] {
					diff := time.Now().Sub(order.Estimated)

					if order.To == nil || diff > 0 {
						to, estim := findAppropriate(order) // 2 sek pr etasje + 2 sek pr stop + leeway
						orderList[i][j].To = to
						orderList[i][j].From = id
						orderLIst[i][j].Delegated = time.Now()
						orderList[i][j].Estimated = estim

						txOrder := typedef.OrderType{orderList[i][j].o}
						orderTx <- txOrder
					}
				}
			}
		}
	}
}

func init() int {
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	fmt.PrintLn("Initializing!")

	quit := make(chan bool)
	syncChan := make(chan [][]typedef.masterOrder)
	unitChan := make(chan typedef.UnitUpdate)

	numFloors = networkinterface.Init(id, syncChan, unitChan, quit)

	orderList = make([numFloors][3]typedef.masterOrder)

	timeOut := time.After(initWaitTime * time.Second)

	done := false
	for done != true {
		select {
		case orderList = <-syncChan:
		case units = <-unitChan:
		case <-timeOut:
			done = true
		}
	}

	close(quit)
	fmt.Println("Done Initializing!")

	if ckeckIfActive() {
		return stateGoActive
	}
	return stateGoPassive
}

func getState(lastState int) int {
	// get and return stateQuit when quit flag is raised NYI
	if checkIfActive() {
		if lastState == stateActive {
			return stateActive
		}
		return stateGoActive
	} else if lastState == statePassive {
		return statePassive
	}
	return stateGoPassive
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

func handleNewOrder(o typedef.MasterOrder, lights *[][]bool) {
	if o.New {
		if orderList[o.Floor][o.Dir] == nil {
			orderList[o.Floor][o.Dir] = o
			lights[o.Floor][o.Dir] = true
		}
		return
	}
	orderList[o.Floor][o.Dir] = typedef.MasterOrder{} //clear order
	lights[o.Floor][o.Dir] = false
}

func findAppropriate(o typedef.MasterOrder) {
	/*
		cost := 10000 //high number
		chosenUnit := id
	*/
	for i, unit := range units.Peers {
		//For testing purposes--------------------------
		if unit.Type == TYPE_SLAVE {
			return unit.ID, time.Now().Add(10 * time.Second)
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
			tempCost += floorChangeCost * floorChanges
			tempCost += stopCost * stops
			tempCost += dirChangeCost * dirChanges

			if tempCost < cost {
				chosenUnit = unit.ID
				cost = tempCost
			}
		}




		*/
	}
	return id, time.Time{} //bounce back to myself
	//return chosenUnit, time.Now().Add((cost+estimateBuffer) * time.Second) //This should be returned
}
