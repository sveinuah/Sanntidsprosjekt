package main

import (
	."./typedef"
	"./elevio/elevio"
	"./localscheduler/localscheduler"
	"netinterface"
	"log"
	"time"
)

type Error struct {
	ErrCode int
	ErrStr string
}

type ExtReport struct {
	Err Error
	CurrentFloor int 
	Dir int 
	Running bool 
	newOrders [] Order
	orderMatrix [][] bool
}

func makeReport()


func main() {
	//Init - Look at init order?
	netwinint() //Feil navn
	
	executedOrdersChan		:= make(chan OrderType, 100)
	allocatedOrdersChan		:= make(chan OrderType, 100)
	driveStatusChan			:= make(chan StatusType)

	buttonPressesChan		:= make(chan OrderType, 100)
	extLightsChan			:= make(chan extLightsMatrix)

	extReportChan			:= make(chan ExtReport)

	//Pass all as pointers?
	go drive(allocatedOrdersChan, executedOrdersChan, driveStatusChan, numFloors)
	go buttonInterface(lightsChan, buttonPressesChan, numFloors)
	go networkinterface(allocatedOrdersChan, lightsChan, extReportChan)

	for {
		//get new button presses and executed orders. Add to report, send int orders to drive
		ordersInChannel := true
		for(ordersInChannel){
			select{
				case order := <-buttonPressesChan:
					//Add to report
					if(order.Dir == DIR_NODIR && order.Arg == true) {
						allocatedOrdersChan <- order
					}
				default:
					ordersInChannel = false
			}
		}

		if(beskjed om å sende status) {
			//send status
			//
		}
	}
}

