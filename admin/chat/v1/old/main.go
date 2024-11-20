package v1

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatMessage struct {
	Id      primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	Type    string             `json:"type" bson:"type"`
	Content interface{}        `json:"content" bson:"content"`
	By      primitive.ObjectID `json:"by,omitempty" bson:"by"`
}

type ChatErrorMessageContent struct {
	Name string      `json:"name"`
	Data interface{} `json:"data,omitempty"`
}

type EmployeeChatRoom struct {
	clients    map[*EmployeeChatRoomClient]*EmployeeChatRoomClientProfile
	broadcast  chan ChatMessage
	register   chan *EmployeeChatRoomClient
	unregister chan *EmployeeChatRoomClient
}

var (
	ErrClientFound = errors.New("client found")
)

func NewEmployeeChatRoom() *EmployeeChatRoom {
	return &EmployeeChatRoom{
		clients:    make(map[*EmployeeChatRoomClient]*EmployeeChatRoomClientProfile),
		broadcast:  make(chan ChatMessage),
		register:   make(chan *EmployeeChatRoomClient),
		unregister: make(chan *EmployeeChatRoomClient),
	}
}

func (a *EmployeeChatRoom) IsNotClientInRoom(id primitive.ObjectID) error {
	for _, v := range a.clients {
		if v.Id == id {
			return ErrClientFound
		}
	}
	return nil
}

func (a *EmployeeChatRoom) Run() {
	for {
		select {
		case client := <-a.register:
			a.clients[client] = client.Profile
			fmt.Println("chat room client connected")
		case client := <-a.unregister:
			if _, ok := a.clients[client]; ok {
				delete(a.clients, client)
				close(client.send)
				fmt.Println("chat room client disconnected")
			}
		case message := <-a.broadcast:
			for client := range a.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(a.clients, client)
				}
			}
		}
	}
}
