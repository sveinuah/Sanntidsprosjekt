package networkmodule

import (
	. "../typedef/"
	"./conn"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
	"strings"
)

var globalTCPReceivePort = "20014"

func TransmitTCP(Target chan UnitType, chans ...interface{}) {
	checkArgs(chans...)

	n := 0
	for range chans {
		n++
	}

	//Fancy stuff from Network-Go
	selectCases := make([]reflect.SelectCase, n)
	typeNames := make([]string, n)
	for i, ch := range chans {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
		typeNames[i] = reflect.TypeOf(ch).Elem().String()
	}

	//We find the IP that we want to receive to
	localAddr := LocalIP() + ":" + globalTCPReceivePort
	//Creating an TCPaddr type
	lAddr, _ := net.ResolveTCPAddr("tcp", localAddr)

	//Setting the IP and port we want to write to
	targetAddr := targetUnit.IP + ":" + targetUnit.Port
	//Creating a TCPaddr type
	tAddr, _ := net.ResolveTCPAddr("tcp", targetAddr)

	//Dialing the target
	conn, err := net.DialTCP("tcp", lAddr, tAddr)
	CheckError(err)

	//Close the connection when terminating
	defer conn.Close()

	//Sending information with fancy stuff from Network-Go
	for {
		chosen, value, _ := reflect.Select(selectCases)
		buf, _ := json.Marshal(value.Interface())
		conn.Write([]byte(typeNames[chosen] + string(buf)))
	}
}

func ReceiveTCP(chans ...interface{}) {
	checkArgs(chans...)

	//
	localAddr := ":" + globalTCPReceivePort
	//Creating an TCPaddr type
	lAddr, _ := net.ResolveTCPAddr("tcp", localAddr)
	listener, _ := net.ListenTCP("tcp", lAddr)

	var buf [1024]byte

	for {
		conn, _ := listener.Accept()

		//Fancy stuff from Network-Go
		for {
			n, _ := conn.Read(buf[0:])
			for _, ch := range chans {
				T := reflect.TypeOf(ch).Elem()
				typeName := T.String()
				if strings.HasPrefix(string(buf[0:n])+"{", typeName) {
					v := reflect.New(T)
					json.Unmarshal(buf[len(typeName):n], v.Interface())

					reflect.Select([]reflect.SelectCase{{
						Dir:  reflect.SelectSend,
						Chan: reflect.ValueOf(ch),
						Send: reflect.Indirect(v),
					}})
				}
			}
		}
	}
}

// Encodes received values from `chans` into type-tagged JSON, then broadcasts
// it on `port`
func TransmitUDP(port int, chans ...interface{}) {
	checkArgs(chans...)

	n := 0
	for range chans {
		n++
	}

	selectCases := make([]reflect.SelectCase, n)
	typeNames := make([]string, n)
	for i, ch := range chans {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
		typeNames[i] = reflect.TypeOf(ch).Elem().String()
	}

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	for {
		chosen, value, _ := reflect.Select(selectCases)
		buf, _ := json.Marshal(value.Interface())
		conn.WriteTo([]byte(typeNames[chosen]+string(buf)), addr)
	}
}

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func ReceiverUDP(port int, chans ...interface{}) {
	checkArgs(chans...)

	var buf [1024]byte
	conn := conn.DialBroadcastUDP(port)
	for {
		n, _, _ := conn.ReadFrom(buf[0:])
		for _, ch := range chans {
			T := reflect.TypeOf(ch).Elem()
			typeName := T.String()
			if strings.HasPrefix(string(buf[0:n])+"{", typeName) {
				v := reflect.New(T)
				json.Unmarshal(buf[len(typeName):n], v.Interface())

				reflect.Select([]reflect.SelectCase{{
					Dir:  reflect.SelectSend,
					Chan: reflect.ValueOf(ch),
					Send: reflect.Indirect(v),
				}})
			}
		}
	}
}

// Checks that args to Tx'er/Rx'er are valid:
//  All args must be channels
//  Element types of channels must be encodable with JSON
//  No element types are repeated
// Implementation note:
//  - Why there is no `isMarshalable()` function in encoding/json is a mystery,
//    so the tests on element type are hand-copied from `encoding/json/encode.go`
func checkArgs(chans ...interface{}) {
	n := 0
	for range chans {
		n++
	}
	elemTypes := make([]reflect.Type, n)

	for i, ch := range chans {
		// Must be a channel
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg#%d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elemType := reflect.TypeOf(ch).Elem()

		// Element type must not be repeated
		for j, e := range elemTypes {
			if e == elemType {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg#%d and arg#%d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elemTypes[i] = elemType

		// Element type must be encodable with JSON
		switch elemType.Kind() {
		case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
			panic(fmt.Sprintf(
				"Channel element type must be supported by JSON, got '%s' instead (arg#%d)",
				elemType.String(), i+1))
		case reflect.Map:
			if elemType.Key().Kind() != reflect.String {
				panic(fmt.Sprintf(
					"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (arg#%d)",
					elemType.String(), i+1))
			}
		}
	}
}

func LocalIP() string {
	var localIP string

	conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
	CheckError(err)
	defer conn.Close()
	localIP = strings.Split(conn.LocalAddr().String(), ":")[0]

	return localIP
}

func CheckError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
