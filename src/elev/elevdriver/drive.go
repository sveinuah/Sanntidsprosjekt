package elevdriver

import (
	. "elev/elevio"
	"fmt"
	"time"
	. "typedef"
)

var status StatusType

const (
	BETWEEN_FLOORS int = -1
	DOOR_OPEN_TIME     = 2 * time.Second
)

func Drive(quitChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, elevStatusChan chan StatusType, setLightsChan chan OrderType, initChan chan bool) {
	driveInit(initChan)
	fmt.Println("Initialized!!!")
	for {
		getOrders(allocateOrdersChan)
		//Check floor and see if elev should stop here, move in either direction or stay put
		//Make neater
		tempFloor := ElevGetFloorSensorSignal()
		if tempFloor == BETWEEN_FLOORS {
			if status.Running == false {
				run(DIR_UP)
			}
		} else {
			if status.CurrentFloor != tempFloor { //@NewFloor
				ElevFloorIndicator(tempFloor)
				status.CurrentFloor = tempFloor
			}
			if status.DoorOpen == false {
				if checkIfStop() {
					stopRoutine(executedOrdersChan, setLightsChan, allocateOrdersChan)
				} else {
					run(determineDirection())
				}
			}
		}
		//Quit if told, update status struct if able
		select {
		case <-quitChan:
			ElevMotorDirection(DIR_NODIR)
			status.Running = false
			return
		case <-elevStatusChan:
			elevStatusChan <- status
		default:
		}

	}
}

func driveInit(initChan chan bool) {
	ElevInit()
	initChan <- true
	ElevMotorDirection(DIR_UP)
	for ElevGetFloorSensorSignal() == BETWEEN_FLOORS {
	}
	status.CurrentFloor = ElevGetFloorSensorSignal()
	fmt.Println("Freedom!")
	fmt.Println(status.CurrentFloor)
	ElevFloorIndicator(status.CurrentFloor)
	ElevMotorDirection(DIR_NODIR)
	status.Direction = DIR_NODIR
	status.Running = false
	status.MyOrders = make([][]bool, N_FLOORS, N_FLOORS)
	for i := range status.MyOrders {
		status.MyOrders[i] = make([]bool, N_BUTTONS, N_BUTTONS)
	}
}

func getOrders(allocateOrdersChan chan OrderType) {
	select {
	case order := <-allocateOrdersChan:
		status.MyOrders[order.Floor][order.Dir] = order.New
	default:
		return
	}
}

func stopRoutine(executedOrdersChan chan OrderType, setLightsChan chan OrderType, allocateOrdersChan chan OrderType) {
	ElevMotorDirection(DIR_NODIR)
	status.Running = false
	ElevDoorOpenLight(true)
	status.DoorOpen = true
	status.Direction = determineDirection()
	clearOrder(executedOrdersChan, setLightsChan)

	doorTimer := time.NewTimer(DOOR_OPEN_TIME)
	//dirSet := false
	for status.DoorOpen == true {
		select {
		case <-doorTimer.C:
			doorTimer.Stop()
			ElevDoorOpenLight(false)
			status.DoorOpen = false
			time.Sleep(100 * time.Millisecond)
		default:
			getOrders(allocateOrdersChan)
			if checkIfStop() {
				clearOrder(executedOrdersChan, setLightsChan)
				doorTimer.Reset(DOOR_OPEN_TIME)
			}
		}
	}

}

func checkIfStop() bool {
	var relevantDirs []int
	switch status.Direction { //Only clear orders if not continuing in this direction
	case DIR_UP:
		relevantDirs = []int{DIR_UP, DIR_NODIR}
	case DIR_DOWN:
		relevantDirs = []int{DIR_DOWN, DIR_NODIR}
	case DIR_NODIR:
		relevantDirs = []int{DIR_NODIR} //DIR_UP, DIR_DOWN,
	}
	for dir := 0; dir < len(relevantDirs); dir++ {
		if status.MyOrders[status.CurrentFloor][relevantDirs[dir]] == true {
			return true
		}
	}
	return false
}

func run(dir int) {
	switch dir {
	case DIR_NODIR:
		if status.Running == true {
			ElevMotorDirection(DIR_NODIR)
			status.Running = false
		}
	default:
		if status.Direction != dir || status.Running == false {
			ElevMotorDirection(dir)
			status.Running = true
		}
	}
	status.Direction = dir
}

func determineDirection() int {
	above := checkOrdersAbove(status.CurrentFloor)
	below := checkOrdersBelow(status.CurrentFloor)
	switch status.Direction {
	case DIR_UP:
		if above == true || status.MyOrders[status.CurrentFloor][DIR_UP] == true {
			return DIR_UP
		}
	case DIR_DOWN:
		if below == true || status.MyOrders[status.CurrentFloor][DIR_DOWN] == true {
			return DIR_DOWN
		}
	default:
	}
	if above == true {
		return DIR_UP
	} else if below == true {
		return DIR_DOWN
	} else if status.MyOrders[status.CurrentFloor][DIR_UP] == true {
		return DIR_UP
	} else if status.MyOrders[status.CurrentFloor][DIR_DOWN] == true {
		return DIR_DOWN
	} else {
		return DIR_NODIR
	}
}

func checkOrdersAbove(currentFloor int) bool {
	for floor := currentFloor + 1; floor < N_FLOORS; floor++ {
		for dir := 0; dir < N_BUTTONS; dir++ {
			if status.MyOrders[floor][dir] == true {
				return true
			}
		}
	}
	return false
}

func checkOrdersBelow(currentFloor int) bool {
	for floor := 0; floor < currentFloor; floor++ {
		for dir := 0; dir < N_BUTTONS; dir++ {
			if status.MyOrders[floor][dir] == true {
				return true
			}
		}
	}
	return false
}

func clearOrder(executedOrdersChan chan OrderType, setLightsChan chan OrderType) {
	var clearOrder OrderType
	clearOrder.Floor = status.CurrentFloor
	clearOrder.Dir = status.Direction
	clearOrder.New = false

	setLightsChan <- clearOrder
	status.MyOrders[status.CurrentFloor][status.Direction] = false //Clear status.MyOrders from list

	//If external order: Report to master and clear internal order aswell
	if status.Direction != DIR_NODIR {
		executedOrdersChan <- clearOrder
		clearOrder.Dir = DIR_NODIR
		setLightsChan <- clearOrder
		status.MyOrders[status.CurrentFloor][DIR_NODIR] = false
	}
}
