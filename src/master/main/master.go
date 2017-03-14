package main

import (
	"flag"
	"fmt"
	"master/masternetworkinterface"
	"runtime"
	"sync"
	"time"
	"typedef"
)

const (
	masterSyncInterval = 100 * time.Millisecond
	initWaitTime       = 3 * time.Second
	reportInterval     = 2 // in seconds

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
var elevReports = make(map[string]typedef.StatusType)

var unitMutex = &sync.Mutex{}
var reportMutex = &sync.Mutex{}

func main() {
	flag.StringVar(&id, "id", "", "The master ID")
	flag.Parse()

	fmt.Println("My ID:", id)
	fmt.Println("Master Initializing!")

	var state int
	var lastState int

	syncTimer := time.NewTicker(masterSyncInterval)
	syncChan := make(chan [][]typedef.MasterOrder, 1)
	quitChan := make(chan bool)
	doneChan := make(chan bool)

	closeChan := make(chan bool)
	newSlaveChan := make(chan typedef.UnitType, 10)

	runtime.GOMAXPROCS(runtime.NumCPU())

	go handleUnits(id, newSlaveChan, closeChan)

	initialSync()

	fmt.Println("Choosing master mode...")

	if checkIfActive() {
		go active(syncChan, newSlaveChan, doneChan, quitChan)
		fmt.Println("Going Active")
		lastState = stateActive
	} else {
		go passive(syncChan, quitChan)
		fmt.Println("Going passive")
		lastState = statePassive
	}

	time.Sleep(250 * time.Millisecond)

	for {
		state = getState(lastState)
		switch state {
		case stateGoPassive:
			quitChan <- true
			<-doneChan
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
			go active(syncChan, newSlaveChan, doneChan, quitChan)
			fmt.Println("Going Active")
			fallthrough
		case stateActive:
			select {
			case <-syncTimer.C:
				syncChan <- orderList
			default:
			}
			lastState = stateActive
		case stateQuit:
			fmt.Println("Quitting")
			close(quitChan)
			close(closeChan)
			return
		default:
			time.Sleep(masterSyncInterval)
		}
	}
}

// passive starts the passive routine in the network interface.
// It terminates them when something is received on the quit channel.
func passive(sync chan [][]typedef.MasterOrder, quitChan chan bool) {
	subQuit := make(chan bool)

	masternetworkinterface.Passive(sync, subQuit)

	<-quitChan
	close(subQuit)
	fmt.Println("Aborting passive go-routine")
}

// active requests and handles received reports, as well as handling orders
// It starts the active routine in the network interface.
// It terminates when something is received on the quit channel.
func active(sync chan [][]typedef.MasterOrder, newSlaveChan <-chan typedef.UnitType, done chan bool, quitChan chan bool) {

	fmt.Println("Orderhandling started!")
	var externalLights = make([][]bool, numFloors, numFloors)
	for i := range externalLights {
		externalLights[i] = make([]bool, 2, 2)
	}

	statusChan := make(chan typedef.StatusType, 100)
	orderRx := make(chan typedef.OrderType, 100)
	orderTx := make(chan typedef.OrderType, 100)
	lightChan := make(chan [][]bool, 10)
	subQuit := make(chan bool)

	masternetworkinterface.Active(orderTx, orderRx, sync, statusChan, lightChan, subQuit)

	for {
		select {
		case unit := <-newSlaveChan:
			if unit.Type == typedef.SLAVE {
				report, foundKey := elevReports[unit.ID]
				if foundKey {
					var order = typedef.OrderType{unit.ID, id, 0, typedef.DIR_NODIR, true}
					for floor := 0; floor < numFloors; floor++ {
						if report.MyOrders[floor][typedef.DIR_NODIR] == true {
							order.Floor = floor
							orderTx <- order
						}
					}
				}
			}

		case report := <-statusChan:
			reportMutex.Lock()
			elevReports[report.From] = report
			reportMutex.Unlock()

		case <-quitChan:
			close(subQuit)
			fmt.Println("aborting active routine")
			time.Sleep(250 * time.Millisecond)
			done <- true
			return
		case order := <-orderRx:
			handleReceivedOrder(order, externalLights)
			lightChan <- externalLights
		default:
			for i := range orderList {
				for j, order := range orderList[i] {
					diff := time.Now().Sub(order.Estimated)
					if order.Order.To == "" && order.Order.From != "" || (int(diff) > 0 && !order.Estimated.IsZero()) {
						to, estim := findSuitedSlave(order)
						if to != id {
							fmt.Println("To:", to)
							orderList[i][j].Order.To = to
							orderList[i][j].Order.From = id
							orderList[i][j].Estimated = estim
							orderTx <- orderList[i][j].Order
						}
					}
				}
			}
		}
	}
}

// initialize uses the network interface to find other Masters/Slaves and get synchronization data.
// terminates initiated go-routines after it is finished.
func initialSync() {

	quitChan := make(chan bool)
	syncChan := make(chan [][]typedef.MasterOrder, 5)
	unitChan := make(chan typedef.UnitUpdate, 5)

	numFloors = masternetworkinterface.Init_MNI(syncChan, quitChan)

	fmt.Println("Got", numFloors, "Floors")

	orderList = make([][]typedef.MasterOrder, numFloors, numFloors)
	for i := range orderList {
		orderList[i] = make([]typedef.MasterOrder, 2, 2)
	}

	timeOut := time.After(initWaitTime)

	done := false
	for done != true {
		select {
		case orderList = <-syncChan:
		case update := <-unitChan:
			units = update
		case <-timeOut:
			close(quitChan)
			done = true
		}
	}
	time.Sleep(250 * time.Millisecond)
	fmt.Println("Done Initializing Master!")
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
			if id > unit.ID && unit.ID != id {
				return false
			}
		}
	}
	return true
}

