package typedef

import "time"

// OrderType haha test lol.
type OrderType struct {
	To    UnitID
	From  UnitID
	Floor int
	Dir   int
	New   bool
}

// MasterOrder skal.
type MasterOrder struct {
	o         OrderType
	delegated time.Time
	estimated time.Time
}

// UnitUpdate skal.
type UnitUpdate struct {
	Peers []UnitType
	New   UnitType
	Lost  []UnitType
}

// UnitID wow!
type UnitID string

// StatusType o.O
type StatusType struct {
	From         UnitID
	ID           int
	CurrentFloor int
	Direction    int
	Running      bool
	MyOrders     [][]bool //floor, dir
	DoorOpen     bool
}

// UnitType shiiiet!
type UnitType struct {
	Type int
	ID   UnitID
}

// AckType .You say AckType?
type AckType struct { //Jeg vil at du skal ha denne i nettwork Interface --Schwung
	To   UnitID
	From UnitID
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
	// DirUp defines up direction
	DirUp int = 0
	// DirDown defines up direction
	DirDown int = 1
	// DirNodir defines up direction
	DirNodir int = 2
	// MasterType defines the master type
	MasterType string = "1"
	// SlaveType defines the slave type
	SlaveType string = "2"
)
