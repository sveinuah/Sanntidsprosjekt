package elevdriver

import (
	. "elevio"
	. "typedef"
)

var lights [][]bool

func ButtonInterface(abortChan chan bool, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, allocateOrdersChan chan OrderType, initChan chan int) {
	numFloors := buttonInterfaceInit(initChan)

	abortFlag := false
	for abortFlag != true {
		//Get new button presses and send order up/to drive
		for floor := 0; floor < numFloors; floor++ {
			for dir := 0; dir < 3; dir++ {
				if ElevGetButtonSignal(floor, dir) == true && lights[floor][dir] == false { //Make variable lightList or read hardware each time?
					var order OrderType
					order.Floor = floor
					order.Dir = dir
					order.New = true
					buttonPressesChan <- order
					if order.Dir == DIR_NODIR {
						ElevButtonLight(order.Floor, order.Dir, order.New)
						allocateOrdersChan <- order
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
						ElevButtonLight(floor, dir, extLights[floor][dir])
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
			case order := <-setLightsChan:
				ElevButtonLight(order.Floor, order.Dir, order.New)
				lights[order.Floor][order.Dir] = order.New
			default:
				ordersInChannel = false
			}
		}
		abortFlag = CheckAbortFlag(abortChan)
	}
}

func buttonInterfaceInit(initChan chan int) {
	numFloors := <-initChan
	//wait for drive to run elevInit, return numFloors
	lights = [N_FLOORS][3]bool{{false}}
	return numFloors
}
