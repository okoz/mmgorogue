package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var theGame Game
var theDatabase Database

func readLine(telnet Telnet, echo bool, n int) (string, error) {
	buffer := make([]byte, 512)
	sb := make([]byte, 0, n)

	for {
		m, err := telnet.Read(buffer)
		if err != nil {
			return "", err
		}

		switch m {
		case 1:
			if buffer[0] == 127 {
				if len(sb) > 0 {
					sb = sb[:len(sb) - 1]
					if echo {
						telnet.Put(buffer[0])
					}
				}
			} else if len(sb) < n {
				sb = append(sb, buffer[0])
				if echo {
					telnet.Put(buffer[0])
				}
			}
		case 2:
			if buffer[0] == '\r' && buffer[1] == '\n' {
				return string(sb), nil
			}
		}

		print(string(sb))
		print("\n")
	}
}

// createConnectionHandler creates a goroutine that handles a single
// telnet connection.
func createConnectionHandler(conn net.Conn) {
	logFormat := fmt.Sprintf("%s: %%s", conn.RemoteAddr().String())

	logPrintf := func(format string, v ...interface{}) {
		log.Printf(fmt.Sprintf(logFormat, format), v...)
	}

	handlerProc := func() {
		buffer := make([]byte, 512)

		telnet := MakeTelnet(conn)
		defer telnet.Close()

		telnet.ShowCursor(false)
		player := theGame.CreatePlayer(telnet)

		for {
			n, err := telnet.Read(buffer)
			if err != nil {
				logPrintf(err.Error())
				break
			}

			text := string(buffer[:n])
			logPrintf("Received %d bytes: %s\n", n, text)

			player.AddCommand(MakeCommand(buffer[:n]))
		}

		theGame.RemoveEntity(player)

		logPrintf("Disconnecting\n")
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
	theDatabase = MakeDatabase()
	theGame = MakeGame()

	theGame.Start()

	listener, err := net.Listen("tcp", ":23")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	createConnectionListener(listener)

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf(err.Error())
			break
		}

		command := strings.TrimSpace(line)

		if command == "quit" {
			break
		} else {
			fmt.Printf("Unknown command: %s\n", command)
		}
	}

	theGame.Stop()
	theDatabase.Close()
}
