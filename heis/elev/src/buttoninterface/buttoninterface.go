package elevdriver

import (
	. "../elevio"
	. "../typedef"
)

var lights [][]bool

func buttonInterface(nameforordersthatgotoBIChan chan OrderType, buttonPressesChan chan OrderType, numFloors int) {
	buttonInterfaceInit(numFloors)
	for {
		//Get new button presses and send order up/to drive
		for floor := 0; floor < numFloors; floor++ {
			for dir := 0; dir < 3; dir++ {
				if elevGetButtonSignal(floor, dir) == true && lightList[floor][dir] == false { //Make variable lightList or read hardware each time?
					var order OrderType
					order.Floor = floor
					order.Dir = dir
					order.Arg = true
					buttonPressesChan <- order
					if order.Dir == DIR_NODIR {
						elevButtonLight(order.Floor, order.Dir, order.Arg)
						allocatedOrdersChan <- order
					}
				}
			}
		}
		//Copy extLights from master if new in channel, set/clear lights that are wrong
		select {
		case extLights := <-extLightsChan:
			for floor := 0; floor < numFloors; floor++ {
				for dir := 0; dir < 2; dir++ {
					if lights[floor][dir] != extLights[floor][dir] {
						elevButtonLight(floor, dir, extLights[floor][dir])
						lights[floor][dir] = extLights[floor][dir]
					}
				}
			}
		default:
		}
		//Get new orders and set/clear lights
		ordersInChannel := true
		for ordersInChannel {
			select {
			case order := <-downChan:
				elevButtonLight(order.Floor, order.Dir, order.Arg)
				lights[Order.Floor][order.Dir] = order.Arg
			default:
				ordersInChannel = false
			}
		}
	}
}

func buttonInterfaceInit() int {
	//wait for drive to run elevInit, return numFloors
	for floor := 0; floor < numFloors; floor++ {
		for dir := 0; dir < 3; dir++ {
			buttonList[floor][dir] = false
		}
	}
	return numFloors
}
