package ws

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	// uuid "github.com/google/uuid"
)

type OjoWS struct {
	mu *sync.Mutex
	*OjoWebsocketRoomManager
	clients   map[string]*OjoWebsocketClient
	rooms     map[string]*OjoWebsocketRoom
	newRoom   chan *OjoWebsocketRoom
	newClient chan *OjoWebsocketClient
	leave     chan *OjoWebsocketClient
}

func NewOjoWS() *OjoWS {
	ojows := &OjoWS{
		clients:   make(map[string]*OjoWebsocketClient),
		rooms:     make(map[string]*OjoWebsocketRoom),
		newRoom:   make(chan *OjoWebsocketRoom, 1000),
		newClient: make(chan *OjoWebsocketClient, 1000),
		leave:     make(chan *OjoWebsocketClient, 1000),
		mu:        &sync.Mutex{},
	}
	ojows.OjoWebsocketRoomManager = NewOjoWebsocketManager(ojows, true)

	go ojows.Run()
	return ojows
}
func (ws *OjoWS) Run() {
	for {
		select {
		case room := <-ws.newRoom:
			ws.mu.Lock()
			ws.rooms[room.Name] = room // daha onceden sul isimde bir room varsa, onceki room-y override edyar, room.Name yerine UUID bersek has dogry bolar!!!!
			// ya-da atdas iki room bolmaz yaly user interface-de bir care ulanmaly!
			ws.mu.Unlock()
		case client := <-ws.newClient:
			ws.mu.Lock()
			ws.clients[client.Id] = client
			ws.mu.Unlock()
		case client := <-ws.leave:
			fmt.Println("leaving from ojo websocket:", client.Id)
			ws.mu.Lock()
			delete(ws.clients, client.Id)
			ws.mu.Unlock()
		}
	}
}

func (ws *OjoWS) Room(name string) *OjoWebsocketRoom {

	if room, ok := ws.rooms[name]; ok {
		return room
	}

	fmt.Println("hazir new room doredyarin..")
	room := &OjoWebsocketRoom{
		join:      make(chan *OjoWebsocketClient),
		leave:     make(chan *OjoWebsocketClient),
		Name:      name,
		Clients:   map[string]*OjoWebsocketClient{},
		broadcast: make(chan *OjoWebsocketMessage),
		mu:        &sync.Mutex{},
	}
	go room.Run()

	ws.newRoom <- room

	return room
}

func (ws *OjoWS) ClientsCount() int {
	return len(ws.clients)
}

func (ws *OjoWS) NewClient(callback func(ws *OjoWS, client *OjoWebsocketClient)) func(*fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		client := &OjoWebsocketClient{
			Id:                    NewClientId(),
			conn:                  c,
			emit:                  make(chan OjoWebsocketMessage,1000),
			done:                  make(chan bool),
			listeners:             map[string]map[int]OjoWebsocketListener{},
			ojoWS:                 ws,
			rooms:                 map[string]*OjoWebsocketRoom{},
			mu:                    &sync.Mutex{},
			Locals:                map[string]interface{}{},
			queueEventsFromClient: map[string][][]any{},
			isAlive:               true,
		}

		// new broadcast manager for client
		client.Broadcast = &OjoWebsocketBroadcast{
			client: client,
		}

		ws.newClient <- client

		go client.emitPump()
		go client.readPump()

		// create default room as `client.Id` for private messages
		client.Join(client.Id)

		callback(ws, client)

		<-client.done

		client.closeConnection()
	})
}
func (ws *OjoWS) ContainsRoom(name string) bool {
	if _, ok := ws.rooms[name]; ok {
		return true
	}
	return false
}
func (ws *OjoWS) getRoom(name string) *OjoWebsocketRoom {
	return ws.rooms[name]
}
func (ws *OjoWS) Emit(eventName string, payload ...(any)) {
	for _, client := range ws.clients {
		client.Emit(eventName, payload...)
	}
}
