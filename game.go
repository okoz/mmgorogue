package main

import (
	"log"
	"math/rand"
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
	return &game{worldMap: MakeMap(24, 24),
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

	p := &playerEntity{commands: make([]byte, 0, 8),
		owner: g,
		screen: MakeScreen(80, 24),
		telnet: t,
		commandLock: &sync.Mutex{}}
	width, height := g.worldMap.GetSize()

	p.x = 1 + rand.Intn(width - 2)
	p.y = 1 + rand.Intn(height - 2)

	g.entities[p] = true
	return p
}

func (g *game) RemoveEntity(e Entity) {
	g.entityLock.Lock()
	defer g.entityLock.Unlock()

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
	Update()
	PostUpdate()
	GetPosition() (int, int)
	SetPosition(x, y int)
}

type PlayerEntity interface {
	Entity
	AddCommand(command byte)
}

type playerEntity struct {
	x, y	 	int
	commands	[]byte
	owner		Game
	screen		Screen
	telnet		Telnet
	commandLock	sync.Locker
}

func (p *playerEntity) Update() {
	p.commandLock.Lock()
	defer p.commandLock.Unlock()

	for _, c := range p.commands {
		x, y := p.x, p.y
		w, h := p.owner.GetMap().GetSize()

		switch c {
		case 'C':
			x = x + 1
		case 'A':
			y = y - 1
		case 'D':
			x = x - 1
		case 'B':
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

	s.Clear(0, 0, viewSize, viewSize)

	for r := 0; r < vh - oy; r++ {
		n := m.GetRow(x0, y0 + r, mw, buffer)
		s.GoTo(ox, oy + r)
		overflow := ox + n - viewSize
		s.Write(buffer[:Mini(n - overflow, n)])
	}

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

	delta := s.GetDelta()

	for _, d := range delta {
		d.Apply(p.telnet)
	}

	s.Flip()
}

func (p playerEntity) GetPosition() (int, int) {
	return p.x, p.y
}

func (p *playerEntity) SetPosition(x, y int) {
	p.x = x
	p.y = y
}

func (p *playerEntity) AddCommand(command byte) {
	p.commandLock.Lock()
	defer p.commandLock.Unlock()

	p.commands = append(p.commands, command)
}
