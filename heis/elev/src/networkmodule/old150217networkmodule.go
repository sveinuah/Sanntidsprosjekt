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

func UDPReceive(port string, dataRxChan chan DataPackage) {

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
			conn.Close()
		default:
		}
	}
}

func TCPReceive(rUnit UnitType, dataRxChan chan DataPackage) {
	rAddr, err := net.ResolveTCPAddr("tcp", rUnit.IP+":"+rUnit.Port)
	var rPackage DataPackage
	for {
		ln, err := net.ListenTCP("tcp", rAddr)
		CheckError(err)
		TCPconn, err := ln.Accept()
		CheckError(err)

		_, err = TCPconn.Read(rPackage.Data)
		CheckError(err)
		rPackage.IP = strings.Split(TCPconn.RemoteAddr().Addr.String(), ":")[0]
		rPackage.Port = strings.Split(TCPconn.RemoteAddr().Addr.String().String(), ":")[1]

		dataRxChan <- rPackage

		TCPconn.Close()
	}

}

func TCPTransmit(rUnit UnitType, dataTxChan chan DataPackage) {
	rAddr, err := net.ResolveTCPAddr("tcp", rUnit.IP+":"+rUnit.Port)
	var tPackage DataPackage
	for {
		select {
		case tPackage = <-dataTxChan:
			tAddr, err := net.ResolveTCPAddr("tcp", tPackage.IP+":"+tPackage.Port)
			TCPconn, err := net.DialTCP("tcp", rAddr, tAddr)
			TCPconn.Write(tPackage.Data)
			TCPconn.Close()

		default:
		}
	}
}
