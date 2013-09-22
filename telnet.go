package main

import (
	"log"
	"net"
)

type TelnetState int

const (
	TopLevel TelnetState = iota
	GotIac
	GotWill
	GotWont
	GotDo
	GotDont
	GotSb
)

const (
	TelnetSe byte	= 240
	TelnetNop	= 241
	TelnetDataMark	= 242
	TelnetBrk	= 243
	TelnetIp	= 244
	TelnetAo	= 245
	TelnetAyt	= 246
	TelnetEc	= 247
	TelnetEl	= 248
	TelnetGa	= 249
	TelnetSb	= 250
	TelnetWill	= 251
	TelnetWont	= 252
	TelnetDo	= 253
	TelnetDont	= 254
	TelnetIac	= 255
)

const (
	TelnetEcho byte		= 1
	TelnetSuppressGoAhead	= 3
	TelnetNaws		= 31
	TelnetTerminalType	= 24
	TelnetTerminalSpeed	= 32
	TelnetLinemode		= 34
)

type Telnet interface {
	Read(b []byte) (int, error)
	Close() error
}

type TelnetData struct {
	conn		net.Conn
	buffer		[]byte
	subBuffer	[]byte
	telnetState	TelnetState
	subCommand	byte
	negotiations	int
}

func MakeTelnet(conn net.Conn) Telnet {
	telnet := &TelnetData{
		conn,
		make([]byte, 512),
		make([]byte, 0, 512),
		TopLevel,
		TelnetNop,
		4}
	telnet.initialize()
	return telnet
}

func (tc *TelnetData) initialize() {
	tc.sendCommand(TelnetIac, TelnetWill, TelnetEcho,
		TelnetIac, TelnetDont, TelnetEcho,
		TelnetIac, TelnetWill, TelnetSuppressGoAhead,
		TelnetIac, TelnetDo, TelnetSuppressGoAhead,
		TelnetIac, TelnetDo, TelnetNaws,
		TelnetIac, TelnetDo, TelnetTerminalType,
		TelnetIac, TelnetDo, TelnetTerminalSpeed,
		TelnetIac, TelnetWont, TelnetLinemode)
}

func (tc *TelnetData) handleSubCommand() {
	switch tc.subCommand {
	case TelnetNaws:
		width := uint16(tc.subBuffer[0] << 8 | tc.subBuffer[1])
		height := uint16(tc.subBuffer[2] << 8 | tc.subBuffer[3])
		log.Printf("width: %d height: %d", width, height)
	case TelnetTerminalType:
		terminalType := string(tc.subBuffer[1:])
		log.Printf("terminal-type: %s", terminalType)

		if tc.negotiations > 0 && (terminalType != "VT100" && terminalType != "XTERM") {
			tc.negotiations--
			tc.sendCommand(TelnetIac, TelnetSb, TelnetTerminalType, 1, TelnetIac, TelnetSe)
		}
	case TelnetTerminalSpeed:
		log.Printf("terminal-speed: %s", string(tc.subBuffer[1:]))
	}
}

func (tc *TelnetData) onWill(command byte) {
	switch command {
	case TelnetTerminalSpeed:
		tc.sendCommand(TelnetIac, TelnetSb, TelnetTerminalSpeed, 1, TelnetIac, TelnetSe)
	case TelnetTerminalType:
		tc.sendCommand(TelnetIac, TelnetSb, TelnetTerminalType, 1, TelnetIac, TelnetSe)
	}
}

func (tc *TelnetData) onDo(command byte) {
	switch command {
	case TelnetEcho:
		tc.sendCommand(TelnetIac, TelnetWill, TelnetEcho)
	case TelnetSuppressGoAhead:
		tc.sendCommand(TelnetIac, TelnetWill, TelnetSuppressGoAhead)
	}
}

func (tc *TelnetData) sendCommand(b ...byte) {
	tc.conn.Write(b)
}

func (tc *TelnetData) Read(b []byte) (int, error) {
	writePos := 0

	appendByte := func(cb byte) {
		if tc.subCommand == TelnetNop {
			b[writePos] = cb
			writePos++
		} else {
			tc.subBuffer = append(tc.subBuffer, cb)
		}
	}

	for writePos == 0 {
		n, err := tc.conn.Read(tc.buffer)
		if err != nil {
			return 0, err
		}

		for i := 0; i < n; {
			cb := tc.buffer[i]

			switch tc.telnetState {
			case TopLevel:
				switch cb {
				case TelnetIac:
					tc.telnetState = GotIac
				default:
					appendByte(cb)
				}
			case GotIac:
				switch cb {
				case TelnetIac:
					appendByte(cb)
					tc.telnetState = TopLevel
				case TelnetSb:
					tc.telnetState = GotSb
				case TelnetWill:
					tc.telnetState = GotWill
				case TelnetWont:
					tc.telnetState = GotWont
				case TelnetDo:
					tc.telnetState = GotDo
				case TelnetDont:
					tc.telnetState = GotDont
				case TelnetSe:
					tc.telnetState = TopLevel
					tc.handleSubCommand()
					tc.subCommand = TelnetNop
				default:
					tc.telnetState = TopLevel
				}
			case GotWill:
				tc.onWill(cb)
				tc.telnetState = TopLevel
			case GotWont:
				tc.telnetState = TopLevel
			case GotDo:
				tc.onDo(cb)
				tc.telnetState = TopLevel
			case GotDont:
				tc.telnetState = TopLevel
			case GotSb:
				tc.subCommand = cb
				tc.subBuffer = tc.subBuffer[:0]

				tc.telnetState = TopLevel
			}

			i++
		}
	}

	return writePos, nil
}

func (tc *TelnetData) Close() error {
	return tc.conn.Close()
}