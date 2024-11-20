package ws

import (
	"fmt"
	"sync"
)

type OjoWebsocketRoom struct {
	mu        *sync.Mutex
	Name      string
	Clients   map[string]*OjoWebsocketClient
	broadcast chan *OjoWebsocketMessage
	join      chan *OjoWebsocketClient
	leave     chan *OjoWebsocketClient
}

func (r *OjoWebsocketRoom) Run() {
	for {
		select {
		case client := <-r.join:
			r.mu.Lock()
			fmt.Println("<-room.join giryar..")
			r.Clients[client.Id] = client
			client.rooms[r.Name] = r
			fmt.Println("<-room.join girdi")
			fmt.Printf("room clients count: %v", len(r.Clients))
			r.mu.Unlock()
		case client := <-r.leave:
			client.mu.Lock()
			delete(client.rooms, r.Name)
			client.mu.Unlock()
			r.mu.Lock()
			delete(r.Clients, client.Id)
			r.mu.Unlock()

			fmt.Printf("\n%v leaved from %v\n", client.Id, r.Name)

		}
	}
}
func (r *OjoWebsocketRoom) ContainsClient(clientId string) bool {
	if _, ok := r.Clients[clientId]; ok {
		return true
	}
	return false
}

func (r *OjoWebsocketRoom) Emit(event string, payload ...any) {
	for _, client := range r.Clients {
		client.Emit(event, payload...)
	}
}
