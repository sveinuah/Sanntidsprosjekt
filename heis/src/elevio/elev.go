package elevio

import (
	"log"
)

const N_BUTTONS = 3
const N_FLOORS = 4
const DEFAULT_MOTOR_SPEED = 2800
const MAX_SPEED = 4000

var motorSpeed = DEFAULT_MOTOR_SPEED
var initialized = false // for å passe på at kun én 

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

type ButtonFunction int

const (
	Up ButtonFunction = iota 
	Down
	Command
	Stop
	Obstruction
	Door
)

type Button struct {
	Floor int
	Type ButtonFunction
	Pushed bool
}

type Light struct {
	Floor 	int
	Type 	ButtonFunction
	On 		bool
}

/* Antar denne bød defineres et annet sted, ettersom den ikke brukes overhodet.
type Status struct {
	CurrentFloor int
	Dir int
	Running	bool
	Door bool
}
*/

func elevInit() {
	if initialized { //for at den ikke skal initialiseres fra flere steder
		return
	}
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
	initialized = true
}

func elevMotorDirection(dir int) {
	if dir == 0 {
		ioWriteAnalog(MOTOR, 0)
	}
	else if dir > 0 {
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

func elevSetLight(light Light) {
	switch light.Type {
		case Up: fallthrough
		case Down: fallthrough
		case Command: {
			if light.On {
				ioSetBit(lightMatrix[light.Floor][light.Type])
			}
			else {
				ioClearBit(lightMatrix[light.FLoor][light.type])
			}
		}
		case Stop: {
			if light.On {
				ioSetBit(LIGHT_STOP)
			}
			else {
				ioClearBit(LIGHT_STOP)
			}
		}
		case Obstruction: {
			if light.On {
				ioSetBit(OBSTRUCTION)
			}
			else {
				ioClearBit(OBSTRUCTION)
			}
		}
		case Door {
			if light.On {
				ioSetBit(LIGHT_DOOR_OPEN)
			}
			else {
				ioClearBit(LIGHT_DOOR_OPEN)
			}
		}
		default:
			{
				log.Fatal("Couldn't change ligth settings. Check if Floor is within range!")
			}
	}
	
}

func elevGetButtonSignal(button Button) Button {
	switch button.Type {
		case Up: fallthrough
		case Down: fallthrough
		case Command: {
			button.Pushed = ioReadBit(buttonMatrix[button.Floor,button.Type])
			return button
		}
		case Stop: {
			button.Pushed = ioReadBit(STOP)
			return button
		}
		case Obstruction: {
			button.Pushed = ioReadBit(OBSTRUCTION)
			return button
		}
		case Door {
			button.Pushed = ioReadBit(LIGHT_DOOR_OPEN)
			return button
		}
		default:
		{
			log.Fatal("Couldn't get button signal. Check if Floor is within range!")
		}
	}
}


func elevSetFloorIndicator(floor int) {
	if (floor < 0 || floor >= N_FLOORS) {
		log.Fatal("floor out of range")
	}

	if floor && 0x02 {
		ioSetBit(LIGHT_FLOOR_IND1)
	}
	else {
		ioClearBit(LIGHT_FLOOR_IND1)
	}

	if floor && 0x01 {
		ioSetBit(LIGHT_FLOOR_IND2)
	}
	else {
		ioClearBit(LIGHT_FLOOR_IND2)
	}
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