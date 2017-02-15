package networkinterface

import (
	"../networkmodule"
	"../typedef"
	"fmt"
	"strconv"
	"time"
)

func main() {

	dataRxChan := make(chan typedef.DataPackage, 3)
	dataTxChan := make(chan typedef.DataPackage, 3)
	port := ":20014"

	go UdpNI.UDPListenAndReceive(port, dataRxChan)
	go UdpNI.UDPTransmit(dataTxChan)

	var rPackage typedef.DataPackage

	go func() {
		for {
			select {
			case rPackage = <-dataRxChan:
				fmt.Println(string(rPackage.Data) + " From: " + rPackage.IP + ":" + rPackage.Port)
			default:
				fmt.Println("Nothing to report..")
			}
			time.Sleep(1 * time.Second)
		}
	}()

	var tPackage typedef.DataPackage
	var iter int
	var str string
	tPackage.IP = "127.0.0.1"
	tPackage.Port = "12000"

	for {

		iter++
		str = "Test UDP-Transmitter, Iteration: " + strconv.Itoa(iter)

		tPackage.Data = []byte(str)
		dataTxChan <- tPackage
		time.Sleep(1 * time.Second)
	}
}
