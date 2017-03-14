package elevdriver

import (
	. "elev/elevio"
	"fmt"
	"time"
	. "typedef"
)

const SLEEP_TIME = 100 * time.Millisecond

var lights [N_FLOORS][N_BUTTONS]bool

//ButtonInterface collects all button presses, sets and clear button lights.
func ButtonInterface(quitChan <-chan bool, extLightsChan <-chan [][]bool, setLightsChan <-chan OrderType, buttonPressesChan chan<- OrderType, allocateOrdersChan chan<- OrderType, initChan <-chan bool) {
	<-initChan //Wait for Drive to initialize Elevio before starting
	for {

		//Get new button presses and send external orders to Master and internal orders to Drive. For internal orders, button lights are set immediately.
		for floor := 0; floor < N_FLOORS; floor++ {
			for dir := 0; dir < 3; dir++ {
				if ElevGetButtonSignal(floor, dir) == true && lights[floor][dir] == false {
					var order OrderType
					order.Floor = floor
					order.Dir = dir
					order.New = true
					if order.Dir == DIR_NODIR {
						ElevButtonLight(order.Floor, order.Dir, order.New)
						allocateOrdersChan <- order
					} else {
						buttonPressesChan <- order
					}
				}
			}
		}

		//Copy external lights from master if updted. Set/clear only lights that have changed
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

		//Get and execute orders to set or clear single lights
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
			fmt.Println("Button Interface Abort!")
			return
		default:
		}
		time.Sleep(SLEEP_TIME)
	}
}
