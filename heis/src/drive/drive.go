package elevdriver

import(
."../elevio"
"time"
)



type StatusType struct {
	currentFloor int
	direction int
	running bool
	buttons bool [][]
	doorOpen bool
}

var (
orderList [][] bool //rows = floors, columns = direction 
status StatusType
previousFloor int
)

func drive(downChan chan OrderType, upChan chan OrderType, statusChan chan StatusType, numFloors int) {
	driveInit(numFloors)
	for {
		//Get all orders from downChan and place in orderList
		ordersInChannel := true
		for(ordersInChannel){
			select{
				case order := <- downChan:
					orderList[Order.Floor][order.Dir] = order.Arg
				default:
					ordersInChannel = false
				}
		}
		//Check floor and see if elev should stop here, move in either direction or stay put
		tempFloor := elevGetFloorSensorSignal()
		if(tempFloor == -1 && status.running == false){
			run(up)
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
	}
}



func driveInit(numFloors int) { //add find floor sequence
	//elevInit() should be called from Comm. Handler?
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

func stopRoutine() { //Make smarter? As is: stop 2 sec, if anyone presses button: 2 sec more (but not 2 sec x buttonpresses)
	elevMotorDirection(0)
	status.running = false
	elevDoorOpenLight(1)
	status.doorOpen = true

	orderList[status.currentFloor][status.direction] = false
	orderList[status.currentFloor][DIR_NODIR] = false

	var order OrderType
	order.Floor = status.currentFloor
	order.Dir = status.direction
	upChan <- order

	time.Sleep(2*time.Second)

	elevDoorOpenLight(0)
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

func GetDrivestatus(){ //Make status exclusivly acessible? Or pass status up through main drive goroutine?
	return status
}