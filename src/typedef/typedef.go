package typedef

type OrderType struct {
	To    UnitID
	From  UnitID
	Floor int
	Dir   int
	New   bool
}

type UnitUpdate struct {
	Peers []UnitType
	New   UnitType
	Lost  []UnitType
}

type UnitID string

type StatusType struct {
	From         UnitID
	ID           int
	CurrentFloor int
	Direction    int
	Running      bool
	MyOrders     [][]bool //floor, dir
	DoorOpen     bool
}

type UnitType struct {
	Type int
	ID   UnitID
}

type AckType struct { //Jeg vil at du skal ha denne i nettwork Interface --Schwung
	To   UnitID
	From UnitID
	Type string
	ID   int
}

type ElevError struct {
	errCode int
	errStr  string
}

func (e ElevError) Error() string  { return e.errStr }
func (e ElevError) ErrorCode() int { return e.errCode }

const (
	DIR_UP      int    = 0
	DIR_DOWN    int    = 1
	DIR_NODIR   int    = 2
	MASTER      string = "1"
	SLAVE       string = "2"
	TYPE_MASTER int    = 1
	TYPE_SLAVE  int    = 2
)
