package main

import (
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
}

func MakeTelnet(conn net.Conn) Telnet {
	telnet := &TelnetData{
		conn,
		make([]byte, 512),
		make([]byte, 0, 512),
		TopLevel,
		TelnetNop}
	telnet.initialize()
	return telnet
}

func (tc *TelnetData) initialize() {
}

func (tc *TelnetData) handleSubCommand() {
}

func (tc *TelnetData) Read(b []byte) (int, error) {
	writePos := 0

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
					if tc.subCommand == TelnetNop {
						b[writePos] = cb
						writePos++
					} else {
						tc.subBuffer = append(tc.subBuffer, cb)
					}
				}
			case GotIac:
				switch cb {
				case TelnetIac:
					if tc.subCommand == TelnetNop {
						b[writePos] = cb
						writePos++
					} else {
						tc.subBuffer = append(tc.subBuffer, cb)
					}
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
				tc.telnetState = TopLevel
			case GotWont:
				tc.telnetState = TopLevel
			case GotDo:
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