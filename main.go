package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	lobbyBroker := NewLobbyBroker()
	lobby := NewLobby(lobbyBroker)
	lobbyBroker.setLobby(lobby)

	e.Use(middleware.CORS())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/lobby", echo.WrapHandler(lobbyBroker))

	e.GET("/join", func(c echo.Context) error {
		/*
			... TODO
		*/
		return nil
	})

	e.POST("/create", func(c echo.Context) error {
		type parameters struct {
			Roomname string `json:"roomname" form:"roomname" query:"roomname"`
		}
		p := new(parameters)
		if err := c.Bind(p); err != nil {
			return err
		}
		if p.Roomname == "" {
			return c.String(http.StatusBadRequest, "roomname is required")
		}
		room := NewRoom(p.Roomname)
		lobby.Add(room)
		return c.String(http.StatusCreated, "Created")
	})

	e.Logger.Fatal(e.Start(":8080"))
}
