package main

import (
	"flag"
	"fmt"
//	"master/masternetworkinterface"
	"master/testmaster"
	"time"
	"typedef"
	"sync"
	"runtime"
)

const (
	masterSyncInterval = 100 // in milliseconds
	initWaitTime       = 3   // in seconds
	reportInterval     = 2   // in seconds

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

var id string
var numFloors int
var units typedef.UnitUpdate
var orderList [][]typedef.MasterOrder

var unitMutex = &sync.Mutex{}


func main() {
	initialize()

	var state int
	var lastState int

	syncTimer := time.Tick(masterSyncInterval * time.Millisecond)
	syncChan := make(chan [][]typedef.MasterOrder)
	quitChan := make(chan bool)
	doneChan := make(chan bool)

	if checkIfActive() {
		go active(syncChan, doneChan, quitChan)
		fmt.Println("Going Active")
		lastState = stateActive
	} else {
		go passive(syncChan, quitChan)
		fmt.Println("Going passive")
		lastState = statePassive
	}

	for {
		state = getState(lastState)
		switch state {
		case stateGoPassive:
			quitChan <- true
			<- doneChan
			go passive(syncChan, quitChan)
			fmt.Println("Going passive")
			fallthrough
		case statePassive:
			select {
			case orderList = <-syncChan:
			default:
			}
			lastState = statePassive
		case stateGoActive:
			quitChan <- true
			time.Sleep(100 * time.Millisecond)
			go active(syncChan, doneChan, quitChan)
			fmt.Println("Going Active")
			fallthrough
		case stateActive:
			select {
			case <-syncTimer:
				syncChan <- orderList
			default:
			}
			lastState = stateActive
		case stateQuit:
			fmt.Println("Quitting")
			close(quitChan)
			return
		default:
		}
	}
}

// initialize uses the network interface to find other Masters/Slaves and get synchronization data.
// terminates initiated go-routines after it is finished.
func initialize() {
	flag.StringVar(&id, "id", "", "The master ID")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())
	
	fmt.Println(id)

	fmt.Println("Master Initializing!")
	fmt.Println(typedef.MasterOrder{})
	fmt.Println(typedef.OrderType{})

	quitChan := make(chan bool)
	syncChan := make(chan [][]typedef.MasterOrder)
	unitChan := make(chan typedef.UnitUpdate)

	numFloors = test.Init(id, syncChan, unitChan, quitChan)//masternetworkinterface.Init(id, syncChan, unitChan, quitChan)

	fmt.Print("Got ", numFloors, " Floors")

	orderList = make([][]typedef.MasterOrder, numFloors, numFloors)
	for i := range orderList {
		orderList[i] = make([]typedef.MasterOrder, 2, 2)
	}

	timeOut := time.After(initWaitTime * time.Second)

	done := false
	for done != true {
		select {
		case orderList = <-syncChan:
		case update := <-unitChan:
			units = update
		case <-timeOut:
			done = true
		}
	}

	close(quitChan)
	time.Sleep(100*time.Millisecond)
	fmt.Println("Done Initializing Master!")
}

// passive updates unit list
// It starts the passive routine in the network interface.
// It terminates them when something is received on the quit channel.
func passive(sync chan [][]typedef.MasterOrder, quitChan chan bool) {
	unitChan := make(chan typedef.UnitUpdate, 1)
	subQuit := make(chan bool)

	test.Passive(sync, unitChan, subQuit)//masternetworkinterface.Passive(sync, unitChan, quitChan)

	for {
		select {
		case update := <-unitChan:
			unitMutex.Lock()
			units = update
			unitMutex.Unlock()
			fmt.Println("Got Units", units)
		case <-quitChan:
			close(subQuit)
			fmt.Println("Aborting passive go-goutine")
			return
		}
	}
}

