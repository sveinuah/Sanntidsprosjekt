package elevdriver

import (
	. "../elevio"
	. "../typedef"
)

var lights [][]bool

func buttonInterface(clearLightChan chan OrderType, buttonPressesChan chan OrderType, initChan chan int) {
	numFloors = buttonInterfaceInit(initChan)

	abortFlag := false
	for abortFlag != true {
		//Get new button presses and send order up/to drive
		for floor := 0; floor < numFloors; floor++ {
			for dir := 0; dir < 3; dir++ {
				if elevGetButtonSignal(floor, dir) == true && lightList[floor][dir] == false { //Make variable lightList or read hardware each time?
					var order OrderType
					order.Floor = floor
					order.Dir = dir
					order.New = true
					buttonPressesChan <- order
					if order.Dir == DIR_NODIR {
						elevButtonLight(order.Floor, order.Dir, order.New)
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
			case order := <-clearLightChan:
				elevButtonLight(order.Floor, order.Dir, order.New)
				lights[Order.Floor][order.Dir] = order.New
			default:
				ordersInChannel = false
			}
		}
		abortFlag = checkAbortFlag(abortChan)
	}
}

func buttonInterfaceInit(initChan chan int) int {
	numFloors := <-initChan
	//wait for drive to run elevInit, return numFloors
	for floor := 0; floor < numFloors; floor++ {
		for dir := 0; dir < 3; dir++ {
			buttonList[floor][dir] = false
		}
	}
	return numFloors
}
