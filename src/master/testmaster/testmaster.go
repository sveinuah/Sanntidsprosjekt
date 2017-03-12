package main

import (
	"fmt"
	"master/testMNI"
	"time"
	. "typedef"
)

func main() {
	fmt.Println("Running tests!")

	reportNum := 0
	statusReqChan := make(chan int, 1)
	statusChan := make(chan StatusType, 10)
	quit := make(chan bool)

	testMNI.Init_tmni(statusReqChan, statusChan, quit)

	for {
		time.Sleep(1 * time.Second)
		reportNum++
		statusReqChan <- reportNum
		time.Sleep(300 * time.Millisecond)
		for len(statusChan) > 0 {
			fmt.Println("Got report:", <-statusChan)
		}
		if reportNum > 142 {
			close(quit)
			time.Sleep(1 * time.Second)
			return
		}
	}
}
