package main

type Room struct {
	Name      string  `json:"name"`
	Game      *Game   `json:"game"`
	Occupants []*User `json:"occupants"`
}

func NewRoom(name string) *Room {
	occupants := []*User{}
	return &Room{
		Name:      name,
		Game:      NewGame(),
		Occupants: occupants,
	}
}
