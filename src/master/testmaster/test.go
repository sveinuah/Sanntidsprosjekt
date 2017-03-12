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

	go makeUnits(true, units,quit)
	go makeSync(sync, quit)

	return numFloors

}

func Passive(sync chan [][]typedef.MasterOrder, units chan typedef.UnitUpdate, quit chan bool) {
	go makeUnits(false, units, quit)
	go makeSync(sync, quit)
}

func Active(units chan typedef.UnitUpdate, orderTx chan typedef.OrderType, orderRx chan typedef.OrderType, sync chan [][]typedef.MasterOrder, statusChan chan typedef.StatusType, statusReqChan chan int, lightChan chan [][]bool, quit chan bool) {
	wait := make(chan bool)

	go makeUnits(true, units,quit)
	go requestReport(statusReqChan, statusChan, quit)
	go pickupSync(sync, quit)
	go createOrders(wait, orderRx, orderTx, quit)
	go pickupLigths(lightChan, quit)
}

func makeUnits(active bool, units chan<- typedef.UnitUpdate, quit <-chan bool) {
	fmt.Println("MakeUnits Starting!")
	tick := time.Tick(1000 * time.Millisecond)
	unit1 := typedef.UnitType{typedef.SLAVE, "Bob"}
	unit2 := typedef.UnitType{typedef.SLAVE, "Alice"}
	peers := []typedef.UnitType{unit1, unit2}
	if !active {
		peers = append(peers, typedef.UnitType{typedef.MASTER, "Bjarne"})
	}
	update := typedef.UnitUpdate{Peers: peers, New: typedef.UnitType{}, Lost: []typedef.UnitType{}}
	lateAdd := time.After(4* time.Second)

	for {
		select {
		case <- quit:
			fmt.Println("makeUnits Quitting!")
			return
		case <- lateAdd:
			if active {
				peers = append(peers, typedef.UnitType{typedef.MASTER, "Bjarne"})
				update = typedef.UnitUpdate{Peers: peers, New: typedef.UnitType{}, Lost: []typedef.UnitType{}}
			} else {
				length := len(peers)
				peers = peers[:length-1]
				update = typedef.UnitUpdate{Peers: peers, New: typedef.UnitType{}, Lost: []typedef.UnitType{}}
			}
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

			orders1 := make([][]bool, numFloors, numFloors)
			for i := range(orders1) {
				orders1[i] = make([]bool,3,3)
			}
			rep1 := typedef.StatusType{"Bob", num, 2, typedef.DIR_UP, false, orders1, false}
			statusChan <- rep1

			orders2 := make([][]bool, numFloors, numFloors)
			for i := range(orders2) {
				orders2[i] = make([]bool,3,3)
			}
			rep2 := typedef.StatusType{"Alice", num, 3, typedef.DIR_DOWN, false, orders2, true}
			statusChan <- rep2
		case <- quit:
			fmt.Println("reqReports quitting!")
			return
		}
	}
}

func pickupSync(sync <-chan [][]typedef.MasterOrder, quit <-chan bool) {
	for {
		select {
		case <- sync:
			//fmt.Println("Got Sync")
		case <- quit:
			fmt.Println("pickupSync Quitting")
			return
		}
	}
}

func pickupLigths(lights <-chan [][]bool, quit <-chan bool) {
	fmt.Println("Starting pickupLigths")
	for {
		select {
		case light := <- lights:
			fmt.Println("Got lights", light)
		case <- quit:
			fmt.Println("pickupLights Quitting")
			return
		}
	}
}

func createOrders(wait chan<- bool, orderRx chan<- typedef.OrderType, orderTx <-chan typedef.OrderType, quit chan bool) {
	tick := time.Tick(2 * time.Second)
	t := time.NewTimer(100*time.Millisecond)
	var order typedef.OrderType
	t.Stop()
	for {
		select {
		case <- tick:
			order := typedef.OrderType{"Carl", "Bob", 3, typedef.DIR_DOWN, true}
			fmt.Println("Made Order!", order)
			orderRx <- order
			time.Sleep(50*time.Millisecond)
		case order = <- orderTx:
			fmt.Println("Sent order:", order)
			t.Reset(100*time.Millisecond)
		case <- t.C:
			order.From = order.To
			order.To = "Carl"
			order.New = false
			fmt.Println("Order Finished!")
			orderRx <- order
		case <- quit:
			fmt.Println("Quitting create/handleOrders!")
			return
		}
	}
}

func receiveOrders(wait <-chan bool, orderTx <-chan typedef.OrderType, orderRx chan<- typedef.OrderType, quit chan bool) {
	for {
		<- wait
		select {
		
		case <- quit:
			fmt.Println("Quitting receiveOrders")
			return
		}
	}
}