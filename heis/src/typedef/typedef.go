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
<<<<<<< HEAD

type OrderPackage struct {
	Order OrderType
	Unit  UnitType
}

const (
	DIR_UP    = 0
	DIR_DOWN  = 1
	DIR_NODIR = 2
)
=======
>>>>>>> 915f7ffd456e7a492b4d83767c964dd5cf806f7d