// active updates unit list, requests and handles received reports
// It starts the active routine in the network interface, and the order handling go-routine.
// It terminates when something is received on the quit channel.
func active(sync chan [][]typedef.MasterOrder, done chan bool, quitChan chan bool) {
	reportNum := 1

	elevReports := make(map[string]typedef.StatusType)

	statusReqChan := make(chan int, 1) //to request reports with id
	statusChan := make(chan typedef.StatusType, 10)
	unitChan := make(chan typedef.UnitUpdate, 1)
	orderRx := make(chan typedef.OrderType, 100)
	orderTx := make(chan typedef.OrderType, 100)
	lightChan := make(chan [][]bool, 1)
	subQuit := make(chan bool)

	reportAccess := make(chan bool, 1)
	ordersDone := make(chan bool)
	reportAccess <- true

	test.Active(unitChan, orderTx, orderRx, sync, statusChan, statusReqChan, lightChan, subQuit)//masternetworkinterface.Active(unitChan, orderTx, orderRx, sync, statusChan, statusReqChan, lightChan, quitChan)

	go orders(elevReports, reportAccess, orderRx, orderTx, lightChan, subQuit, ordersDone)

	reportTime := time.Tick(reportInterval * time.Second)

	for {
		select {
		case update := <-unitChan:
			unitMutex.Lock()
			units = update
			unitMutex.Unlock()
			fmt.Println("Got Units", units)
		case report := <-statusChan:
			fmt.Println("Got report", report)
			<- reportAccess
			elevReports[report.From] = report
			reportAccess <- true
		case <-reportTime:
			fmt.Println("Request report!")
			reportNum++
			statusReqChan <- reportNum
		//case <-errChan:
		//handleErr
		case <-quitChan:
			close(subQuit)
			<- ordersDone
			done <- true
			fmt.Println("aborting active routine")
			return
		default:
		}
	}
}

