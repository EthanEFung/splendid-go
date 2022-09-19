package main

import (
	"log"

	"github.com/labstack/echo/v4"
)

type User struct {
	ID string `json:"id"`
	Name string `json:"name"`
}

func NewUser(c echo.Context) *User {
	userID := c.Get("user-id")
	log.Println("new user", userID)
	return &User{
		ID: userID.(string),
	}
}
