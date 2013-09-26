package main

import (
	"math/rand"
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
	CreatePlayer() PlayerEntity
	RemoveEntity(e Entity)
}

type game struct {
	worldMap Map
	entities map[Entity]bool
}

func MakeGame() Game {
	return &game{MakeMap(24, 24), make(map[Entity]bool)}
}

func (g *game) GetMap() Map {
	return g.worldMap
}

func (g *game) GetEntities() map[Entity]bool {
	return g.entities
}

func (g *game) CreatePlayer() PlayerEntity {
	p := &playerEntity{commands: make([]byte, 0, 8), owner: g}
	width, height := g.worldMap.GetSize()

	p.x = 1 + rand.Intn(width - 2)
	p.y = 1 + rand.Intn(height - 2)

	g.entities[p] = true
	return p
}

func (g *game) RemoveEntity(e Entity) {
	delete(g.entities, e)
}

// Entities.

type Entity interface {
	Update()
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
}

func (p *playerEntity) Update() {
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

func (p playerEntity) GetPosition() (int, int) {
	return p.x, p.y
}

func (p *playerEntity) SetPosition(x, y int) {
	p.x = x
	p.y = y
}

func (p *playerEntity) AddCommand(command byte) {
	p.commands = append(p.commands, command)
}
