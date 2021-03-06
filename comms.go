package main

import (
	"bufio"
	"log"
	"net"
	"time"
)

var isIgnoreNext = false

/*
 * Sync over local subnet, by UDP broadcast
 */

// Send a udp broadcast message on port 51234
func udpSend(msg string) {

	// Ignore the next UDP message, because it comes from us
	isIgnoreNext = true

	sock, err := net.Dial("udp", "255.255.255.255:51234")
	if err != nil {
		log.Fatal(err)
	}
	defer sock.Close()

	sock.Write([]byte(msg))
	log.Println("UDP broadcast sent")
}

// Listen for UDP broadcast message on port 51234,
// and put them on the channel. Run it in a go routine.
func udpListen(channel chan string) {

	listener, err := net.ListenPacket("udp", "255.255.255.255:51234")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		buf := make([]byte, 1024)
		listener.ReadFrom(buf)
		if isIgnoreNext {
			isIgnoreNext = false
			continue
		}
		log.Println("UDP msg received:", string(buf))
		channel <- string(buf)
	}

}

/*
 * Sync anywhere, via TCP to remote server
 */

// Global because shared by tcpListen and tcpSend. Maybe make an object?
var remoteConn net.Conn

// Listen for messages from the server. Auto-reconnect.
func tcpListen(serverAddr string, channel chan string) {

	for { // Loop for auto-reconnect
		remoteConn = getRemoteConnection(serverAddr, true)
		defer remoteConn.Close()
		//Info("Connected to remote sync server")

		bufRead := bufio.NewReader(remoteConn)

		for { // Connection work loop
			content, err := bufRead.ReadString('\n')
			if err != nil {
				log.Println("Remote read error - re-connecting")
				remoteConn.Close()
				break
			}
			log.Println("Remote sent: " + content)

			channel <- content
		}
	}

}

// Get a connection to remote server which tells us when to pull
func getRemoteConnection(serverAddr string, isReconnect bool) net.Conn {

	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", serverAddr)
		if err == nil {
			break
		}
		if isReconnect {
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}
	return conn
}

// Send update notification to remote server
func tcpSend(conn net.Conn, msg string) error {
	_, err := conn.Write([]byte(msg))
	return err
}
