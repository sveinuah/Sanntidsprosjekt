package elevdriver

<<<<<<< HEAD
import ."elev"
=======
import ."elevio/elevio"
>>>>>>> 780ec1bffb1eae48d17203dde79adf13d6837432

var state int 
const (
	DIR_STOP = 0
	DIR_UP = 1
	DIR_DOWN = -1
)

type ButtonFunction int

const (
	Up ButtonFunction = iota 
	Down
	Command
	Stop
	Obstruction
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

type Status struct {
	CurrentFloor int
	Dir int
	Running	bool
	Door bool
}


func driverInit() {
	elevInit()
	elevMotorDirection(0)
<<<<<<< HEAD
	dire = idle
=======
	state = IDLE
}

func setLights(buttons) {

>>>>>>> 780ec1bffb1eae48d17203dde79adf13d6837432
}

func run() {
	for {
		if dir == 1 {
			elevMotorDirection(1)
		}
	}
}

func setIdle() {
	elevMotorDirection(0)
	state = idle
}

func getStatus(statusChan) {
	

}

func stopRoutine() {

}


