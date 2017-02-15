package elevio

import (
	"log"
	. "typedef"
)

const N_BUTTONS int = 3
const N_FLOORS int = 4
const N_STATUS_BUTTONS int = 3
const DEFAULT_MOTOR_SPEED int = 2800
const MAX_SPEED int = 4000

var motorSpeed = DEFAULT_MOTOR_SPEED

var buttonLightMatrix [N_FLOORS][N_BUTTONS]int

var buttonMatrix [N_FLOORS][N_BUTTONS]int

func ElevInit() int {
	buttonLightMatrix = [N_FLOORS][N_BUTTONS]int{
		{LIGHT_UP1, LIGHT_DOWN1, LIGHT_COMMAND1},
		{LIGHT_UP2, LIGHT_DOWN2, LIGHT_COMMAND2},
		{LIGHT_UP3, LIGHT_DOWN3, LIGHT_COMMAND3},
		{LIGHT_UP4, LIGHT_DOWN4, LIGHT_COMMAND4},
	}

	buttonMatrix = [N_FLOORS][N_BUTTONS]int{
		{BUTTON_UP1, BUTTON_DOWN1, BUTTON_COMMAND1},
		{BUTTON_UP2, BUTTON_DOWN2, BUTTON_COMMAND2},
		{BUTTON_UP3, BUTTON_DOWN3, BUTTON_COMMAND3},
		{BUTTON_UP4, BUTTON_DOWN4, BUTTON_COMMAND4},
	}

	success := IoInit()
	if success == false {
		log.Fatal("Unable to initialize elevator hardware")
	}

	for floor := 0; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			ElevButtonLight(floor, button, false)
		}
	}

	ElevStopLight(false)
	ElevDoorOpenLight(false)
	ElevFloorIndicator(0)
	return N_FLOORS
}

func ElevMotorDirection(dir int) {
	if dir == DIR_NODIR {
		IoWriteAnalog(MOTOR, 0)
	} else if dir == DIR_UP {
		IoClearBit(MOTORDIR)
		IoWriteAnalog(MOTOR, motorSpeed)
	} else {
		IoSetBit(MOTORDIR)
		IoWriteAnalog(MOTOR, motorSpeed)
	}
}

func ChangeMotorSpeed(speed int) int {
	if speed <= 0 {
		return motorSpeed
	} else if speed <= MAX_SPEED {
		motorSpeed = speed
		IoWriteAnalog(MOTOR, motorSpeed)
		return motorSpeed
	} else {
		motorSpeed = MAX_SPEED
		IoWriteAnalog(MOTOR, MAX_SPEED)
		return MAX_SPEED
	}
}

func ElevButtonLight(floor int, button int, val bool) {
	if floor < 0 || button < 0 || floor >= N_FLOORS || button >= N_BUTTONS {
		log.Fatal("floor/button out of range")
	}

	if val {
		IoSetBit(buttonLightMatrix[floor][button])
	} else {
		IoClearBit(buttonLightMatrix[floor][button])
	}
}

func ElevFloorIndicator(floor int) {
	if floor < 0 || floor >= N_FLOORS {
		log.Fatal("floor out of range")
	}

	if (floor & 0x02) != 0 {
		IoSetBit(LIGHT_FLOOR_IND1)
	} else {
		IoClearBit(LIGHT_FLOOR_IND1)
	}

	if (floor & 0x01) != 0 {
		IoSetBit(LIGHT_FLOOR_IND2)
	} else {
		IoClearBit(LIGHT_FLOOR_IND2)
	}
}

func ElevDoorOpenLight(val bool) {
	if val {
		IoSetBit(LIGHT_DOOR_OPEN)
	} else {
		IoClearBit(LIGHT_DOOR_OPEN)
	}
}

func ElevStopLight(val bool) {
	if val {
		IoSetBit(LIGHT_STOP)
	} else {
		IoClearBit(LIGHT_STOP)
	}
}

func ElevGetButtonSignal(floor int, button int) bool {
	if floor < 0 || button < 0 || floor >= N_FLOORS || button >= N_BUTTONS {
		log.Fatal("floor/button out of range")
	}
	return IoReadBit(buttonMatrix[floor][button])
}

func ElevGetFloorSensorSignal() int {
	if IoReadBit(SENSOR_FLOOR1) {
		return 0
	} else if IoReadBit(SENSOR_FLOOR2) {
		return 1
	} else if IoReadBit(SENSOR_FLOOR3) {
		return 2
	} else if IoReadBit(SENSOR_FLOOR4) {
		return 3
	} else {
		return -1
	}
}

func ElevGetStopSignal() bool {
	return IoReadBit(STOP)
}

func ElevGetObstructionSignal() bool {
	return IoReadBit(OBSTRUCTION)
}
