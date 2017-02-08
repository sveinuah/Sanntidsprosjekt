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
