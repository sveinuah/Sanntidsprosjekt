package elevdriver

import (
	. "../elevio"
	. "../typedef"
)

var extLights [][] bool
var intLights [] bool

//or: 

//Should this module be able to receive buttonList type arguments?
func buttonInterface(downChan chan OrderType, upChan chan OrderType, numFloors int) {
	buttonInterfaceInit(numFloors)
	for {
		//Get new button presses and send order up
		for floor := 0; floor < numFloors; i++ {
			for dir := 0; dir < 3; j++ {
				if elevGetButtonSignal(floor, dir) == true && lightList[floor][dir] == false { //Make variable lightList or read hardware each time?
					var order OrderType
					order.Floor = floor
					order.Dir = dir
					order.Arg = true
					upChan <- order
					if order.Dir == DIR_NODIR {
						elevButtonLight(order.Floor, order.Dir, order.Arg)
					}
				}
			}
		}
		//Copy extLights from master if new in channel
		select {
			case TempExtLights := <- extLightsChan:
				if()

			default:
		}				
		//Get new orders and set/clear lights
		ordersInChannel := true
		for ordersInChannel {
			select {
			case order := <-downChan:
				elevButtonLight(order.Floor, order.Dir, order.Arg)
				buttonList[Order.Floor][order.Dir] = order.Arg
			default:
				ordersInChannel = false
			}
		}
	}
}

func buttonInterfaceInit(numFloors int) {
	for floor := 0; floor < numFloors; i++ {
		for dir := 0; dir < 3; j++ {
			buttonList[floor][dir] = false
		}
	}
}
