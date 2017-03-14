package main

import (
	"flag"
	"fmt"
	"master/masternetworkinterface"
	"math/rand"
	"runtime"
	"sync"
	"time"
	. "typedef"
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
	floorChangeCost int = 3
	dirChangeCost   int = 6
	timeOutCost     int = 10
	estimateBuffer  int = 2
)

var id string
var numFloors int
var units []UnitType
var orderList [][]MasterOrder
var elevReports = make(map[string]StatusType)

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
	syncChan := make(chan [][]MasterOrder, 1)
	quitChan := make(chan bool)
	doneChan := make(chan bool)

	closeChan := make(chan bool)
	newSlaveChan := make(chan UnitType, 10)

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
func passive(sync chan [][]MasterOrder, quitChan chan bool) {
	subQuit := make(chan bool)

	masternetworkinterface.Passive(sync, subQuit)

	<-quitChan
	close(subQuit)
	fmt.Println("Aborting passive go-routine")
}

// active requests and handles received reports, as well as handling orders
// It starts the active routine in the network interface.
// It terminates when something is received on the quit channel.
func active(sync chan [][]MasterOrder, newSlaveChan <-chan UnitType, done chan bool, quitChan chan bool) {

	var externalLights = make([][]bool, numFloors, numFloors)
	for i := range externalLights {
		externalLights[i] = make([]bool, 2, 2)
	}

	statusChan := make(chan StatusType, 100)
	orderRx := make(chan OrderType, 100)
	orderTx := make(chan OrderType, 100)
	lightChan := make(chan [][]bool, 10)
	subQuit := make(chan bool)

	masternetworkinterface.Active(orderTx, orderRx, sync, statusChan, lightChan, subQuit)

	for {
		select {
		case unit := <-newSlaveChan:
			if unit.Type == SLAVE {
				report, foundKey := elevReports[unit.ID]
				if foundKey {
					var order = OrderType{unit.ID, id, 0, DIR_NODIR, true}
					for floor := 0; floor < numFloors; floor++ {
						if report.MyOrders[floor][DIR_NODIR] == true {
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
							fmt.Println("To:", to, "Estimate:", estim.Sub(time.Now()).String())
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

// handleUnits handles updating of the unit list, and should be run as a go-routine
// It starts the Peers function in the masterNetworkInterface
// It terminates when something is received on the quit channel.
func handleUnits(id string, newSlaveChan chan<- UnitType, quitChan chan bool) {
	unitUpdateChan := make(chan UnitUpdate)

	masternetworkinterface.Peers(id, unitUpdateChan, quitChan)

	for {
		select {
		case update := <-unitUpdateChan:
			unitMutex.Lock()
			units = update.Peers
			unitMutex.Unlock()
			fmt.Println("Got Units", update)

			if update.New.ID != "" {
				newSlaveChan <- update.New
			}
		case <-quitChan:
			return
		}
	}
}

// initialSync uses the network interface to find other Masters/Slaves and get synchronization data.
// It also allocates memory for the global order list, based on the number of floors received from Init_MNI.
// It terminates initiated go-routines after timeout is received.
func initialSync() {

	quitChan := make(chan bool)
	syncChan := make(chan [][]MasterOrder, 5)

	numFloors = masternetworkinterface.Init_MNI(syncChan, quitChan)

	fmt.Println("Got", numFloors, "Floors")

	orderList = make([][]MasterOrder, numFloors, numFloors)
	for i := range orderList {
		orderList[i] = make([]MasterOrder, 2, 2)
	}

	timeOut := time.After(initWaitTime)

	done := false
	for done != true {
		select {
		case orderList = <-syncChan:
		case <-timeOut:
			close(quitChan)
			done = true
		}
	}
	time.Sleep(250 * time.Millisecond)
	fmt.Println("Done Initializing Master!")
}

func getState(lastState int) int {
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

// checkIfActive reads from the global variable units.
// It uses the mutex unitMutex to prevent race conditions.
func checkIfActive() bool {
	unitMutex.Lock()
	defer unitMutex.Unlock()
	for _, unit := range units {
		if unit.Type == MASTER {
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
func handleReceivedOrder(o OrderType, lights [][]bool) {
	if o.New {
		if orderList[o.Floor][o.Dir].Order.From == id && o.From != id {
			return
		}
		orderList[o.Floor][o.Dir] = MasterOrder{Order: o}
		lights[o.Floor][o.Dir] = true
		return
	}
	orderList[o.Floor][o.Dir] = MasterOrder{}
	(lights)[o.Floor][o.Dir] = false
}

// findSuitedSlave provides a cost function to find which slave is best suited for the order
// It calculates an estimate for when the order should be finished.
// This function reads from global values, and uses the mutexes unitMutex and reportMutex.
func findSuitedSlave(o MasterOrder) (string, time.Time) {
	cost := 10000 //high number
	chosenUnit := id
	unitMutex.Lock()
	tempUnits := make([]UnitType, len(units))
	for i, unit := range units {
		tempUnits[i] = unit
	}
	unitMutex.Unlock()

	for _, unit := range tempUnits {
		reportMutex.Lock()
		report := elevReports[unit.ID]
		reportMutex.Unlock()
		if unit.Type == MASTER || report.From == "" {
			continue
		}

		tempCost := 0
		floorChanges := 0
		stops := 0
		dirChanges := 0

		orderFloor := o.Order.Floor
		orderDir := o.Order.Dir
		reportFloor := report.CurrentFloor
		reportDir := report.Direction

		// If the order is already delegated to an elevator, it means that it has reached the timeout
		if report.MyOrders[orderFloor][orderDir] {
			tempCost += timeOutCost
		}

		switch {
		case reportDir == DIR_NODIR:

			floorChanges = abs(orderFloor - reportFloor)

		case orderFloor > reportFloor && reportDir == DIR_DOWN:

			dirChanges = 1
			if orderDir == DIR_DOWN {
				dirChanges++
			}
			lowestFloor := reportFloor
			for i := reportFloor; i > 0; i-- {
				if report.MyOrders[i][DIR_DOWN] {
					stops++
					lowestFloor = i - 1
				}
			}

			for i := lowestFloor; i < orderFloor; i++ {
				switch {
				case report.MyOrders[i][DIR_UP]:
					fallthrough
				case report.MyOrders[i][DIR_NODIR]:
					stops++
				}
			}
			floorChanges = (reportFloor - lowestFloor) + (orderFloor - lowestFloor)

		case orderFloor > reportFloor && reportDir == DIR_UP:

			if orderDir == DIR_DOWN {
				dirChanges = 1
			}
			for i := reportFloor; i < orderFloor; i++ {
				switch {
				case report.MyOrders[i][DIR_UP]:
					fallthrough
				case report.MyOrders[i][DIR_NODIR]:
					stops++
				}
			}

			floorChanges = orderFloor - reportFloor

		case orderFloor < reportFloor && reportDir == DIR_DOWN:

			if orderDir == DIR_UP {
				dirChanges = 1
			}
			for i := reportFloor; i > orderFloor; i-- {
				switch {
				case report.MyOrders[i][DIR_DOWN]:
					fallthrough
				case report.MyOrders[i][DIR_NODIR]:
					stops++
				}
			}

			floorChanges = reportFloor - orderFloor

		case orderFloor < reportFloor && reportDir == DIR_UP:

			dirChanges = 1
			if orderDir == DIR_UP {
				dirChanges++
			}

			highestFloor := reportFloor
			for i := reportFloor; i < numFloors; i++ {
				if report.MyOrders[i][DIR_UP] {
					stops++
					highestFloor = i
				}
			}

			for i := highestFloor; i > orderFloor; i-- {
				switch {
				case report.MyOrders[i][DIR_DOWN]:
					fallthrough
				case report.MyOrders[i][DIR_NODIR]:
					stops++
				}
			}

			floorChanges = (highestFloor - reportFloor) + (highestFloor - orderFloor)

		default:
		}
		tempCost += floorChangeCost * floorChanges
		tempCost += stopCost * stops
		tempCost += dirChangeCost * dirChanges

		if tempCost < cost {
			chosenUnit = unit.ID
			cost = tempCost
		} else if tempCost == cost {
			chosenUnit = chooseRandom(chosenUnit, unit.ID)
		}
	}
	cost += estimateBuffer
	fmt.Println(elevReports[chosenUnit])
	return chosenUnit, time.Now().Add(time.Duration(cost) * time.Second)
}

func chooseRandom(roger string, rud string) string {
	arr := []string{roger, rud}
	return arr[rand.Int()%2]
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
