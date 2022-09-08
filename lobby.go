package main

import "encoding/json"

type Lobby struct {
	broker *LobbyBroker
	Rooms  map[string]*Room
}

func NewLobby(broker *LobbyBroker) *Lobby {
	rooms := make(map[string]*Room)
	return &Lobby{
		broker: broker,
		Rooms:  rooms,
	}
}

func (l *Lobby) Add(r *Room) error {
	l.Rooms[r.Name] = r
	return l.post()
}

func (l *Lobby) Remove(r *Room) error {
	delete(l.Rooms, r.Name)
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
