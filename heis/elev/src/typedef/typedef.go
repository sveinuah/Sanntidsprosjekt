package typedef

type OrderType struct {
	Floor int
	Dir   int
	Arg   bool
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

const (
	DIR_UP    = 0
	DIR_DOWN  = 1
	DIR_NODIR = 2
)

type DataPackage struct {
	IP   string
	Port string
	Data []byte
}
