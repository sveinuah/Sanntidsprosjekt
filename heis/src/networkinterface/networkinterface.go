package networkinterface

import (
	"../networkmodule"
	"../typedef"
)

var OrderChan := make(chan, typedef.UnitType)

func SendOrder(directedOrder typedef.OrderPackage) {
	TransmitTCP()
}
