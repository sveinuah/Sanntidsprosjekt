package typedef

type OrderType struct {
	Floor     int
	Direction int
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
