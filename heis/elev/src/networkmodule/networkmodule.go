package networkmodule

import (
	"../typedef"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func InitiateTCPCon(Target chan UnitType, receiveAddress string) (conn net.TCPConn) {

	tempAddress := Target.IP + ":" + Target.Port
	targetAddress, _ := net.ResolveTCPAddr("tcp", tempAddress)
	localAddress, _ := net.ResolveTCPAddr("tcp", receiveAddress)
	conn, err := net.DialTCP("tcp", nil, targetAddress)
	CheckError(err)

	conn.Write([]byte("Connect to:" + receiveAddress + "\x00"))

	ln, err := net.ListenTCP("tcp", localAddress)
	CheckError(err)

	conn, err = ln.Accept()
	CheckError(err)

	return conn
}


func TransmitTCP(connChan chan net.TCPConn, dataChan chan []byte) {
	for{
		select {
			case: data <- dataChan
				conn <- connChan
				conn.Write(data)
				break
			case: 
		}
	}
}