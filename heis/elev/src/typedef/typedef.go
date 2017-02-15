package typedef

type OrderType struct {
	Floor int
	Dir   int
	New   bool
}

type StatusType struct {
	currentFloor int
	direction int
	running bool
	buttons [][]bool
	doorOpen bool
}

type ExtReport struct {
	internalStatus StatusType
	newOrders [] OrderType
}

type UnitType struct {
	Type int
	IP   string
	Port string
}

type OrderPackage struct {
	Order OrderType
	Unit  UnitType
}

type DataPackage struct {
	IP   string
	Port string
	Data []byte
}

type Error struct {
	ErrCode int
	ErrStr string
}

const (
	DIR_UP    = 0
	DIR_DOWN  = 1
	DIR_NODIR = 2
)

func checkAbortFlag(abortChan chan bool)
	abortFlag := <- abortChan
	abortChan <- abortFlag
	return abortFlag
}
