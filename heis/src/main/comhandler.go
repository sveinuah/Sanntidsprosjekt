package main

import (
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
	newOrders [] Button
	completedOrders [] Button
}

func makeReport()


func main() {
	

	go drive()
}