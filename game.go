package main

type Game struct {
	Players []*Player `json:"players"`
	Max     int       `json:"max"`
}

func NewGame() *Game {
	return &Game{
		Players: []*Player{},
	}
}
