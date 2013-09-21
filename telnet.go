package main

import (
	"errors"
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
	InSb
)

type TelnetCommand uint8

const (
	TelnetSe	= 240
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
	telnetState	TelnetState
	subCommand	TelnetCommand
}

func MakeTelnet(conn net.Conn) Telnet {
	telnet := &TelnetData{conn, make([]byte, 512), TopLevel, TelnetNop}
	telnet.initialize()
	return telnet
}

func (tc *TelnetData) initialize() {
}

func (tc *TelnetData) Read(b []byte) (int, error) {
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
					// Write to main buffer.
				} else {
					// Write to sub command buffer.
				}
			}
		case GotIac:
		case GotWill:
		case GotWont:
		case GotDo:
		case GotDont:
		case GotSb:
		case InSb:
		}

		i++
	}

	return 0, errors.New("Hello, world!")
}

func (tc *TelnetData) Close() error {
	return tc.conn.Close()
}