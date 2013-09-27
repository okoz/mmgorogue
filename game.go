package main

import (
	"bufio"
	"container/list"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Map interface {
	GetSize() (int, int)
	GetTile(x, y int) byte
	GetRow(x, y, width int, b []byte) int
}

type map2D struct {
	tiles	[]byte
	width	int
	height	int
}

func MakeMap(width, height int) Map {
	m := &map2D{make([]byte, width * height), width, height}

	for i := range m.tiles {
		m.tiles[i] = '.'
	}

	for i := 0; i < width; i++ {
		m.tiles[i] = '#'
		m.tiles[width * (height - 1) + i] = '#'
	}

	for i := 0; i < height; i++ {
		m.tiles[width * i] = '#'
		m.tiles[width * i + width - 1] = '#'
	}

	return m
}

func MakeMapFromFile(filename string) Map {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	m := &map2D{}
	data := make([]string, 0, 100)

	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		data = append(data, scanner.Text())
	}

	file.Close()

	m.width, m.height = len(data[0]), len(data)	
	m.tiles = make([]byte, m.width * m.height)

	for r := 0; r < m.height; r++ {
		copy(m.tiles[r * m.width:], []byte(data[r]))
	}

	return m
}

func (m *map2D) GetSize() (int, int) {
	return m.width, m.height
}

func (m *map2D) GetTile(x, y int) byte {
	return m.tiles[m.width * y + x]
}

func (m *map2D) GetRow(x, y, width int, b []byte) int {
	remainingWidth := m.width - x
	rowWidth := width
	if remainingWidth < rowWidth {
		rowWidth = remainingWidth
	}
	if len(b) < rowWidth {
		rowWidth = len(b)
	}

	for i := 0; i < rowWidth; i++ {
		b[i] = m.tiles[m.width * y + x + i]
	}

	return rowWidth
}

// Game.

type Game interface {
	GetMap() Map
	GetEntities() map[Entity]bool
	CreatePlayer(t Telnet) PlayerEntity
	RemoveEntity(e Entity)
	Update()
	Start()
	Stop()
	GetChat() ChatService
}

type game struct {
	worldMap	Map
	chatService	ChatService
	entities	map[Entity]bool
	quit		chan bool
	done		chan bool
	entityLock	sync.Locker
}

func MakeGame() Game {
	return &game{worldMap: MakeMapFromFile("world/map.txt"),//MakeMap(24, 24),
		chatService: CreateChatService(),
		entities: make(map[Entity]bool),
		quit: make(chan bool),
		done: make(chan bool),
		entityLock: &sync.Mutex{}}
}

func (g *game) GetMap() Map {
	return g.worldMap
}

func (g *game) GetEntities() map[Entity]bool {
	return g.entities
}

func (g *game) CreatePlayer(t Telnet) PlayerEntity {
	g.entityLock.Lock()
	defer g.entityLock.Unlock()

	p := &playerEntity{commands: make([]Command, 0, 8),
		owner: g,
		screen: MakeScreen(80, 24),
		telnet: t,
		commandLock: &sync.Mutex{},
		chatting: false}
	
	minX, maxX, minY, maxY := 0, 0, 0, 0
	switch rand.Intn(3) {
	case 0:
		minX = 9
		minY = 37
		maxX = 20
		maxY = 40
	case 1:
		minX = 21
		minY = 40
		maxX = 24
		maxY = 45
	case 2:
		minX = 27
		minY = 41
		maxX = 30
		maxY = 46
	}

	p.x = minX + rand.Intn(maxX - minX)
	p.y = minY + rand.Intn(maxY - minY)

	p.chatBox = p.screen.MakeRegion(25, 23, 55, 1)
	p.chatBuffer = make([]byte, 0, 128)
	p.chatArea = p.screen.MakeRegion(25, 0, 55, 23)
	p.chatHistory = list.New()

	g.entities[p] = true
	p.Initialize()
	return p
}

func (g *game) RemoveEntity(e Entity) {
	g.entityLock.Lock()
	defer g.entityLock.Unlock()

	e.Terminate()
	delete(g.entities, e)
}

func (g *game) Update() {
	g.entityLock.Lock()
	defer g.entityLock.Unlock()

	for e, _ := range g.entities {
		e.Update()
	}

	for e, _ := range g.entities {
		e.PostUpdate()
	}
}

