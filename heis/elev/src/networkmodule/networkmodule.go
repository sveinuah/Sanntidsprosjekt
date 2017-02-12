package networkmodule

import (
	. "../../typedef"
	"log"
	"net"
	"strings"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func UDPListenAndReceive(port string, dataRxChan chan DataPackage) {

	buf := make([]byte, 1024)

	rPackage := new(DataPackage)

	laddr, err := net.ResolveUDPAddr("udp", port)
	CheckError(err)

	conn, err := net.ListenUDP("udp", laddr)
	CheckError(err)

	defer conn.Close()

	for {
		n, addr, err := conn.ReadFrom(buf)
		CheckError(err)

		rPackage.IP = strings.Split(addr.String(), ":")[0]
		rPackage.Port = strings.Split(addr.String(), ":")[1]
		rPackage.Data = buf[0:n]

		dataRxChan <- *rPackage
	}
}

func UDPTransmit(dataTxChan chan DataPackage) {

	var tPackage DataPackage

	for {
		select {
		case tPackage = <-dataTxChan:

			tAddr, err := net.ResolveUDPAddr("udp", tPackage.IP+":"+tPackage.Port)
			CheckError(err)
			conn, err := net.DialUDP("udp", nil, tAddr)
			CheckError(err)

			_, err = conn.Write(tPackage.Data)
			CheckError(err)

		default:
		}
	}
}
