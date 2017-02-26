package typedef

type OrderType struct {
	To    unitID
	From  unitID
	ID    int
	Floor int
	Dir   int
	New   bool
}

type UnitID string

type StatusType struct {
	From         unitID
	ID           int
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

type AckType struct {
	To   UnitID
	From UnitID
	Type string
	ID   int
}

type UnitType struct {
}

type ElevError struct {
	errCode int
	errStr  string
}

func (e *ElevError) Error() string {
	return e.errStr
}

func (e *ElevError) ErrorCode() int {
	return e.errCode
}

const (
	DIR_UP    int = 0
	DIR_DOWN  int = 1
	DIR_NODIR int = 2
)

func CheckAbortFlag(abortChan chan bool) bool {
	abortFlag := <-abortChan
	abortChan <- abortFlag
	return abortFlag
}
