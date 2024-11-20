package mobilechatservice

import (
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SliceContainsBy[T comparable](s []T, element T) bool {
	for _, value := range s {
		if value == element {
			return true
		}
	}
	return false
}

type MobileChatClient struct {
	Id primitive.ObjectID //* customer_id
	// tempOjoSocketId []string           //* mutable
	// sellerId *primitive.ObjectID
	rooms       map[string]*MobileChatRoom
	activeRooms map[primitive.ObjectID]bool
	mu          *sync.Mutex
	service     *MobileChatService
}

type MobileChatClientRoom struct {
	Id          primitive.ObjectID               `json:"_id" bson:"_id"`
	Logo        string                           `json:"logo" bson:"logo"`
	Name        string                           `json:"name" bson:"name"`
	LastMessage *MobileChatClientRoomLastMessage `json:"last_message" bson:"last_message"`
}

type MobileChatClientRoomLastMessage struct {
	Type string `json:"type" bson:"type"`
	Text string `json:"text" bson:"text"`
}

func (c *MobileChatClient) ActivateRealtime(j func(string)) {
	fmt.Printf("\nclient rooms: len(%v)\n", len(c.rooms))
	for _, room := range c.rooms {
		j(c.service.generateRoomWsId(room.Id))
	}
}
func (c *MobileChatClient) ActivateRoom(id primitive.ObjectID, j func(string), emit func(string, ...any)) (*MobileChatRoom, error) {
	room, err := c.service.Room(id)
	if err != nil {
		return nil, err
	}

	_, err = c.service.db.Collection(CollClients).UpdateOne(context.Background(), bson.M{
		"_id": c.Id,
	}, bson.M{
		"$push": bson.M{
			"rooms": room.Id,
		},
	})
	if err != nil {
		return nil, err
	}

	room.mu.Lock()
	room.clients[c.Id.Hex()] = c
	room.mu.Unlock()

	c.mu.Lock()
	c.activeRooms[room.Id] = true
	c.mu.Unlock()

	j(room.ojoRoom.Name)

	emit("new-room", true)
	// room.ojoRoom.Emit("")

	return room, nil
}

func (c *MobileChatClient) GetRoomByClients(clients []primitive.ObjectID) (*MobileChatRoom, error) {
	for _, room := range c.rooms {

		s := 0

		roomClients := []primitive.ObjectID{}
		for _, cc := range room.clients {
			roomClients = append(roomClients, cc.Id)
		}

		for _, client := range clients {
			if SliceContainsBy(roomClients, client) {
				s++
			}
		}

		if s == len(clients) {
			if !c.activeRooms[room.Id] {
				return room, RoomNotActive
			}
			return room, nil
		}

	}

	return nil, RoomNotFound
}

// func (c *MobileChatClient) SetWsId(id string, args ...any) {
// 	if args[0] == "delete" {
// 		c.mu.Lock()
// 		t := []string{}
// 		for _, v := range c.tempOjoSocketId {
// 			if v != id {
// 				t = append(t, v)
// 			}
// 		}

// 		c.tempOjoSocketId = t
// 		c.mu.Unlock()
// 	} else {
// 		c.mu.Lock()
// 		c.tempOjoSocketId = append(c.tempOjoSocketId, id)
// 		c.mu.Unlock()
// 	}
// }

func (c *MobileChatClient) Room(roomId primitive.ObjectID) (*MobileChatRoom, error) {

	// fmt.Printf("\n[client.Room] - rooms: %v | activeRooms: %v | roomId: %v | room: %+v\n", len(c.rooms), len(c.activeRooms), roomId, c.rooms)

	if room, ok := c.rooms[roomId.Hex()]; ok {
		// fmt.Printf("\ntapyldy room .........\n")
		if !c.activeRooms[roomId] {
			return nil, RoomNotActive
		}

		return room, nil
	}

	return nil, RoomNotFound
}

func (c *MobileChatClient) DeleteRoom(roomId primitive.ObjectID, l func(string)) error {
	room, err := c.service.Room(roomId)
	if err != nil {
		return err
	}

	coll := c.service.db.Collection(CollClients)
	ctx := context.Background()
	_, err = coll.UpdateOne(ctx, bson.M{
		"_id": c.Id,
	}, bson.M{
		"$pull": bson.M{
			"rooms": roomId,
		},
	})
	if err != nil {
		return err
	}

	c.mu.Lock()
	delete(c.rooms, roomId.Hex())
	delete(c.activeRooms, roomId)
	c.mu.Unlock()

	l(room.ojoRoom.Name)

	return err
}

func (c *MobileChatClient) Rooms(page int, limit int) ([]MobileChatClientRoom, error) {
	coll := c.service.db.Collection(CollClients)
	ctx := context.Background()
	cur, err := coll.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": c.Id,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         CollRooms,
				"localField":   "rooms",
				"foreignField": "_id",
				"as":           "rooms",
				"pipeline": bson.A{
					bson.M{
						"$skip": page * limit,
					},
					bson.M{
						"$limit": limit,
					},
					bson.M{
						"$addFields": bson.M{
							"clients": bson.M{
								"$filter": bson.M{
									"input": "$clients",
									"as":    "c",
									"cond": bson.M{
										"$ne": bson.A{
											c.Id, "$$c",
										},
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path": "$clients",
						},
					},
					bson.M{
						"$lookup": bson.M{
							"from":         "customers",
							"localField":   "clients",
							"foreignField": "_id",
							"as":           "customer",
							"pipeline": bson.A{
								bson.M{
									"$lookup": bson.M{
										"from":         "sellers",
										"localField":   "seller_id",
										"foreignField": "_id",
										"as":           "seller",
										"pipeline": bson.A{
											bson.M{
												"$project": bson.M{
													"logo": 1,
													"name": 1,
												},
											},
										},
									},
								},
								bson.M{
									"$unwind": bson.M{
										"path":                       "$seller",
										"preserveNullAndEmptyArrays": true,
									},
								},
								bson.M{
									"$project": bson.M{
										"seller": bson.M{
											"$ifNull": bson.A{"$seller", nil},
										},
										"logo": 1,
										"name": 1,
									},
								},
							},
						},
					},

					bson.M{
						"$unwind": bson.M{
							"path":                       "$customer",
							"preserveNullAndEmptyArrays": true,
						},
					},

					bson.M{
						"$addFields": bson.M{
							"customer": bson.M{
								"$ifNull": bson.A{"$customer", nil},
							},
						},
					},
					bson.M{
						"$addFields": bson.M{
							"selected": bson.M{
								"$cond": bson.A{
									bson.M{"$eq": bson.A{"$customer.seller", nil}},
									"$customer",
									"$customer.seller",
								},
							},
						},
					},
					bson.M{
						"$lookup": bson.M{
							"from":         CollRoomMessages,
							"localField":   "last_message",
							"foreignField": "_id",
							"as":           "last_message",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
										"type": 1,
										"text": "$content.text",
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path":                       "$last_message",
							"preserveNullAndEmptyArrays": true,
						},
					},
					bson.M{
						"$project": bson.M{
							"logo":         "$selected.logo",
							"_id":          1,
							"name":         "$selected.name",
							"last_message": "$last_message",
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$rooms",
			},
		},
		bson.M{
			"$replaceRoot": bson.M{
				"newRoot": "$rooms",
			},
		},
	})
	if err != nil {
		return []MobileChatClientRoom{}, err
	}

	rooms := []MobileChatClientRoom{}

	for cur.Next(ctx) {
		var room MobileChatClientRoom
		err := cur.Decode(&room)
		if err != nil {
			return []MobileChatClientRoom{}, err
		}

		rooms = append(rooms, room)
	}

	if err := cur.Err(); err != nil {
		return []MobileChatClientRoom{}, err
	}

	return rooms, nil
}
