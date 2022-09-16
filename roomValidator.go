package main

import "github.com/google/uuid"

type RoomValidator struct {
	Lobby *Lobby
}

func NewRoomValidator() *RoomValidator {
	return &RoomValidator{}
}

func (v *RoomValidator) Validate(id uuid.UUID) bool {
	_, exists := v.Lobby.Rooms[id]
	return exists
}

func (v *RoomValidator) setLobby(l *Lobby) {
	v.Lobby = l
}
