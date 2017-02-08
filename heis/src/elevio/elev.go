package elevio

import (
	"log"
)

const N_BUTTONS = 3
const N_FLOORS = 4
const DEFAULT_MOTOR_SPEED = 2800
const MAX_SPEED = 4000

var motorSpeed = DEFAULT_MOTOR_SPEED

var lightMatrix [N_FLOORS][N_BUTTONS] int = {
	{LIGHT_UP1, LIGTH_DOWN1, LIGHT_COMMAND1},
	{LIGHT_UP2, LIGTH_DOWN2, LIGHT_COMMAND2},
	{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
	{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4}
}

var buttonMatrix [N_FLOORS][N_BUTTONS] int = {
	{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
	{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
    {BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
    {BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
}

func elevInit() {
	success := ioInit()
	if success == false {
		log.Fatal("Unable to initialize elevator hardware")
	}

	for floor := range(N_FLOORS) {
		for button := range(N_BUTTONS) {
			elevButtonLamp(floor, button, false)
		}
	}

	elevStopLight(false)
	elevDoorOpenLight(false)
	elevFloorIndicator(0)
}

func elevMotorDirection(dir int) {
	if dir == DIR_NODIR {
		ioWriteAnalog(MOTOR, 0)
	}
	else if dir == DIR_UP {
		ioClearBit(MOTORDIR)
		ioWriteAnalog(MOTOR,motorSpeed)
	}
	else {
		ioSetBit(MOTORDIR)
		ioWriteAnalog(MOTOR,motorSpeed)
	}
}

func changeMotorSpeed(speed int) int
{
	if(speed <= 0) {
		return motorSpeed
	}
	else if speed <= MAX_SPEED {
		motorSpeed = speed
		ioWriteAnalog(MOTOR,motorSpeed)
		return motorSpeed
	}
	else {
		motorSpeed = MAX_SPEED
		ioWriteAnalog(MOTOR,MAX_SPEED)
		return MAX_SPEED
	}
}

func elevButtonLight(floor int, button int, val bool) { 
	if (floor < 0 || button < 0 || floor >= N_FLOORS || button >= N_BUTTONS) {
		log.Fatal("floor/button out of range")
	}
	
	if val {
		ioSetBit(lightMatrix[floor][button])
	}
	else {
		ioClearBit(lightMatrix[floor][button])
	} 
}

func elevFloorIndicator(floor int) {
	if (floor < 0 || floor >= N_FLOORS) {
		log.Fatal("floor out of range")
	}

	if floor && 0x02 {
		ioSetBit(LIGHT_FLOOR_IND1)
	}
	else {
		ioClearBit(LIGHT_FLOOR_IND1)
	}

	if floor &&0x01 {
		ioSetBit(LIGHT_FLOOR_IND2)
	}
	else {
		ioClearBit(LIGHT_FLOOR_IND2)
	}
}

func elevDoorOpenLight(val bool) {
    if (value) {
        ioSetBit(LIGHT_DOOR_OPEN);
    } else {
        ioClearBit(LIGHT_DOOR_OPEN);
    }
}


func elevStopLight(val int) {
    if (value) {
        ioSetBit(LIGHT_STOP);
    } else {
        ioClearBit(LIGHT_STOP);
    }
}

func elevGetButtonSignal(floor int, button int) int {
	if (floor < 0 || button < 0 || floor >= N_FLOORS || button >= N_BUTTONS) {
		log.Fatal("floor/button out of range")
	}

	return ioReadBit(buttonMatrix[floor][button])
}

func elevGetFloorSensorSignal() int {
	if ioReadBit(SENSOR_FLOOR1) {
		return 0
	}
	else if ioReadBit(SENSOR_FLOOR2) {
		return 1
	}
	else if ioReadBit(SENSOR_FLOOR3) {
		return 2
	}
	else if ioReadBit(SENSOR_FLOOR4) {
		return 3
	}
	else {
		return -1
	}
}

func elevGetStopSignal() int {
	return ioReadBit(STOP)
}

func elevGetObstructionSignal() int {
	return ioReadBit(OBSTRUCTION)
}