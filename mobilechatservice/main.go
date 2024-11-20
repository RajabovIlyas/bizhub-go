package mobilechatservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	transactionmanager "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/devzatruk/bizhubBackend/ws"
	"github.com/savsgio/gotils/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ClientNotFound  = errors.New("client not found")
	RoomNotFound    = errors.New("room not found")
	RoomNotActive   = errors.New("room not active")
	ArgsNotProvided = errors.New("arguments not provided")
)

var (
	CollRooms        = "mobile_rooms"
	CollRoomMessages = "mobile_room_messages"
	CollClients      = "customers"
)

type MobileChatService struct {
	clients     map[string]*MobileChatClient
	rooms       map[string]*MobileChatRoom
	mu          *sync.Mutex
	db          *mongo.Database
	newRoomChan chan *MobileChatRoom
	ws          *ws.OjoWS
}

func NewMobileChatService() *MobileChatService {
	service := &MobileChatService{
		clients: map[string]*MobileChatClient{},
		rooms:   map[string]*MobileChatRoom{},
		mu:      &sync.Mutex{},
	}

	return service
}

// init and run()
func (s *MobileChatService) Init(db *mongo.Database, ws *ws.OjoWS) {
	s.db = db
	s.ws = ws
	s.run()
}

func (s *MobileChatService) run() {
	s.loadClients()
	s.loadRooms()
}

func (s *MobileChatService) loadRooms() {
	roomsColl := s.db.Collection(CollRooms)
	ctx := context.Background()
	c, err := roomsColl.Aggregate(ctx, bson.A{})
	if err != nil {
		panic(fmt.Errorf("mobilechatservice: %v", err))
	}

	rooms := map[string]*MobileChatRoom{}

	for c.Next(ctx) {
		var room MobileChatRoomExported
		err := c.Decode(&room)
		if err != nil {
			panic(err)
		}
		clients := map[string]*MobileChatClient{}

		for _, c := range room.Clients {
			client, err := s.Client(c)
			if err != nil {
				panic(err)
			}
			// if client.activeRooms[room.Id] {
			clients[client.Id.Hex()] = client
			// }
		}

		roomAsStruct := &MobileChatRoom{
			Id:         room.Id,
			clients:    clients,
			mu:         &sync.Mutex{},
			service:    s,
			newMessage: make(chan *MobileChatMessage, 1000),
			done:       make(chan bool),
			ojoRoom:    s.ws.Room(s.generateRoomWsId(room.Id)),
		}

		fmt.Printf("\nnew mobile chat room: %+v\n", roomAsStruct)

		roomAsStruct.run()

		for _, cl := range clients {
			cl.mu.Lock()
			cl.rooms[roomAsStruct.Id.Hex()] = roomAsStruct
			cl.mu.Unlock()
			fmt.Printf("\nclient[%v] rooms count: %v\n", cl.Id, len(cl.rooms))
		}

		rooms[roomAsStruct.Id.Hex()] = roomAsStruct
	}

	s.mu.Lock()
	s.rooms = rooms
	s.mu.Unlock()
}

func (s *MobileChatService) generateRoomWsId(id primitive.ObjectID) string {
	return fmt.Sprintf("mcr[%v]", id.Hex())
}

func (s *MobileChatService) loadClients() {
	clientsColl := s.db.Collection(CollClients)
	ctx := context.Background()
	c, err := clientsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"status": bson.M{"$ne": "deleted"},
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":   1,
				"rooms": 1,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	clients := map[string]*MobileChatClient{}

	for c.Next(ctx) {
		var client struct {
			Id    primitive.ObjectID   `bson:"_id"`
			Rooms []primitive.ObjectID `bson:"rooms"`
		}
		err := c.Decode(&client)
		if err != nil {
			panic(err)
		}

		activeRooms := map[primitive.ObjectID]bool{}

		for _, r := range client.Rooms {
			activeRooms[r] = true
		}

		clientAsStruct := &MobileChatClient{
			rooms:       map[string]*MobileChatRoom{},
			mu:          &sync.Mutex{},
			service:     s,
			Id:          client.Id,
			activeRooms: activeRooms,
			// tempOjoSocketId: []string{},
		}

		clients[clientAsStruct.Id.Hex()] = clientAsStruct
		fmt.Printf("\n[mobileChatService] - client - %+v | len(activeRooms): %+v\n", clientAsStruct.Id, len(clientAsStruct.activeRooms))
	}

	s.mu.Lock()
	s.clients = clients
	fmt.Printf("\n[MobileChatService] - clients.len() - %v\n", len(s.clients))
	s.mu.Unlock()
}

