package elevdriver

import (
	. "elevio"
	"fmt"
	"time"
	. "typedef"
)

var status StatusType

const (
	BETWEEN_FLOORS = -1
	DOOR_OPEN_TIME = 2 * time.Second
)

func Drive(quitChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, elevStatusChan chan StatusType, setLightsChan chan OrderType, initChan chan bool) {
	driveInit(initChan)
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
	for ElevGetFloorSensorSignal() == -1 {

	}
	status.CurrentFloor = ElevGetFloorSensorSignal()
	ElevMotorDirection(DIR_NODIR)
	status.Direction = DIR_NODIR
	status.Running = false
	status.MyOrders = [N_FLOORS][3]bool{{false}}
}

func getOrders(allocateOrdersChan chan OrderType) {
	select {
	case order := <-allocateOrdersChan:
		status.MyOrders[order.Floor][order.Dir] = order.New
	default:
		return
	}
}

func checkIfStop() bool {
	var relevantDirs []int
	switch status.Direction { //Only clear orders if not continuing in this direction
	case DIR_UP:
		relevantDirs = []int{DIR_UP, DIR_NODIR}
	case DIR_DOWN:
		if checkOrdersBelow(status.CurrentFloor) == false {
			relevantDirs = []int{DIR_UP, DIR_DOWN, DIR_NODIR}
		} else {
			relevantDirs = []int{DIR_DOWN, DIR_NODIR}
		}
	case DIR_NODIR:
		relevantDirs = []int{DIR_UP, DIR_DOWN, DIR_NODIR}
	}
	for dir := 0; dir < len(relevantDirs); dir++ {
		if status.MyOrders[status.CurrentFloor][relevantDirs[dir]] == true {
			return true
		}
	}
	return false
}

func determineDirection() int {
	switch status.Direction {
	case DIR_UP:
		if checkOrdersAbove(status.CurrentFloor) == true {

			fmt.Println("bump")
			return DIR_UP
		}
	case DIR_DOWN:
		if checkOrdersBelow(status.CurrentFloor) == true {
			return DIR_DOWN
		}
	default:
		if checkOrdersAbove(status.CurrentFloor) == true {
			return DIR_UP
		} else if checkOrdersBelow(status.CurrentFloor) == true {
			return DIR_DOWN
		}
	}
	return DIR_NODIR
}

func stopRoutine(executedOrdersChan chan OrderType, setLightsChan chan OrderType, allocateOrdersChan chan OrderType) {
	ElevMotorDirection(DIR_NODIR)
	status.Running = false
	ElevDoorOpenLight(true)
	status.DoorOpen = true
	clearOrder(executedOrdersChan, setLightsChan)

	doorTimer := time.NewTimer(DOOR_OPEN_TIME)
	for status.DoorOpen == true {
		select {
		case <-doorTimer.C:
			doorTimer.Stop()
			ElevDoorOpenLight(false)
			status.DoorOpen = false
		default:
			getOrders(allocateOrdersChan)
			if checkIfStop() {
				clearOrder(executedOrdersChan, setLightsChan)
				doorTimer.Reset(DOOR_OPEN_TIME)
			}
		}
	}
	/*
		//If idle and called by external button: set direction this way
		if status.Direction == DIR_NODIR {
			if status.MyOrders[status.CurrentFloor][DIR_UP] == true {
				status.Direction = DIR_UP
			} else if status.MyOrders[status.CurrentFloor][DIR_DOWN] == true {
				status.Direction = DIR_DOWN
			}
		}
	*/

}

//Heis clearer ikke ordre i nederste og øverste etasje hvis den er på vei ned, men
//Bør kontinuerlig cleare ordre i den etasjen man står i?
//Kommer til en ny etasje: Sjekk om det er flere ordre lenger fram - hvis ikke, snu retning og clear ordre den veien også

func run(dir int) {
	switch dir {
	case DIR_NODIR:
		if status.Running == true {
			ElevMotorDirection(0)
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

func checkOrdersAbove(currentFloor int) bool {
	for floor := currentFloor + 1; floor < N_FLOORS; floor++ {
		for dir := 0; dir < 3; dir++ {
			if status.MyOrders[floor][dir] == true {
				return true
			}
		}
	}
	return false
}

func checkOrdersBelow(currentFloor int) bool {
	for floor := 0; floor < currentFloor; floor++ {
		for dir := 0; dir < 3; dir++ {
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
