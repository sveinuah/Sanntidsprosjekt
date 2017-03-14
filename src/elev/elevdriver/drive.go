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

//Drive is responsible for running the elevator to service orders, including opening the door and toggling the floor indicator lamps.
func Drive(quitChan <-chan bool, allocateOrdersChan <-chan OrderType, executedOrdersChan chan<- OrderType, elevStatusChan chan StatusType, setLightsChan chan<- OrderType, initChan chan<- bool) {
	driveInit(initChan)
	fmt.Println("Drive Initialized!")

	for {
		getOrders(allocateOrdersChan)

		tempFloor := ElevGetFloorSensorSignal()
		if tempFloor == BETWEEN_FLOORS {
			if status.Running == false {
				run(DIR_UP)
			}
		} else {
			if status.CurrentFloor != tempFloor {
				ElevFloorIndicator(tempFloor)
				status.CurrentFloor = tempFloor
			}

			if status.DoorOpen == false {
				if checkIfStop() {
					stopRoutine(executedOrdersChan, setLightsChan, allocateOrdersChan)
				} else {
					run(determineNextDirection())
				}
			}
		}

		select {
		case <-quitChan:
			ElevMotorDirection(DIR_NODIR)
			status.Running = false
			return
		case <-elevStatusChan:
			elevStatusChan <- status
		default:
			elevStatusChan <- status
		}

	}
}

//Initializes Elevio and reports this to ButtonInterface.
//Runs the elevator up intil it reaches a floor, and fills in status struct.
func driveInit(initChan chan<- bool) {

	ElevInit()
	initChan <- true

	ElevMotorDirection(DIR_UP)
	for ElevGetFloorSensorSignal() == BETWEEN_FLOORS {
	}
	status.CurrentFloor = ElevGetFloorSensorSignal()
	ElevFloorIndicator(status.CurrentFloor)
	ElevMotorDirection(DIR_NODIR)

	status.Direction = DIR_NODIR
	status.Running = false
	status.MyOrders = make([][]bool, N_FLOORS, N_FLOORS)
	for i := range status.MyOrders {
		status.MyOrders[i] = make([]bool, N_BUTTONS, N_BUTTONS)
	}
}

func getOrders(allocateOrdersChan <-chan OrderType) {
	select {
	case order := <-allocateOrdersChan:
		status.MyOrders[order.Floor][order.Dir] = order.New
	default:
		return
	}
}

//Handles everything related to stopping at a floor to let passengers on and off.
//The door-open time is extended if a new order is received in the current floor, in a direction that is being serviced.
//New orders in the current direction are prioritized when driving recommences.
func stopRoutine(executedOrdersChan chan<- OrderType, setLightsChan chan<- OrderType, allocateOrdersChan <-chan OrderType) {
	ElevMotorDirection(DIR_NODIR)
	status.Running = false
	ElevDoorOpenLight(true)
	status.DoorOpen = true
	status.Direction = determineNextDirection()
	clearOrder(executedOrdersChan, setLightsChan)

	doorTimer := time.NewTimer(DOOR_OPEN_TIME)

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
	switch status.Direction {
	case DIR_UP:
		relevantDirs = []int{DIR_UP, DIR_NODIR}
	case DIR_DOWN:
		relevantDirs = []int{DIR_DOWN, DIR_NODIR}
	case DIR_NODIR:
		relevantDirs = []int{DIR_NODIR}
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

func determineNextDirection() int {
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

func clearOrder(executedOrdersChan chan<- OrderType, setLightsChan chan<- OrderType) {
	var clearOrder OrderType
	clearOrder.Floor = status.CurrentFloor
	clearOrder.Dir = status.Direction
	clearOrder.New = false

	switch {
	case status.Direction != DIR_NODIR:
		executedOrdersChan <- clearOrder
		status.MyOrders[status.CurrentFloor][status.Direction] = false

		//If order is external order: Report to master and clear internal order aswell
		clearOrder.Dir = DIR_NODIR
		fallthrough
	default:
		setLightsChan <- clearOrder
		status.MyOrders[status.CurrentFloor][DIR_NODIR] = false
	}
}
