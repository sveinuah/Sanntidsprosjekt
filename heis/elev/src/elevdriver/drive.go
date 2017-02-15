package elevdriver

import (
	. "elevio"
	"time"
	. "typedef"
)

var (
	status    StatusType
	numFloors int
)

func Drive(abortChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, elevStatusChan chan StatusType, setLightsChan chan OrderType, initChan chan int) {
	driveInit(initChan)
	abortFlag := false
	for abortFlag != true {
		//Get all status.Orders from allocateOrdersChan and place in status.Orders
		ordersInChannel := true
		for ordersInChannel {
			select {
			case order := <-allocateOrdersChan:
				status.Orders[order.Floor][order.Dir] = order.New
			default:
				ordersInChannel = false
			}
		}
		//Check floor and see if elev should stop here, move in either direction or stay put
		tempFloor := ElevGetFloorSensorSignal()
		if tempFloor == -1 {
			if status.Running == false {
				run(DIR_UP)
			}
		} else {
			if status.CurrentFloor != tempFloor {
				ElevFloorIndicator(tempFloor)
				status.CurrentFloor = tempFloor
			}
			if status.Orders[status.CurrentFloor][status.Direction] == true || status.Orders[status.CurrentFloor][DIR_NODIR] == true { //order to stop here: stop here
				stopRoutine(executedOrdersChan, setLightsChan)
			} else {
				if status.Direction == DIR_DOWN {
					if checkOrdersBelow(status.CurrentFloor) == true {
						run(DIR_DOWN)
					} else if checkOrdersAbove(status.CurrentFloor) == true {
						run(DIR_UP)
					} else {
						run(DIR_NODIR)
					}
				} else {
					if checkOrdersAbove(status.CurrentFloor) == true {
						run(DIR_UP)
					} else if checkOrdersBelow(status.CurrentFloor) == true {
						run(DIR_DOWN)
					} else {
						run(DIR_NODIR)
					}
				}
			}
		}
		//Update status struct
		<-elevStatusChan
		elevStatusChan <- status

		abortFlag = CheckAbortFlag(abortChan)
	}
}

func driveInit(initChan chan int) {
	numFloors = ElevInit() //should be called from Comm. Handler?
	initChan <- numFloors
	ElevMotorDirection(0)
	status.CurrentFloor = ElevGetFloorSensorSignal()
	status.Direction = DIR_NODIR
	status.Running = false
	for floor := 0; floor < numFloors; floor++ {
		for dir := 0; dir < 3; dir++ {
			status.Orders[floor][dir] = false
		}

	}
}

func stopRoutine(executedOrdersChan chan OrderType, setLightsChan chan OrderType) {
	ElevMotorDirection(DIR_NODIR)
	status.Running = false
	ElevDoorOpenLight(true)
	status.DoorOpen = true

	//If idle and called by external button: set direction this way
	if status.Direction == DIR_NODIR {
		if status.Orders[status.CurrentFloor][DIR_UP] == true {
			status.Direction = DIR_UP
		} else if status.Orders[status.CurrentFloor][DIR_DOWN] == true {
			status.Direction = DIR_DOWN
		}
	}

	var order OrderType //Make clear order
	order.Floor = status.CurrentFloor
	order.Dir = status.Direction
	order.New = false

	setLightsChan <- order
	status.Orders[status.CurrentFloor][status.Direction] = false //Clear status.Orders from list

	//If external order: Report to master and clear internal order aswell
	if status.Direction != DIR_NODIR {
		executedOrdersChan <- order
		order.Dir = DIR_NODIR
		setLightsChan <- order
		status.Orders[status.CurrentFloor][DIR_NODIR] = false
	}

	time.Sleep(2 * time.Second) //Make smarter wait time??

	ElevDoorOpenLight(false)
	status.DoorOpen = false
}

func run(dir int) { //Feil så lenge elevMotorDirection bruker -1, 0 og 1
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
	for floor := currentFloor; floor < numFloors; floor++ {
		for dir := 0; dir < 3; dir++ {
			if status.Orders[floor][dir] == true {
				return true
			}
		}
	}
	return false
}

func checkOrdersBelow(currentFloor int) bool {
	for floor := 0; floor < currentFloor; floor++ {
		for dir := 0; dir < 3; dir++ {
			if status.Orders[floor][dir] == true {
				return true
			}
		}
	}
	return false
}
