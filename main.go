package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// createConnectionHandler creates a goroutine that handles a single
// telnet connection.
func createConnectionHandler(conn net.Conn) {
	logFormat := fmt.Sprintf("%s: %%s", conn.RemoteAddr().String())

	logPrintf := func(format string, v ...interface{}) {
		log.Printf(fmt.Sprintf(logFormat, format), v...)
	}

	handlerProc := func() {
		buffer := make([]byte, 512)

		for {
			n, err := conn.Read(buffer)
			if err != nil {
				logPrintf(err.Error())
				break
			}

			//log.Printf("Received %d bytes\n", n)
			logPrintf("Received %d bytes\n", n)
		}

		logPrintf("Disconnecting\n")
		conn.Close()
	}

	go handlerProc()
}

// createConnectionListener creates a goroutine that listens for incoming
// telnet connections.
func createConnectionListener(listener net.Listener) {
	listenerProc := func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf(err.Error())
				break
			}
			log.Printf("Connection from: %s\n", conn.RemoteAddr().String())
			
			createConnectionHandler(conn)
		}
	}

	go listenerProc()
}

func main() {
	listener, err := net.Listen("tcp", ":23")
	if err != nil {
		log.Fatal(err)
	}

	createConnectionListener(listener)

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf(err.Error())
			break
		}

		command := strings.TrimSpace(line)
		print(line)

		if command == "quit" {
			break
		}
	}

	listener.Close()
}