func (s *MobileChatService) GetClientIdBySellerId(id primitive.ObjectID) (primitive.ObjectID, error) {
	ctx := context.Background()

	c, err := s.db.Collection(CollClients).Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": id,
			},
		},
		bson.M{
			"$project": bson.M{
				"_id": 1,
			},
		},
	})
	if err != nil {
		return primitive.NilObjectID, err
	}

	if c.Next(ctx) {
		var d struct {
			Id primitive.ObjectID `bson:"_id"`
		}
		err := c.Decode(&d)
		if err != nil {
			return primitive.NilObjectID, err
		}

		return d.Id, nil
	}

	if err := c.Err(); err != nil {
		return primitive.NilObjectID, err
	}

	return primitive.NilObjectID, ClientNotFound
}

func (s *MobileChatService) Client(id primitive.ObjectID) (*MobileChatClient, error) {
	for _, client := range s.clients {
		if client.Id == id {
			return client, nil
		}
	}

	return nil, ClientNotFound
}

func (s *MobileChatService) generateId(prefix string) string {
	return strings.ToUpper(fmt.Sprintf("%v%v", prefix, uuid.V4()))
}

func (s *MobileChatService) AddClient(id primitive.ObjectID) (*MobileChatClient, error) {

	client := &MobileChatClient{
		// tempOjoSocketId: []string{},
		rooms:       map[string]*MobileChatRoom{},
		mu:          &sync.Mutex{},
		Id:          id,
		service:     s,
		activeRooms: map[primitive.ObjectID]bool{},
	}

	s.mu.Lock()
	s.clients[client.Id.Hex()] = client
	s.mu.Unlock()

	return client, nil
}

func (s *MobileChatService) RemoveClient(id primitive.ObjectID, t string) error {
	clientId := "-"

	s.mu.Lock()
	for _, client := range s.clients {
		if id == client.Id {
			clientId = client.Id.Hex()
			break
		}
	}

	if clientId == "-" {
		s.mu.Unlock()
		return ClientNotFound
	}

	delete(s.clients, clientId)
	s.mu.Unlock()

	return nil
}

func (s *MobileChatService) Room(id primitive.ObjectID) (*MobileChatRoom, error) {
	room, ok := s.rooms[id.Hex()]
	if ok {
		return room, nil
	}

	return nil, RoomNotFound
}

type CreateRoomClient struct {
	Id   primitive.ObjectID
	Type string //* customer, seller
}

func (s *MobileChatService) CreateRoom(clients []primitive.ObjectID) (*MobileChatRoom, error) {
	clientsAsStruct := []*MobileChatClient{}

	for _, v := range clients {
		client, err := s.Client(v)
		if err != nil {
			fmt.Printf("\ncreate room client: %v\n", v)
			return nil, err
		}
		clientsAsStruct = append(clientsAsStruct, client)
	}

	roomId := primitive.NewObjectID()

	clientsOfRoom := map[string]*MobileChatClient{}

	for _, c := range clientsAsStruct {
		clientsOfRoom[c.Id.Hex()] = c
	}

	room := &MobileChatRoom{
		clients:    clientsOfRoom,
		mu:         &sync.Mutex{},
		service:    s,
		Id:         roomId,
		newMessage: make(chan *MobileChatMessage, 1000),
		done:       make(chan bool),
		ojoRoom:    s.ws.Room(s.generateRoomWsId(roomId)),
	}

	ctx := context.Background()
	tran := transactionmanager.NewTransaction(&ctx, s.db, 3)
	model := transactionmanager.NewModel()
	model.SetDocument(room.export())
	_, err := tran.Collection(CollRooms).InsertOne(model)
	if err != nil {
		tran.Rollback()
		return nil, err
	}

	s.mu.Lock()
	s.rooms[room.Id.Hex()] = room
	s.mu.Unlock()

	room.run()

	for _, c := range clientsAsStruct {
		model2 := transactionmanager.NewModel()
		model2.SetFilter(bson.M{
			"_id": c.Id,
		})
		model2.SetUpdate(bson.M{
			"$push": bson.M{
				"rooms": room.Id,
			},
		})
		model2.SetRollbackUpdate(bson.M{
			"$pop": bson.M{
				"rooms": room.Id,
			},
		})
		_, err := tran.Collection(CollClients).FindOneAndUpdate(model2)
		if err != nil {
			tran.Rollback()
			return nil, err
		}
		c.mu.Lock()
		c.rooms[room.Id.Hex()] = room
		c.activeRooms[room.Id] = true
		c.mu.Unlock()
	}

	if err := tran.Err(); err != nil {
		tran.Rollback()
		return nil, err
	}

	return room, nil
}
