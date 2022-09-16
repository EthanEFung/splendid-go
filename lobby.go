package main

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Lobby struct {
	Rooms  map[uuid.UUID]*Room
	lobbyBroker *LobbyBroker
	roomBroker *RoomBroker
}

func NewLobby(lb *LobbyBroker, rb *RoomBroker) *Lobby {
	rooms := make(map[uuid.UUID]*Room)
	return &Lobby{
		Rooms:  rooms,
		lobbyBroker: lb,
		roomBroker: rb,
	}
}

func (l *Lobby) Add(r *Room) error {
	l.Rooms[r.ID] = r
	l.roomBroker.Add(r)
	return l.post()
}

func (l *Lobby) Remove(r *Room) error {
	delete(l.Rooms, r.ID)
	l.roomBroker.Remove(r)
	return l.post()
}

func (l *Lobby) post() error {
	msg, err := json.Marshal(l.Rooms)
	if err != nil {
		return err
	}
	l.lobbyBroker.messages <- msg
	return nil
}
