package networkinterface

import (
	"../networkmodule"
	. "../typedef"
)

var OrderChan := make(chan, UnitType)

func NetworkInit() {

	OrderChan := make(chan, OrderType)
	TargetChan := make(chan, UnitType)

	go networkmodule.TransmitTCP(TargetChan, OrderChan)
	go networkmodule.ReceiveTCP(OrderChan)
	
	go networkmodule.TransmitTCP(TargetChan, StatusChan)

	ẗarget := UnitType{0,"129.241.187.43","34933"}

	go func() {
		order := OrderType{0,0};
		for{
			order.Floor++
			if order.Dir != 0 {
				order.Dir = 0
			} else order.Dir = 1
			TargetChan <- target
			OrderChan <- order
		}
	}
}