// orders handles all incoming orders, as well as delegating them to the slaves.
// It is made to run as a go-routine.
// It terminates when something is received on the quit channel.
func orders(reports map[string]typedef.StatusType, reportAccess chan bool, orderRx chan typedef.OrderType, orderTx chan typedef.OrderType, lightChan chan [][]bool, quitChan chan bool, done chan<- bool) {
	var externalLights = make([][]bool, numFloors, numFloors)
	for i := range externalLights {
		externalLights[i] = make([]bool, 2, 2)
	}
	for {
		select {
		case order := <-orderRx:
			handleNewOrder(order, &externalLights)
			lightChan <- externalLights
			//fmt.Println("Order:", order)
		case <-quitChan:
			done <- true
			fmt.Println("go orders aborting")
			return
		default:
			for i := range orderList {
				for j, order := range orderList[i] {
					diff := time.Now().Sub(order.Estimated)
					if order.Order.To == id || (int(diff) > 0 && !order.Estimated.IsZero()) {
						to, estim := findAppropriate(reports, reportAccess, order) // 2 sek pr etasje + 2 sek pr stop + leeway
						
						if to != id {
							orderList[i][j].Order.To = to
							orderList[i][j].Order.From = id
							orderList[i][j].Estimated = estim
							orderTx <- orderList[i][j].Order
						} else {
							orderLst[i][j].Order.To = id
						}
					}
				}
			}
		}
	}
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

// This function reads from the global variable units
func checkIfActive() bool {
	unitMutex.Lock()
	defer unitMutex.Unlock()
	for _, unit := range units.Peers {
		if unit.Type == typedef.MASTER {
			if id > unit.ID {
				return false
			}
		}
	}
	return true
}

// handleOrders adds all new orders sent to "", to the order list.
// receiving a false order, means that it is executed by an elevator.
// This function writes to the global orderList and the given Light matrix.
func handleNewOrder(o typedef.OrderType, lights *[][]bool) {
	if o.New {
		if orderList[o.Floor][o.Dir].Order.To == "" {
			orderList[o.Floor][o.Dir] = typedef.MasterOrder{Order: o}
			(*lights)[o.Floor][o.Dir] = true
		}
		return
	}
	orderList[o.Floor][o.Dir] = typedef.MasterOrder{} //clear order
	(*lights)[o.Floor][o.Dir] = false
}

// findAppropriate provides a cost function to find which slave is best suited for the order
// It calculates an estimate for when the order should be finished.
// This function reads from global values.
func findAppropriate(reports map[string]typedef.StatusType, reportAccess chan bool, o typedef.MasterOrder) (string, time.Time) {
	cost := 10000 //high number
	chosenUnit := id

	unitMutex.Lock()
	for _, unit := range units.Peers {
		unitMutex.Unlock()
		/*For testing purposes--------------------------
		if unit.Type == typedef.SLAVE {
			return unit.ID, time.Now().Add(10 * time.Second)
		}
		//----------------------------------------------*/
		//ACTUAL ALGORITHM
		<- reportAccess
		report := reports[unit.ID]
		reportAccess <- true

		if unit.Type == typedef.SLAVE && report.From != "" {
			tempCost := 0
			floorChanges := 0
			stops := 0
			dirChanges := 0

			if report.Running && report.CurrentFloor == o.Order.Floor {
				if report.Direction == typedef.DIR_UP {
					if report.CurrentFloor < numFloors - 1 {
						report.CurrentFloor++
					}
				}
				if report.CurrentFloor > 0 {
					report.CurrentFloor--
				}
			}

			if o.Order.Floor > report.CurrentFloor && report.Direction == typedef.DIR_DOWN {
				dirChanges = 1
				if o.Order.Dir == typedef.DIR_DOWN {
					dirChanges++
				}

				lowestFloor := report.CurrentFloor
				for i:= report.CurrentFloor; i > -1; i-- {
					switch {
						case report.MyOrders[i][typedef.DIR_DOWN] == true:
							fallthrough
						case report.MyOrders[i][typedef.DIR_NODIR] == true:
							stops++
							lowestFloor = i
					}
				}

				for i := lowestFloor; i < o.Order.Floor; i++ {
					switch {
						case report.MyOrders[i][typedef.DIR_UP] == true:
							fallthrough
						case report.MyOrders[i][typedef.DIR_NODIR] == true:
							stops++
					}
				}

				floorChanges = (report.CurrentFloor - lowestFloor) + (o.Order.Floor - lowestFloor)

			} else if o.Order.Floor > report.CurrentFloor && report.Direction == typedef.DIR_UP {
				if o.Order.Dir == typedef.DIR_DOWN {
					dirChanges = 1
				}

				for i := report.CurrentFloor; i < o.Order.Floor; i++ {
					switch {
						case report.MyOrders[i][typedef.DIR_UP] == true:
							fallthrough
						case report.MyOrders[i][typedef.DIR_NODIR] == true:
							stops++
					}
				}

				floorChanges = o.Order.Floor - report.CurrentFloor

			} else if o.Order.Floor < report.CurrentFloor && report.Direction == typedef.DIR_DOWN {
				if o.Order.Dir == typedef.DIR_UP {
					dirChanges = 1
				}

				for i := o.Order.Floor; i < report.CurrentFloor; i++ {
					switch {
						case report.MyOrders[i][typedef.DIR_DOWN] == true:
							fallthrough
						case report.MyOrders[i][typedef.DIR_NODIR] == true:
							stops++
					}
				}

				floorChanges = report.CurrentFloor - o.Order.Floor

			} else if o.Order.Floor < report.CurrentFloor && report.Direction == typedef.DIR_UP {
				dirChanges = 1
				if o.Order.Dir == typedef.DIR_UP {
					dirChanges++
				}

				highestFloor := report.CurrentFloor
				for i:= report.CurrentFloor; i < numFloors + 1; i++ {
					switch {
						case report.MyOrders[i][typedef.DIR_UP] == true:
							fallthrough
						case report.MyOrders[i][typedef.DIR_NODIR] == true:
							stops++
							highestFloor = i
					}
				}

				for i := highestFloor; i > o.Order.Floor; i-- {
					switch {
						case report.MyOrders[i][typedef.DIR_DOWN] == true:
							fallthrough
						case report.MyOrders[i][typedef.DIR_NODIR] == true:
							stops++
					}
				}

				floorChanges = (highestFloor - report.CurrentFloor) + (highestFloor - o.Order.Floor)

			} else {
				if o.Order.Floor > report.CurrentFloor {
					floorChanges = o.Order.Floor - report.CurrentFloor

					if o.Order.Dir == typedef.DIR_DOWN {
						dirChanges = 1
					}
				} else {
					floorChanges = report.CurrentFloor - o.Order.Floor

					if o.Order.Dir == typedef.DIR_UP {
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
		unitMutex.Lock()
	}
	unitMutex.Unlock()
	//return id, time.Time{} //bounce back to myself
	cost += estimateBuffer
	return chosenUnit, time.Now().Add(time.Duration(cost) * time.Second) //This should be returned
}