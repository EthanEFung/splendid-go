package main

import (
	"log"
	"net/http"
)

/*
	LobbyBroker is responsible for sending push notifications to subscribers regarding
	the current rooms that are open and players that join the current rooms.
*/
type LobbyBroker struct {
	/*
		subscribers is a map of the current user channels that will receive the messages.
	*/
	subscribers map[chan []byte]bool
	/*
		subscribing is the channel that will be used to add a new user channel
		to `subscribers`.
	*/
	subscribing chan chan []byte

	/*
		unsubscribing is the channel used to remove an existing user channel
		from `subscribers`.
	*/
	unsubscribing chan chan []byte
	/*
		messages is the channel that receives all the messages that should be broadcasted
		to subscribers.
	*/
	messages chan []byte
}

func NewLobbyBroker() *LobbyBroker {
	broker := &LobbyBroker{
		subscribers:   make(map[chan []byte]bool),
		subscribing:   make(chan chan []byte),
		unsubscribing: make(chan chan []byte),
		messages:      make(chan []byte),
	}

	go broker.listen()

	return broker
}

func (b *LobbyBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	messageChan := make(chan []byte)

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
			for channel := range b.subscribers {
				channel <- msg
			}
		}
	}
}