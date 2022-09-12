package main

import "github.com/google/uuid"

type RoomValidator struct {
	Lobby *Lobby
}

func NewRoomValidator(l *Lobby) *RoomValidator {
	return &RoomValidator{l}
}

func (v *RoomValidator) Validate(id uuid.UUID) bool {
	_, exists := v.Lobby.Rooms[id]
	return exists
}
