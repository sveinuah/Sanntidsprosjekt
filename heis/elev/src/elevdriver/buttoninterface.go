package elevdriver

import (
	. "elevio"
	"fmt"
	. "typedef"
)

var lights [4][3]bool

func ButtonInterface(quitChan chan bool, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, allocateOrdersChan chan OrderType, initChan chan bool) {
	for {
		//Get new button presses and send order up/to drive
		for floor := 0; floor < N_FLOORS; floor++ {
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
			for floor := 0; floor < N_FLOORS; floor++ {
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
		select {
		case <-quitChan:
			fmt.Println("BI Abort!!!")
			return
		default:
		}
	}
}

func buttonInterfaceInit(initChan chan int) {
	<-initChan
	//wait for drive to run elevInit, return N_FLOORS
	lights = [N_FLOORS][3]bool{{false}}
}
