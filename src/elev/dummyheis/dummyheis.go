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

	go makeOrders(buttonPressesChan)

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
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			if newLights[i][j] != oldLights[i][j] {
				fmt.Println(newLights)
				oldLights = newLights
				return
			}
		}
	}
}

func makeOrders(buttonPressesChan chan OrderType) {
	var order = OrderType{Floor: 0, Dir: 0, New: true}
	timer := time.NewTicker(1 * time.Second)
	for {
		<-timer.C
		order.Floor++
		if order.Floor == 3 {
			order.Floor = 0
		}
		buttonPressesChan <- order
	}

}
