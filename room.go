package main

import "github.com/google/uuid"

type Room struct {
	ID        uuid.UUID        `json:"id"`
	Name      string           `json:"name"`
	Game      *Game            `json:"game"`
	Occupants map[string]*User `json:"occupants"`
	broker    *RoomBroker
}

func NewRoom(name string, rb *RoomBroker) *Room {
	occupants := make(map[string]*User)
	return &Room{
		ID:        uuid.New(),
		Name:      name,
		Game:      NewGame(),
		Occupants: occupants,
		broker:    rb,
	}
}

/*
	Join will add the user to the occupants map
*/
func (r *Room) Join(user *User) {
	r.Occupants[user.Name] = user
}

/*
	Leave will remove the user from the occupants map
*/
func (r *Room) Leave(user *User) {
	delete(r.Occupants, user.Name)
}

func (r *Room) setPlayers(n int) {
	r.Game.Max = n
}
