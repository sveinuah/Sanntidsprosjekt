package elevdriver

import (
	. "elev/elevio"
	"fmt"
	"time"
	. "typedef"
)

const DIR_INTERNAL = 2

var lights [N_FLOORS][N_BUTTONS]bool
var buttonSent [N_FLOORS][N_BUTTONS]bool

func ButtonInterface(quitChan chan bool, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, allocateOrdersChan chan OrderType, initChan chan bool) {
	for {
		//Get new button presses and send order up/to drive
		for floor := 0; floor < N_FLOORS; floor++ {
			for dir := 0; dir < 3; dir++ {
				if ElevGetButtonSignal(floor, dir) == true && lights[floor][dir] == false && buttonSent[floor][dir] == false {
					var order OrderType
					order.Floor = floor
					order.Dir = dir
					order.New = true
					if order.Dir == DIR_INTERNAL {
						ElevButtonLight(order.Floor, order.Dir, order.New)
						allocateOrdersChan <- order
					} else {
						buttonPressesChan <- order
					}
					buttonSent[floor][dir] = true
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
						buttonSent[floor][dir] = false
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
				buttonSent[order.Floor][order.Dir] = false
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
		time.Sleep(200 * time.Millisecond)
	}
}

func buttonInterfaceInit(initChan chan int) {
	<-initChan
	//wait for drive to run elevInit, return N_FLOORS
	lights = [N_FLOORS][3]bool{{false}}
}
