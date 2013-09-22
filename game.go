package main

var clients map[Telnet]bool

func Initialize() {
	clients = make(map[Telnet]bool)
}

func AddClient(telnet Telnet) {
	clients[telnet] = true
}

func RemoveClient(telnet Telnet) {
	delete(clients, telnet)
}

func Update() {
}

func createUpdateProcess() {
	go Update()
}