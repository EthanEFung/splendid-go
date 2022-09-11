package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type msgChan chan []byte

/*
	LobbyBroker is responsible for sending push notifications to subscribers regarding
	the current rooms that are open and players that join the current rooms.
*/
type LobbyBroker struct {
	/*
		lobby is the entity the broker is responsible for managing messages to and from.
	*/
	lobby *Lobby
	/*
		subscribers is a map of the current user channels that will receive the messages.
	*/
	subscribers map[msgChan]bool
	/*
		subscribing is the channel that will be used to add a new user channel
		to `subscribers`.
	*/
	subscribing chan msgChan
	/*
		unsubscribing is the channel used to remove an existing user channel
		from `subscribers`.
	*/
	unsubscribing chan msgChan
	/*
		messages is the channel that receives all the messages that should be broadcasted
		to subscribers.
	*/
	messages msgChan
	/*
		rooms is a map with the uuid as the key and msgChannels the broker should
		broadcast to
	*/
	rooms map[string]Room
}

func NewLobbyBroker() *LobbyBroker {
	broker := &LobbyBroker{
		subscribers:   make(map[msgChan]bool),
		subscribing:   make(chan msgChan),
		unsubscribing: make(chan msgChan),
		messages:      make(msgChan),
	}

	go broker.listen()

	return broker
}

/*
	ServeHttp implements the http.Handler interface and does the following:
	ServeHTTP will set up a server-sent event stream for the client, writes
	a connection message and flushes the msg to the client. At which it will
	begin watching for events from the main message channel. Upon receiving
	it will flush the message to the client. Any time the process receives
	a `Done` signal from the request context a go routine will be waiting
	and will send a message to the broker that the client has unsubscribed.
*/
func (b *LobbyBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	messageChan := make(msgChan)

	b.subscribing <- messageChan

	ctx := r.Context()
	go func() {
		<-ctx.Done()
		b.unsubscribing <- messageChan
		log.Println("HTTP connection closed")
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	/*
		the first thing that should be done is have the current state of
		the rooms flushed to the client. This is to ensure that when
		the connection is successful, the sole connecting client can be
		fed the current state of the rooms.
	*/
	if err := b.connect(w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	for {
		msg, open := <-messageChan
		if !open {
			// if our messageChan was closed this means the client has disconnected
			break
		}

		b.post(w, msg)
		log.Println("flushing message")
		flusher.Flush()
	}
}

func (b *LobbyBroker) post(w http.ResponseWriter, msg []byte) {
	msg = append([]byte("event: message\ndata:"), msg...)
	msg = append(msg, '\n', '\n')
	w.Write(msg)
}

func (b *LobbyBroker) listen() {
	for {
		select {
		case s := <-b.subscribing:
			b.subscribers[s] = true
			log.Printf("Client added. %d clients registered\n", len(b.subscribers))
		case s := <-b.unsubscribing:
			delete(b.subscribers, s)
			log.Printf("Client removed. %d clients registered\n", len(b.subscribers))
		case msg := <-b.messages:
			/*
				TODO: the messages implementation should change where the message gives us
				context of where the message should be distributed
			*/
			for channel := range b.subscribers {
				channel <- msg
			}
		}
	}
}

func (b *LobbyBroker) setLobby(l *Lobby) {
	b.lobby = l
}

func (b *LobbyBroker) connect(w http.ResponseWriter) error {
	// just write to the current response writer
	bytes, err := json.Marshal(b.lobby.Rooms)

	if err != nil {
		return err
	}
	b.post(w, bytes)
	return nil
}
