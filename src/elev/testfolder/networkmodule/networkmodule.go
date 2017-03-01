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

func TCPListen(rUnit UnitType) *net.TCPConn {
	rAddr, err := net.ResolveTCPAddr("tcp", rUnit.IP+":"+rUnit.Port)
	ln, err := net.ListenTCP("tcp", rAddr)
	CheckError(err)
	TCPconn, err := ln.AcceptTCP()
	CheckError(err)
	return TCPConn
}

func TCPConnect(rUnit UnitType, tUnit UnitType, closeChan chan bool) *net.TCPConn {
	rAddr, err := net.ResolveTCPAddr("tcp", rUnit.IP+":"+rUnit.Port)
	CheckError(err)
	tAddr, err := net.ResolveTCPAddr("tcp", tUnit.IP+":"+tUnit.Port)
	CheckError(err)
	TCPconn, err := net.DialTCP("tcp", rAddr, tAddr)
	CheckError(err)
	return TCPconn
}

func TCPTransmit(TxConn *net.TCPConn, data chan []bool) {
	_, err := TxConn.Write(data)
	CheckError(err)
}

func TCPReceive(RxConn *net.TCPConn, dataRxChan chan DataPackage) {
	var rPackage DataPackage

	for {
		_, err := RxConn.Read(rPackage.Data)
		CheckError(err)

		rPackage.IP = strings.Split(RxConn.RemoteAddr().Addr.String(), ":")[0]
		rPackage.Port = strings.Split(RxConn.RemoteAddr().Addr.String(), ":")[1]

		dataRxChan <- rPackage
	}
}
