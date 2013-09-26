package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var theGame Game

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

		player := theGame.CreatePlayer()

		updateIn := make(chan byte)
		updateOut := make(chan byte)
		updaterProc := func() {
			defer func() { updateOut <- 1 }()

			s := MakeScreen(80, 24)
			t := s.MakeRegion(25, 0, 55, 1)
			t.GoTo(0, 0)
			t.Write([]byte("Hello, world!"))

			for {
				select {
				case <-updateIn:
					return
				default:
					m := theGame.GetMap()
					mw, mh := m.GetSize()
					for r := 0; r < mh; r++ {
						m.GetRow(0, r, mw, buffer)
						s.GoTo(0, r)
						s.Write(buffer[:mw])
					}

					entities := theGame.GetEntities()
					for e := range entities {
						x, y := e.GetPosition()
						s.GoTo(x, y)
						s.Put('@')
					}

					delta := s.GetDelta()

					for _, d := range delta {
						d.Apply(telnet)
					}

					s.Flip()

					time.Sleep(100 * time.Millisecond)
				}
			}
		}

		go updaterProc()

		for {
			n, err := telnet.Read(buffer)
			if err != nil {
				logPrintf(err.Error())
				break
			}

			text := string(buffer[:n])
			logPrintf("Received %d bytes: %s\n", n, text)

			if n == 3 {
				player.AddCommand(buffer[2])
				/*x, y := player.GetPosition()
				w, h := theGame.GetMap().GetSize()

				switch buffer[2] {
				case 'C': // Right.
					x = x + 1
				case 'A': // Up.
					y = y - 1
				case 'D': // Left.
					x = x - 1
				case 'B': // Down.
					y = y + 1
				}

				if x < 1 {
					x = 1
				} else if x > w - 2 {
					x = w - 2
				}

				if y < 1 {
					y = 1
				} else if y > h - 2 {
					y = h - 2
				}

				player.SetPosition(x, y)*/
			}
		}

		updateIn <- 1
		<-updateOut
		
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
	theGame = MakeGame()

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
		print(line)

		if command == "quit" {
			break
		}
	}
}
