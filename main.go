package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
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
	}
}

var title = []string{
	"   *      *              )  (       )                  ",
	" (  `   (  `   (      ( /(  )\\ ) ( /( (                ",
	" )\\))(  )\\))(  )\\ )   )\\())(()/( )\\()))\\ )      (  (   ",
	"((_)()\\((_)()\\(()/(  ((_)\\  /(_)((_)\\(()/(      )\\ )\\  ",
	"(_()((_(_()((_)/(_))_  ((_)(_))   ((_)/(_))_ _ ((_((_) ",
	"|  \\/  |  \\/  (_)) __|/ _ \\| _ \\ / _ (_)) __| | | | __|",
	"| |\\/| | |\\/| | | (_ | (_) |   /| (_) || (_ | |_| | _| ",
	"|_|  |_|_|  |_|  \\___|\\___/|_|_\\ \\___/  \\___|\\___/|___|",
}

type stateFunc func(int, int) (stateFunc, error)

// doAuthentication handles the authentication process.
func doAuthentication(telnet Telnet) bool {
	for i := range title {
		telnet.GoTo(1, uint16(i + 1))
		telnet.Write([]byte(title[i]))
	}

	telnet.ShowCursor(true)
	
	writeLine := func(x, y int, s string) {
		telnet.GoTo(uint16(x), uint16(y))
		telnet.Write([]byte(s))
	}

	clearRect := func(x, y, w, h int) {
		clear := strings.Repeat(" ", w)

		for r := 0; r < h; r++ {
			writeLine(x, y + r, clear)
		}
	}

	var writeMenu stateFunc
	var logIn stateFunc
	var createAccount stateFunc

	writeMenu = func(x, y int) (sf stateFunc, err error) {
		sf = writeMenu
		err = nil

		clearRect(x, y, 18, 5)

		writeLine(x, y, "1. Log in")
		writeLine(x, y + 1, "2. Create account")
		writeLine(x, y + 2, "3. Disconnect")
		writeLine(x, y + 4, "Selection: ")

		selection, err := readLine(telnet, true, 1)
		if err != nil {
			sf = nil
			return
		}
		
		switch selection {
		case "1":
			sf = logIn
			clearRect(x, y, 18, 5)
		case "2":
			sf = createAccount
			clearRect(x, y, 18, 5)
		case "3":
			sf = nil
		}

		return
	}

	createAccount = func(x, y int) (sf stateFunc, err error) {
		sf = writeMenu
		err = nil

		userNameLine := y
		userNameErrorLine := y + 2

		passwordLine := y + 1
		repeatPasswordLine := y + 2
		passwordErrorLine := y + 3

		emailLine := y + 3
		
		name, password, email := "", "", ""
		for {
			writeLine(x, userNameLine, "User name: ")
			name, err = readLine(telnet, true, 16)
			if err != nil {
				return
			}

			if !theDatabase.UserExists(name) {
				clearRect(x, y, 27, 3)
				break
			}

			writeLine(x, userNameErrorLine, "User already exists")
			clearRect(x, y, 27, 1)
		}

		for {
			writeLine(x, passwordLine, "Password: ")
			password, err = readLine(telnet, false, 64)
			if err != nil {
				return
			}
			
			var repeatPassword string
			writeLine(x, repeatPasswordLine, "Repeat Password: ")
			repeatPassword, err = readLine(telnet, false, 64)
			if err != nil {
				return
			}

			if password == repeatPassword {
				clearRect(x, passwordErrorLine, 27, 3)
				break
			}

			writeLine(x, passwordErrorLine, "Passwords do not match")
			clearRect(x, repeatPasswordLine, 27, 3)
		}

		writeLine(x, emailLine, "E-Mail for password recovery: ")
		email, err = readLine(telnet, false, 254)
		print(email, "\n")

		return
	}

	authenticated := false
	logIn = func(x, y int) (sf stateFunc, err error) {
		sf = logIn
		err = nil

		writeLine(x, y, "User name: ")
		name, err := readLine(telnet, true, 16)
		if err != nil {
			return
		}

		writeLine(x, y + 1, "Password: ")
		password, err := readLine(telnet, false, 64)
		if err != nil {
			return
		}

		authenticated = theDatabase.Authenticate(name, password)
		if !authenticated {
			clearRect(x, y, 27, 2)
			writeLine(x, y + 2, "Invalid credentials")
		} else {
			sf = nil
		}

		return
	}


	curState := writeMenu

	for {
		nextState, err := curState(1, 10)
		if err != nil {
			log.Printf(err.Error())
			return false
		}

		curState = nextState
		if curState == nil {
			break
		}		
	}

	// Clear the screen.
	w, h := telnet.GetScreenSize()
	clearRect(0, 0, int(w), int(h))

	return authenticated
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

		if doAuthentication(telnet) {
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

				// Process commands that are packed in a buffer.  This will
				// mess up in certain cases when the codes are longer than
				// 3 bytes or an incomplete character is received.
				for i := 0; i < n; {
					// Escape chcaracter.
					if buffer[i] == 27 {
						end := Mini(i + 3, n)

						// Check up to 5 characters for an "end" character.  It seems to
						// show up in the longer character codes.
						for k := 1; k < 5 && i + k < n; k++ {
							if buffer[i + k] == '~' {
								end = i + k + 1
								break
							}
						}

					
						player.AddCommand(MakeCommand(buffer[i:end]))
						i = end
					} else if buffer[i] == '\r' {
						if i + 1 < n && buffer[i + 1] == '\n' {
							player.AddCommand(MakeCommand(buffer[i:i+2]))
							i += 2
						} else {
							player.AddCommand(MakeCommand([]byte{'\r', '\n'}))
							i++
						}
					} else {
						player.AddCommand(MakeCommand(buffer[i:i + 1]))
						i++
					}
				}
			}

			theGame.RemoveEntity(player)
		}

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
	c := make(chan os.Signal, 1)
	wait := make(chan int, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			os.Stdin.Close()
			<-wait
			os.Exit(0)
		}
	}()

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
	wait <- 0
}
