package ws

import (
	"fmt"
	"sync"
)

type OjoWebsocketRoomManager struct {
	rooms       map[string]*OjoWebsocketRoom
	exceptRooms map[string]*OjoWebsocketRoom
	ojoWS       *OjoWS
	mu          *sync.Mutex
	reset       bool
}

func NewOjoWebsocketManager(ojoWS *OjoWS, reset bool) *OjoWebsocketRoomManager {
	manager := &OjoWebsocketRoomManager{
		rooms:       map[string]*OjoWebsocketRoom{},
		exceptRooms: map[string]*OjoWebsocketRoom{},
		ojoWS:       ojoWS,
		mu:          &sync.Mutex{},
		reset:       reset,
	}
	return manager
}

func (m *OjoWebsocketRoomManager) In(roomName string) *OjoWebsocketRoomManager {

	fmt.Printf("\n m.In(): %v gozlenyarr.. \n", roomName)

	if m.ojoWS.ContainsRoom(roomName) {
		fmt.Printf("\n m.In(): %v tapyldy \n", roomName)
		m.mu.Lock()
		m.rooms[roomName] = m.ojoWS.getRoom(roomName)
		m.mu.Unlock()
	}
	return m
}

func (m *OjoWebsocketRoomManager) Except(roomName string) *OjoWebsocketRoomManager {

	fmt.Printf("\n m.Except(): %v edyaris...\n", roomName)
	if m.ojoWS.ContainsRoom(roomName) {
		fmt.Printf("\n m.Except(): %v edildi...\n", roomName)
		m.mu.Lock()
		m.exceptRooms[roomName] = m.ojoWS.getRoom(roomName)
		m.mu.Unlock()
	}
	return m
}

func (m *OjoWebsocketRoomManager) Emit(eventName string, payload ...(any)) {

	fmt.Println("roomlara hazir ugradyar.....")
	fmt.Printf("\nevent: %v\nrooms: %v\n", eventName, m.rooms)
	clients := map[string]*OjoWebsocketClient{}

	for _, room := range m.rooms {
		for _, client := range room.Clients {
			fmt.Printf("\nroom: %v client: %v\n", room.Name, client)
			s := false
			for _, exRoom := range m.exceptRooms {
				s = exRoom.ContainsClient(client.Id)
				if s == true {
					break
				}
			}
			if !s {
				clients[client.Id] = client
			}
		}
	}
	if m.reset {
		m.mu.Lock()
		m.rooms = map[string]*OjoWebsocketRoom{}
		m.exceptRooms = map[string]*OjoWebsocketRoom{}
		m.mu.Unlock()
	}

	fmt.Println("in gepciler clients:", len(clients))

	for _, client := range clients {
		client.Emit(eventName, payload...)
	}

}
