package main

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Room struct {
	ID        uuid.UUID        `json:"id"`
	Name      string           `json:"name"`
	Game      *Game            `json:"game"`
	Occupants map[string]*User `json:"occupants"`
	Host      string           `json:"host"`
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
func (r *Room) Join(c echo.Context) {
	user := NewUser(c)
	r.Occupants[user.ID] = user
}

/*
	Leave will remove the user from the occupants map
*/
func (r *Room) Leave(c echo.Context) {
	userID := c.Get("user-id")
	delete(r.Occupants, userID.(string))
}

func (r *Room) setHost(c echo.Context) {
	id := c.Get("user-id")
	r.Host = id.(string)
}

func (r *Room) setPlayers(n int) {
	r.Game.Max = n
}
