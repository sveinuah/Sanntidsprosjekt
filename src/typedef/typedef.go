package typedef

import "time"

// OrderType haha test lol.
type OrderType struct {
	To    string
	From  string
	Floor int
	Dir   int
	New   bool
}

// MasterOrder skal.
type MasterOrder struct {
	Order     OrderType
	Estimated time.Time
}

// UnitUpdate skal.
type UnitUpdate struct {
	Peers []UnitType
	New   UnitType
	Lost  []UnitType
}

// StatusType o.O
type StatusType struct {
	From         string
	ID           int
	CurrentFloor int
	Direction    int
	Running      bool
	MyOrders     [][]bool //floor, dir
	DoorOpen     bool
}

// UnitType shiiiet!
type UnitType struct {
	Type string
	ID   string
}

// AckType .You say AckType?
type AckType struct { //Jeg vil at du skal ha denne i nettwork Interface --Schwung
	To   string
	From string
	Type string
	ID   int
}

// ElevError is cool!
type ElevError struct {
	errCode int
	errStr  string
}

func (e ElevError) Error() string { return e.errStr }

// ErrorCode must
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
