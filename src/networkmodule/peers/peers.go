package peers

import (
	"fmt"
	"net"
	"networkmodule/conn"
	"sort"
	"time"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

const interval = 100 * time.Millisecond
const timeout = 500 * time.Millisecond

func Transmitter(port int, id string, quitChan chan bool) {
	fmt.Println("Starting Peer Transmitter")
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	t := time.NewTicker(interval)

	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting peer transmitter!")
			t.Stop()
			return
		case <-t.C:
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate, quitChan chan bool) {
	fmt.Println("Starting Peer Receiver!")
	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		select {
		case <-quitChan:
			fmt.Println("Quitting peer receiver!")
			return
		default:
			updated := false

			conn.SetReadDeadline(time.Now().Add(interval))
			n, _, _ := conn.ReadFrom(buf[0:])

			id := string(buf[:n])

			// Adding new connection
			p.New = ""
			if id != "" {
				if _, idExists := lastSeen[id]; !idExists {
					p.New = id
					updated = true
				}

				lastSeen[id] = time.Now()
			}

			// Removing dead connection
			p.Lost = make([]string, 0)
			for k, v := range lastSeen {
				if time.Now().Sub(v) > timeout {
					updated = true
					p.Lost = append(p.Lost, k)
					delete(lastSeen, k)
				}
			}

			// Sending update
			if updated {
				p.Peers = make([]string, 0, len(lastSeen))

				for k := range lastSeen {
					p.Peers = append(p.Peers, k)
				}

				sort.Strings(p.Peers)
				sort.Strings(p.Lost)
				peerUpdateCh <- p
			}
		}
	}
}
