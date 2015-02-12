// Server Side Events
package sse

import (
	"fmt"
	"log"
	"net/http"
)

type Broker struct {
	clients map[chan string]bool

	newClient    chan chan string
	deleteClient chan chan string

	messageChan chan string
}

func NewBroker() *Broker {
	b := &Broker{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string, 30),
	}
	go b.run()
	return b
}

func (b *Broker) run() {
	for {
		select {
		case c := <-b.newClient:
			b.clients[c] = true
			log.Println("New SSE client")

		case c := <-b.deleteClient:
			delete(b.clients, c)
			log.Println("Deleted SSE client")

		// received a message and broadcast to all clients
		case msg := <-b.messageChan:
			for keyChan, _ := range b.clients {
				keyChan <- msg
			}
			log.Println("Broadcast SSE msg to", len(b.clients))
		}
	}
}

func (b *Broker) Send(msg string) {
	b.messageChan <- msg
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming is not supported :(", http.StatusInternalServerError)
		return
	}

	messageChan := make(chan string)
	b.newClient <- messageChan

	// channel to notify if the connection is closed
	notify := w.(http.CloseNotifier).CloseNotify()

	// headers for event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		// connection is closed
		case <-notify:
			b.deleteClient <- messageChan
			log.Println("SSE HTTP connection closed")
			f.Flush()
			return

		// message recieved
		case msg := <-messageChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			log.Printf("data: %s\n\n", msg)
			f.Flush()
		}
	}

}
