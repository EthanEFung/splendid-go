package main

type Game struct {
	Players []*Player `json:"players"`
}

func NewGame() *Game {
	return &Game{
		Players: []*Player{},
	}
}