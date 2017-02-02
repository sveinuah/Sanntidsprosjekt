package elevdriver

import(
."../elevio"
)

const(
DIR_UP = 0
DIR_DOWN = 1
DIR_NODIR = 2
)

type StatusType struct {
	currentFloor int
	direction int
	running bool
	buttons bool [][]
	doorOpen bool
}

type OrderType struct {
	Floor int
	Direction int
}


var (
orderList [][] bool //rows = floors, columns = button number
status StatusType
previousFloor int
)

func drive(downChan chan int, upChan chan int, statusChan chan status, numFloors int) {
	driveInit(numFloors)
	for {
		//Get all orders from downChan and place in orderList
		ordersInChannel := true
		for(ordersInChannel){
			select{
				case Order := <- downChan:
					orderList[Order.Floor][Order.Direction] = true
				default:
					ordersInChannel = false
				}
		}
		//Check floor and see if elev should stop here, move oin either direction or stay put
		status.currentFloor = elevGetFloorSensorSignal()
		if(status.currentFloor == -1){
			if(status.running == false) {
				run(up)
			}
			//Between floors - make this run up/down if not running to handle restarts?
		} 
		else if(orderList[status.currentFloor][status.direction] == true || orderList[status.currentFloor][DIR_NODIR] == true) { //order to stop here: stop here
			stopRoutine()
		} 
		else if(status.running == false) {
			if(status.direction == DIR_DOWN) {
				if(checkOrdersBelow(status.currentFloor) == true) {
					run(DIR_DOWN)
				}
				else if(checkOrdersAbove(status.currentFloor) == true ){
					run(DIR_UP)
				}
				else {
					run(DIR_NODIR)
				}
			} 
			else {
				if(checkOrdersAbove() == true) {
					run(DIR_UP)
				}
				else if(checkOrdersBelow() == true) {
					run(DIR_DOWN)
				}
				else {
					run(DIR_NODIR)
				}
			}
		}
	}
}



func driveInit(numFloors int) { //add find floor sequence
	elevInit()
	elevMotorDirection(0)
	status.running = false
	status.
	status.direction = DIR_NODIR
	for floor := 0; floor < numFloors; i++ {
		for dir := 0; dir < 3; j++ {
			orderList[floor][dir] = false
		}
		
	}
}

func stopRoutine() {
	elevMotorDirection(0)
	status.running = false
	elevDoorOpenLight(1)
	status.doorOpen = true

	orderList[status.currentFloor][status.direction] = false
	orderList[status.currentFloor][DIR_NODIR] = false

	var Order OrderType
	Order.Floor = status.currentFloor
	Order.Direction = status.direction
	upChan <- Order

	//Timer = 2 sec - problem if this go routine is stuck here? mtp fetch orders

	elevDoorOpenLight(0)
	status.doorOpen = false
}

func run(dir int) {
	elevMotorDirection(dir)
	status.direction = dir
	if(dir == DIR_NODIR) {
		status.running = false
	}
	else {
		status.running = true
	}
}

func checkOrdersAbove(int currentFloor) {
	for floor := currentFloor; floor < numFloors; floor++ {
		for button := 0; button < 3; button++ {
			if (orderList[floor][button] == true) {
				return true
			}
		}
	}
	return false
}

func checkOrdersBelow(int currentFloor) {
	for floor := 0; floor < currentFloor; floor++ {
		for button := 0; button < 3; button++ {
			if(orderList[floor][button] == true) {
				return true
			}
		}
	}
	return false
}

func GetDrivestatus(){
	return status
}