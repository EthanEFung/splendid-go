package main

type Game struct {
	Players []*Player `json:"players"`
	Max     int       `json:"max"`
	Started bool      `json:"started"`
}

func NewGame() *Game {
	return &Game{
		Players: []*Player{},
	}
}

func (g *Game) Start() error {
	return nil
}
