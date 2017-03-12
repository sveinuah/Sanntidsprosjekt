package main

import(
	"fmt"
	"time"
	"master/testMNI"
	. "typedef"
)

func main() {
	fmt.Println("Running tests!")

	reportNum := 0
	statusReqChan := make(chan int, 1)
	statusChan := make(chan StatusType, 10)
	quit := make(chan bool)

	testMNI.Init(statusReqChan, statusChan, quit)

	for {
		time.Sleep(1 * time.Second)
		reportNum++
		statusReqChan <- reportNum
		time.Sleep(300 * time.Millisecond)
		for len(statusChan) > 0 {
			fmt.Println("Got report:", <- statusChan)
		}
		if reportNum > 14 {
			close(quit)
			time.Sleep(1* time.Second)
			return
		}
	}
}