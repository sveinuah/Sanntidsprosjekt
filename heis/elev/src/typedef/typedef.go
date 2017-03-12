package typedef

type OrderType struct {
	To    string
	From  string
	Floor int
	Dir   int
	New   bool
}

type unitID string

type StatusType struct {
	ID           unitID
	CurrentFloor int
	Direction    int
	Running      bool
	Orders       [4][3]bool //floor, dir
	DoorOpen     bool
}

type ExtReport struct {
	internalStatus StatusType
	newOrders      []OrderType
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

type Error struct {
	ErrCode int
	ErrStr  string
}

const (
	DIR_UP    int = 0
	DIR_DOWN  int = 1
	DIR_NODIR int = 2
)

func CheckAbortFlag(abortChan chan bool) bool { //Hva hvis flagget blir borte??
	abortFlag := <-abortChan
	abortChan <- abortFlag
	return abortFlag
}
