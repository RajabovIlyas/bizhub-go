package manager

import (
	"context"
	"sync"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/ojologger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmployeesChatClient struct {
	Id       primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	ClientId string             `json:"client_id" bson:"client_id"`
	FullName string             `json:"full_name" bson:"full_name"`
	Avatar   string             `json:"avatar" bson:"avatar"`
	Job      string             `json:"job" bson:"job"`
	IsOnline bool               `json:"is_online" bson:"is_online"`
}

type EmployeesChatMessage struct {
	Id        primitive.ObjectID   `json:"_id" bson:"_id"`
	By        primitive.ObjectID   `json:"by" bson:"by"`
	Type      string               `json:"type" bson:"type"`
	Content   any                  `json:"content" bson:"content"`
	Client    *EmployeesChatClient `json:"-" bson:"-"`
	CreatedAt time.Time            `json:"created_at" bson:"created_at"`
}

type EmployeesChatManager struct {
	Clients  map[string]*EmployeesChatClient
	Messages map[string]*EmployeesChatMessage
	mu       *sync.Mutex
	logger   *ojologger.OjoLogger
}

func NewEmployeesChatManager() *EmployeesChatManager {
	manager := &EmployeesChatManager{
		mu:       &sync.Mutex{},
		Clients:  map[string]*EmployeesChatClient{},
		Messages: map[string]*EmployeesChatMessage{},
		logger:   ojologger.LoggerService.Logger("EmployeesChatManager"),
	}
	// temporary manager
	// load messages from db to here

	return manager
}

func (m *EmployeesChatManager) Init() {
	log := m.logger.Group("Init()")

	clients := map[string]*EmployeesChatClient{}

	cursor, err := config.MI.DB.Collection("employees").Aggregate(context.Background(), bson.A{
		bson.M{
			"$match": bson.M{
				"exited_on": nil,
			},
		},

		bson.M{
			"$project": bson.M{
				"full_name": bson.M{
					"$concat": bson.A{"$name", " ", "$surname"},
				},
				"avatar": 1,
				"job":    "$job.name",
			},
		},
	})

	if err != nil {
		log.Error(err)
		panic(err)
	}

	for cursor.Next(context.Background()) {
		var client EmployeesChatClient
		err := cursor.Decode(&client)
		if err != nil {
			log.Error(err)
			panic(err)
		}

		clients[client.Id.Hex()] = &client

	}

	log.Logf("loaded clients: %v", clients)

	m.Clients = clients

}

func (m *EmployeesChatManager) NewClient(clientId string, id primitive.ObjectID, fullName string, avatar string, job string) *EmployeesChatClient {
	if _, ok := m.Clients[id.Hex()]; ok {
		m.mu.Lock()
		m.Clients[id.Hex()].ClientId = clientId
		m.Clients[id.Hex()].IsOnline = true
		m.mu.Unlock()
	} else {
		m.mu.Lock()
		m.Clients[id.Hex()] = &EmployeesChatClient{
			Id:       id,
			ClientId: clientId,
			FullName: fullName,
			Avatar:   avatar,
			Job:      job,
			IsOnline: true,
		}
		m.mu.Unlock()
	}
	return m.Clients[id.Hex()]
}

func (m *EmployeesChatManager) LeaveClient(id primitive.ObjectID) {
	if _, ok := m.Clients[id.Hex()]; ok {
		m.mu.Lock()
		m.Clients[id.Hex()].IsOnline = false
		m.mu.Unlock()
	}
}

func (m *EmployeesChatManager) NewMessage(message *EmployeesChatMessage) {
	m.mu.Lock()
	m.Messages[message.Id.Hex()] = message
	m.mu.Unlock()
}

func (m *EmployeesChatManager) GetLatestMessages() []*EmployeesChatMessage {
	list := []*EmployeesChatMessage{}
	for _, message := range m.Messages {
		list = append(list, message)
	}
	return list
}

func (m *EmployeesChatManager) GetAllClientsWithoutMe(id primitive.ObjectID) []*EmployeesChatClient {
	list := []*EmployeesChatClient{}

	for _, client := range m.Clients {
		if client.Id != id {
			list = append(list, client)
		}
	}

	return list
}

func (m *EmployeesChatManager) GetAllMessages() []*EmployeesChatMessage {
	list := []*EmployeesChatMessage{}

	for _, message := range m.Messages {
		list = append(list, message)
	}

	return list
}
