package elevdriver

import ."elev"

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

type Order struct {
	Floor int
	Direction int
}


var (
orderList [][] bool
status StatusType
previousFloor int
)

func drive(downChan chan int, upChan chan int, statusChan chan Status, numFloors int) {
	driveInit(numFloors)
	for {
		for() {//(anything in downChan)
			Order := <- downChan
			orderList[Order.Floor][Order.Direction] = true //Fetch orders and place in order list
		}
		//Check floor and see if should stop here or move on
		Status.currentFloor = elevGetFloorSensorSignal()
		if(Status.currentFloor == -1){
			//Do nothing
		} 
		else if(orderList[Status.currentFloor][Status.direction] == true || orderList[Status.currentFloor][DIR_NODIR] == true) { //order to stop here: stop here
			stopRoutine()
		} 
		else if(Status.running == false) {
			if(Status.direction == DIR_DOWN) {
				if(checkOrdersBelow(Status.currentFloor) == true) {
					run(DIR_DOWN)
				}
				else if(checkOrdersAbove(Status.currentFloor) == true ){
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
	for i := 0; i < numFloors; i++ {
		for j := 0; j < 3; j++ {
			orderList[i][j] = false
		}
		
	}
}

func stopRoutine() {
	elevMotorDirection(0)
	Status.running = false
	elevDoorOpenLight(1)
	Status.doorOpen = true

	orderList[Status.currentFloor][Status.direction] = false
	orderList[Status.currentFloor][DIR_NODIR] = false
	Order := {Status.currentFloor, Status.direction} //Ganske sikker på at dette er feil
	upChan <- Order

	//Timer = 2 sec - problem if this go routine is stuck here? mtp fetch orders

	elevDoorOpenLight(0)
	Status.doorOpen = false
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

func checkOrdersAbove(int floor) {
	for i := floor; i < numFloors; i++ {
		for j := 0; j < 3; j++ {
			if (orderList[i][j] == true) {
				return true
			}
		}
		
	}
	return false
}

func checkOrdersBelow(int floor) {
	for i := 0; i < floor; i++ {
		for j := 0; j < 3; j++ {
			if (orderList[i][j] == true) {
				return true
			}
		}
		
	}
	return false
}

func getDriveStatus(){

}