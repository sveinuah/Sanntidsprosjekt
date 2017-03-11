package test

import (
	"fmt"
	"typedef"
	"time"
)

var name string
var numFloors int

func Init(id string, sync chan [][]typedef.MasterOrder,  units chan typedef.UnitUpdate, quit chan bool) int {
	numFloors = 4
	name = id

	go makeUnits(units,quit)
	go makeSync(sync, quit)

	return numFloors

}

func Passive(sync chan [][]typedef.MasterOrder, units chan typedef.UnitUpdate, quit chan bool) {
	go makeUnits(units, quit)
	go makeSync(sync, quit)
}

func Active(units chan typedef.UnitUpdate, orderTx chan typedef.OrderType, orderRx chan typedef.OrderType, sync chan [][]typedef.MasterOrder, statusChan chan typedef.StatusType, statusReqChan chan int, lightChan chan [][]bool, quit chan bool) {
	wait := make(chan bool)

	go makeUnits(units,quit)
	go requestReport(statusReqChan, statusChan, quit)
	go pickupSync(sync, quit)
	go createOrders(wait, orderRx, quit)
	go receiveOrders(wait, orderTx, orderRx, quit)
}

func makeUnits(units chan<- typedef.UnitUpdate, quit <-chan bool) {
	fmt.Println("MakeUnits Starting!")
	tick := time.Tick(1000 * time.Millisecond)
	unit1 := typedef.UnitType{typedef.SLAVE, "Bob"}
	unit2 := typedef.UnitType{typedef.SLAVE, "Alice"}
	peers := []typedef.UnitType{unit1, unit2}
	update := typedef.UnitUpdate{Peers: peers, New: typedef.UnitType{}, Lost: []typedef.UnitType{}}

	for {
		select {
		case <- quit:
			fmt.Println("makeUnits Quitting!")
			return
		case <- tick:
			units <- update
		}
	}
}

func makeSync(sync chan [][]typedef.MasterOrder, quit <-chan bool) {
	fmt.Println("makeSync Starting!")
	var list = make([][]typedef.MasterOrder, numFloors, numFloors)
	for i := range(list) {
		list[i] = make([]typedef.MasterOrder, 2, 2)
	}

	tick := time.Tick(100 * time.Millisecond)
	for {
		select {
		case <- tick:
			sync <- list
		case <- quit:
			fmt.Println("makeSync Quitting")
			return
		}
	}
}

func requestReport(statusReqChan <-chan int, statusChan chan<- typedef.StatusType, quit chan bool) {
	for {
		select {
		case num := <- statusReqChan:
			time.Sleep(50 * time.Millisecond)

			rep1 := typedef.StatusType{"Bob", num, 2, typedef.DIR_UP, false, [][]bool{}, false}
			statusChan <- rep1
			rep2 := typedef.StatusType{"Alice", num, 3, typedef.DIR_DOWN, false, [][]bool{}, true}
			statusChan <- rep2
		case <- quit:
			fmt.Println("reqReports quitting!")
		}
	}
}

func pickupSync(sync <-chan [][]typedef.MasterOrder, quit <-chan bool) {
	for {
		select {
		case <- sync:
			fmt.Println("Got Sync")
		case <- quit:
			fmt.Println("pickupSync Quitting")
			return
		}
	}
}

func pickupLigths(lights <-chan [][]bool, quit <-chan bool) {
	for {
		select {
		case <- lights:
			fmt.Println("Got lights")
		case <- quit:
			fmt.Println("pickupLights Quitting")
			return
		}
	}
}

func createOrders(wait chan<- bool, orderRx chan<- typedef.OrderType, quit chan bool) {
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <- tick:
			fmt.Println("Made Order!")
			order := typedef.OrderType{"Carl", "Bob", 3, typedef.DIR_DOWN, true}
			orderRx <- order
			time.Sleep(50*time.Millisecond)
			wait <- true
		case <- quit:
			fmt.Println("Quitting createOrders!")
			return
		}
	}
}

func receiveOrders(wait <-chan bool, orderTx <-chan typedef.OrderType, orderRx chan<- typedef.OrderType, quit chan bool) {
	for {
		<- wait
		select {
		case order := <- orderTx:
			fmt.Println(order)
			order.From = order.To
			order.To = "Carl"
			order.New = false
			time.Sleep(100 * time.Millisecond)
			fmt.Println("Order Finished!")
			orderRx <- order
		case <- quit:
			fmt.Println("Quitting receiveOrders")
			return
		}
	}
}