func (g *game) run() {
	defer func() { g.done <- true }()

	for {
		select {
		case <- g.quit:
			return
		default:
			g.Update()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (g *game) Start() {
	go g.run()
	log.Printf("Game started\n")
}

func (g *game) Stop() {
	g.quit <- true
	<- g.done
	log.Printf("Game stopped\n")
}

func (g *game) GetChat() ChatService {
	return g.chatService
}

// Entities.

type Entity interface {
	Initialize()
	Update()
	PostUpdate()
	Terminate()
	GetPosition() (int, int)
	SetPosition(x, y int)
}

type Command []byte

func MakeCommand(b []byte) Command {
	cmd := make(Command, len(b))
	copy(cmd, b)
	return cmd
}

type PlayerEntity interface {
	Entity
	AddCommand(command Command)
}

type playerEntity struct {
	x, y	 	int
	commands	[]Command
	owner		Game
	screen		Screen
	telnet		Telnet
	commandLock	sync.Locker

	chatBox		Region
	chatBuffer	[]byte
	chatting	bool
	chatArea	Region
	chatHistory	*list.List
}

func (p *playerEntity) onChat(o Entity, m string) {
	p.chatHistory.PushBack(m)
	if p.chatHistory.Len() >= 25 {
		p.chatHistory.Remove(p.chatHistory.Front())
	}
}

func (p *playerEntity) Initialize() {
	p.owner.GetChat().Register(p, p.onChat)
}

func (p *playerEntity) Update() {
	p.commandLock.Lock()
	defer p.commandLock.Unlock()

	m := p.owner.GetMap()

	for _, c := range p.commands {
		x, y := p.x, p.y
		
		switch len(c) {
		case 3:
			switch c[2] {
			case 'C':
				x = x + 1
			case 'A':
				y = y - 1
			case 'D':
				x = x - 1
			case 'B':
				y = y + 1
			}
		case 2:
			if c[0] == '\r' && c[1] == '\n' {
				p.owner.GetChat().Send(p, string(p.chatBuffer))
				p.chatBuffer = p.chatBuffer[:0]
			}
		case 1:
			buf := p.chatBuffer
			bufLen := len(buf)

			// Escape.
			if c[0] == 27 {
				p.chatBuffer = p.chatBuffer[:0]
			}

			// Backspace.
			if c[0] == 127 && bufLen > 0 {
				p.chatBuffer = p.chatBuffer[:bufLen - 1]
			}
				
			// Printable character.
			if c[0] >= 32 && c[0] != 127 && len(buf) < cap(buf) {
				p.chatBuffer = append(p.chatBuffer, c[0])
			}
		}

		if m.GetTile(x, y) == '~' || m.GetTile(x, y) == '#' {
			continue
		}

		p.x, p.y = x, y
	}

	p.commands = p.commands[:0]
}

func (p *playerEntity) PostUpdate() {
	s := p.screen
	m := p.owner.GetMap()
	mw, mh := m.GetSize()
	buffer := make([]byte, 512)

	// Compute visible rectangle.
	viewSize := 24
	x0, y0 := p.x - viewSize / 2, p.y - viewSize / 2
	
	ox, oy := 0, 0
	if x0 < 0 {
		ox = -x0
		x0 = 0
	}
	if y0 < 0 {
		oy = -y0
		y0 = 0
	}

	x1, y1 := Mini(mw, x0 + viewSize), Mini(mh, y0 + viewSize)
	_, vh := x1 - x0, y1 - y0

	// Draw the world.
	s.Clear(0, 0, viewSize, viewSize, '~')
	for r := 0; r < vh - oy; r++ {
		n := m.GetRow(x0, y0 + r, mw, buffer)
		s.GoTo(ox, oy + r)
		overflow := ox + n - viewSize
		s.Write(buffer[:Mini(n - overflow, n)])
	}

	// Draw world entities visible to the player.
	entities := p.owner.GetEntities()
	for e := range entities {
		x, y := e.GetPosition()
		x, y = ox + x - x0, oy + y - y0
		if x < 0 || x >= viewSize || y < 0 || y >= viewSize {
			continue
		}
		s.GoTo(x, y)
		s.Put('@')
	}

	// Draw chat area.
	w, h := p.chatBox.GetSize()
	p.chatBox.Clear(0, 0, w, h, ' ')
	p.chatBox.GoTo(0, 0)
	p.chatBox.Write(p.chatBuffer[Maxi(0, len(p.chatBuffer) - 55):])

	w, h = p.chatArea.GetSize()
	p.chatArea.Clear(0, 0, w, h, ' ')
	r := h - 1
	for e := p.chatHistory.Back(); e != nil; e = e.Prev() {
		if r < 0 {
			break
		}

		msg := []byte(e.Value.(string))

		p.chatArea.GoTo(0, r)
		p.chatArea.Write(msg)
		r--
	}

	delta := s.GetDelta()

	for _, d := range delta {
		d.Apply(p.telnet)
	}

	s.Flip()
}

func (p *playerEntity) Terminate() {
	p.owner.GetChat().Unregister(p)
}

func (p playerEntity) GetPosition() (int, int) {
	return p.x, p.y
}

func (p *playerEntity) SetPosition(x, y int) {
	p.x = x
	p.y = y
}

func (p *playerEntity) AddCommand(command Command) {
	p.commandLock.Lock()
	defer p.commandLock.Unlock()

	p.commands = append(p.commands, command)
}