// handleReceivedOrder adds valid new orders to the global order list.
// receiving an order with New == false, means that it is executed by an elevator.
// This function writes to the global orderList and the given Light matrix.
func handleReceivedOrder(o typedef.OrderType, lights [][]bool) {
	if o.New {
		if orderList[o.Floor][o.Dir].Order.From == id && o.From != id {
			return
		}
		orderList[o.Floor][o.Dir] = typedef.MasterOrder{Order: o}
		lights[o.Floor][o.Dir] = true
		return
	}
	orderList[o.Floor][o.Dir] = typedef.MasterOrder{}
	(lights)[o.Floor][o.Dir] = false
}

// findAppropriate provides a cost function to find which slave is best suited for the order
// It calculates an estimate for when the order should be finished.
// This function reads from global values.
func findSuitedSlave(o typedef.MasterOrder) (string, time.Time) {
	cost := 10000 //high number
	chosenUnit := id
	unitMutex.Lock()
	tempUnits := make([]typedef.UnitType, len(units.Peers))
	for i, unit := range units.Peers {
		tempUnits[i] = unit
	}
	unitMutex.Unlock()

	for _, unit := range tempUnits {
		reportMutex.Lock()
		report := elevReports[unit.ID]
		reportMutex.Unlock()

		if unit.Type == typedef.SLAVE && report.From != "" {
			tempCost := 0
			floorChanges := 0
			stops := 0
			dirChanges := 0

			if report.Running && report.CurrentFloor == o.Order.Floor {
				if report.Direction == typedef.DIR_UP {
					if report.CurrentFloor < numFloors-1 {
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
				for i := report.CurrentFloor; i > -1; i-- {
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
				for i := report.CurrentFloor; i < numFloors; i++ {
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
	}
	cost += estimateBuffer
	return chosenUnit, time.Now().Add(time.Duration(cost) * time.Second) //This should be returned
}

func handleUnits(id string, newSlaveChan chan<- typedef.UnitType, quitChan chan bool) {
	unitUpdateChan := make(chan typedef.UnitUpdate)

	fmt.Println("Starting handleUnits!")

	masternetworkinterface.Peers(id, unitUpdateChan, quitChan)

	for {
		select {
		case update := <-unitUpdateChan:
			unitMutex.Lock()
			units = update
			unitMutex.Unlock()
			fmt.Println("Got Units", units)

			if update.New.ID != "" {
				newSlaveChan <- update.New
			}
		case <-quitChan:
			fmt.Println("Quitting handleUnits!")
			return
		}
	}
}
