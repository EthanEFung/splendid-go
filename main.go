package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/ethanefung/splendid-go/middlewares"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var secret string
var store *sessions.CookieStore

func init() {
	flag.StringVar(&secret, "s", "", "secret used to set up cookie store")
	store = sessions.NewCookieStore([]byte(secret))
}

func main() {
	flag.Parse()
	if secret == "" {
		panic("cannot run without a session secret")
	}

	/***********
	* entities *
	***********/

	e := echo.New()
	lobbyBroker := NewLobbyBroker()
	roomValidator := NewRoomValidator()
	roomBroker := NewRoomBroker(roomValidator)
	lobby := NewLobby(lobbyBroker, roomBroker)
	lobbyBroker.setLobby(lobby)
	roomBroker.setLobby(lobby)
	roomValidator.setLobby(lobby)

	/**************
	* middlewares *
	**************/

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		/*
			! This should change in production. Do not deploy unless this is properly set.
		*/
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
	}))
	e.Use(session.Middleware(store))

	/*********
	* routes *
	*********/

	e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	e.GET("/lobby", echo.WrapHandler(lobbyBroker), middlewares.CreateSessionToken)
	e.POST("/create", func(c echo.Context) error {
		type parameters struct {
			Roomname string `json:"roomname" form:"roomname" query:"roomname"`
			Players  string `json:"players"`
		}
		p := new(parameters)
		if err := c.Bind(p); err != nil {
			log.Printf("\n-----ERR:\n%v", err)
			return err
		}
		if p.Roomname == "" {
			return c.String(http.StatusBadRequest, "roomname is required")
		}
		players, err := strconv.Atoi(p.Players)
		if err != nil {
			return c.String(http.StatusBadRequest, "players must be a number")
		}
		if players < 2 || players > 4 {
			return c.String(http.StatusBadRequest, "players must be 2 - 4 only")
		}
		room := NewRoom(p.Roomname, lobby.roomBroker)
		room.setPlayers(players)
		lobby.Add(room)
		room.setHost(c)
		log.Printf("\n_________________________\n")
		return c.String(http.StatusCreated, room.ID.String())
	}, middlewares.AuthenticateToken)
	e.GET("/room/:id", roomBroker.HanderFunc, middlewares.AuthenticateToken)

	e.POST("/room/:id/start", func(c echo.Context) error {
		// validate if this is a room
		roomStr := c.Param("id")
		roomID, err := uuid.Parse(roomStr)
		if err != nil {
			log.Println("roomID could not be parsed")
			return err
		}
		room, exists := lobby.Rooms[roomID]
		if !exists {
			return errors.New("no such room")
		}
		// validate this is coming from the host
		userID := c.Get("user-id")
		log.Printf("userID: %v", userID)

		if room.Host != userID.(string) {
			/*
				! currently if the user who is hosting leaves the room,
				  the user upon returning is given a new user ID. This
					should be investigated
			*/
			return c.String(http.StatusForbidden, "user is not host")
		}

		return room.Game.Start()
	}, middlewares.AuthenticateToken)
	e.POST("/room/:id/play", func(c echo.Context) error {
		return nil
	}, middlewares.AuthenticateToken)

	/*************
	* operations *
	*************/

	e.Logger.Fatal(e.Start(":8080"))
}
