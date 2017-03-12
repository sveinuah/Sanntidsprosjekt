package dummyheis

import (
	"fmt"
	"time"
	. "typedef"
)

var (
	oldLights [][]bool
	status    StatusType
)

func DummyHeis(quitChan chan bool, allocateOrdersChan chan OrderType, executedOrdersChan chan OrderType, extLightsChan chan [][]bool, setLightsChan chan OrderType, buttonPressesChan chan OrderType, elevStatusChan chan StatusType) {
	status.CurrentFloor = 3
	status.Direction = 0
	status.Running = true
	status.MyOrders = make([][]bool, 4, 4)
	for i := range status.MyOrders {
		status.MyOrders[i] = make([]bool, 3, 3)
	}

	status.DoorOpen = false
	for {
		select {
		case newOrder := <-allocateOrdersChan:
			go handleOrder(newOrder, executedOrdersChan)
		case newLights := <-extLightsChan:
			getLights(newLights)
		default:
			select {
			case oldStatus := <-elevStatusChan:
				status.DoorOpen = !oldStatus.DoorOpen
				elevStatusChan <- status
			default:
				elevStatusChan <- status
			}
		}
	}
}

func handleOrder(order OrderType, executedOrdersChan chan OrderType) {
	timer := time.After(4 * time.Second)
	<-timer
	order.New = false
	executedOrdersChan <- order
}

func getLights(newLights [][]bool) {
	if oldLights[1][1] != newLights[1][1] {
		fmt.Println(newLights)
	}
}
