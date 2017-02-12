package elevdriver

import(
."../typedef"
."../elevio"
"time"
)

var (
orderList [][] bool //rows = floors, columns = direction 
status StatusType
numFloors int
)

func drive(allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, setLightsChan chan OrderType, elevStatusChan chan StatusType, intChan chan StatusType) {
	driveInit()
	abortFlag := false
	for abortFlag != true{
		//Get all orders from allocateOrdersChan and place in orderList
		ordersInChannel := true
		for(ordersInChannel){
			select{
				case order := <- allocateOrdersChan:
					orderList[Order.Floor][order.Dir] = order.New
				default:
					ordersInChannel = false
				}
		}
		//Check floor and see if elev should stop here, move in either direction or stay put
		tempFloor := elevGetFloorSensorSignal()
		if(tempFloor == -1){
			if(status.running == false){
				run(up)
			}
		} else {
			if(status.currentFloor != tempFloor) {
				elevFloorIndicator(tempFloor)
				status.currentFloor = tempFloor
			}
			if(orderList[status.currentFloor][status.direction] == true || orderList[status.currentFloor][DIR_NODIR] == true) { //order to stop here: stop here
				stopRoutine()
			} else {
				if(status.direction == DIR_DOWN) {
					if(checkOrdersBelow(status.currentFloor) == true) {
						run(DIR_DOWN)
					} else if(checkOrdersAbove(status.currentFloor) == true ){
						run(DIR_UP)
					} else {
						run(DIR_NODIR)
					}
				} else {
					if(checkOrdersAbove(status.currentFloor) == true) {
						run(DIR_UP)
					} else if(checkOrdersBelow(status.currentFloor) == true) {
						run(DIR_DOWN)
					} else {
						run(DIR_NODIR)
					}
				}
			}
		}
		//Update status struct
		<- intStatusChan
		intStatusChan <- status
		
		abortFlag = checkAbortFlag(abortChan)
	}
}



func driveInit(numFloors int) {
	numFloors = elevInit() should be called from Comm. Handler?
	initChan <- numFloors
	elevMotorDirection(0)
	status.running = false
	status.currentFloor = elevGetFloorSensorSignal()
	status.direction = DIR_NODIR
	for floor := 0; floor < numFloors; i++ {
		for dir := 0; dir < 3; j++ {
			orderList[floor][dir] = false
		}
		
	}
}

func stopRoutine() { 
	elevMotorDirection(false)
	status.running = false
	elevDoorOpenLight(true)
	status.doorOpen = true

	//If idle and called by external button: set direction this way
	if(status.direction == DIR_NODIR){
		if(orderList[status.currentFloor][DIR_UP] == true){
			status.direction = DIR_UP
		} else if(orderList[status.currentFloor][DIR_DOWN] == true){
			status.direction = DIR_DOWN
		}
	}

	var order OrderType //Make clear order
	order.Floor = status.currentFloor
	order.Dir = status.direction
	order.New = false

	setLightChan <- order
	orderList[status.currentFloor][status.direction] = false //Clear orders from list

	//If external order: Report to master and clear internal order aswell
	if(status.direction != DIR_NODIR){
		executedOrdersChan <- order
		order.Dir = DIR_NODIR
		setLightChan <- order 
		orderList[status.currentFloor][DIR_NODIR] = false
	}

	time.Sleep(2*time.Second) //Make smarter wait time??

	elevDoorOpenLight(false)
	status.doorOpen = false
}

func run(dir int) { //Feil så lenge elevMotorDirection bruker -1, 0 og 1
	switch dir{
	case DIR_NODIR:
		if(status.running == true){
			elevMotorDirection(0)
			status.running = false
		}
	default:
		if (status.direction != dir || status.running == false) {
		elevMotorDirection(dir)
		status.running = true
		}
	}
	status.direction = dir
}

func checkOrdersAbove(int currentFloor) {
	for floor := currentFloor; floor < numFloors; floor++ {
		for dir := 0; dir < 3; dir++ {
			if (orderList[floor][dir] == true) {
				return true
			}
		}
	}
	return false
}

func checkOrdersBelow(int currentFloor) {
	for floor := 0; floor < currentFloor; floor++ {
		for dir := 0; dir < 3; dir++ {
			if(orderList[floor][dir] == true) {
				return true
			}
		}
	}
	return false
}