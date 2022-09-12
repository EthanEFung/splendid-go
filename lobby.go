package main

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Lobby struct {
	Rooms  map[uuid.UUID]*Room
	broker *LobbyBroker
}

func NewLobby(broker *LobbyBroker) *Lobby {
	rooms := make(map[uuid.UUID]*Room)
	return &Lobby{
		Rooms:  rooms,
		broker: broker,
	}
}

func (l *Lobby) Add(r *Room) error {
	l.Rooms[r.ID] = r
	return l.post()
}

func (l *Lobby) Remove(r *Room) error {
	delete(l.Rooms, r.ID)
	return l.post()
}

func (l *Lobby) post() error {
	msg, err := json.Marshal(l.Rooms)
	if err != nil {
		return err
	}
	l.broker.messages <- msg
	return nil
}
