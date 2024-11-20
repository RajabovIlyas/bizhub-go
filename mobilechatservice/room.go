package mobilechatservice

import (
	"context"
	"fmt"
	"sync"

	"github.com/devzatruk/bizhubBackend/models"
	transactionmanager "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/devzatruk/bizhubBackend/ws"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//* single chat room : a & b users chat
type MobileChatRoom struct {
	Id         primitive.ObjectID
	clients    map[string]*MobileChatClient
	mu         *sync.Mutex
	service    *MobileChatService
	newMessage chan *MobileChatMessage
	ojoRoom    *ws.OjoWebsocketRoom //! ojo room ulanmaly
	done       chan bool
}

type MobileChatRoomExported struct {
	Id          primitive.ObjectID   `json:"_id" bson:"_id"`
	Clients     []primitive.ObjectID `json:"clients" bson:"clients"`
	LastMessage *primitive.ObjectID  `json:"last_message" bson:"last_message,omitempty"`
}

func (r *MobileChatRoom) SendMessage(m MobileChatMessage) error {
	r.newMessage <- &m
	return nil
}

func (r *MobileChatRoom) Messages(page int, limit int, culture models.Culture) ([]MobileChatMessage, error) {
	ctx := context.Background()
	c, err := r.service.db.Collection(CollRoomMessages).Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"room": r.Id,
			},
		},
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
		bson.M{
			"$skip": page * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "products",
				"localField":   "content.product._id",
				"foreignField": "_id",
				"as":           "content.product.detail",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"image": bson.M{
								"$first": "$images",
							},
							"heading":  culture.Stringf("$heading.%v"),
							"price":    1,
							"discount": 1,
							"status":   1,
						},
					},
					bson.M{
						"$addFields": bson.M{
							"is_new": false,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$content.product.detail",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"content.product": bson.M{
					"$cond": bson.A{
						bson.M{
							"$eq": bson.A{
								bson.M{
									"$ifNull": bson.A{"$content.product.text", nil},
								},
								nil,
							},
						},
						nil,
						"$content.product",
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         CollRoomMessages,
				"foreignField": "_id",
				"localField":   "comment_of",
				"as":           "comment_of",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"content": 1,
							"type":    1,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$comment_of",
				"preserveNullAndEmptyArrays": true,
			},
		},
	})
	if err != nil {
		return []MobileChatMessage{}, err
	}

	msgs := []MobileChatMessage{}

	for c.Next(ctx) {
		var m MobileChatMessage
		err := c.Decode(&m)
		if err != nil {
			return []MobileChatMessage{}, err
		}

		msgs = append(msgs, m)
	}

	if err := c.Err(); err != nil {
		return []MobileChatMessage{}, err
	}

	return msgs, nil
}

func (r *MobileChatRoom) run() {
	go r.pump()
}

func (r *MobileChatRoom) pump() {
	ctx := context.Background()
	for {
		select {
		case message, ok := <-r.newMessage:
			if !ok {
				continue
			}
			notPtMsg := *message
			if notPtMsg.Content.Product != nil {
				notPtMsg.Content.Product.Detail = nil
			}
			tran := transactionmanager.NewTransaction(&ctx, r.service.db, 3)
			model := transactionmanager.NewModel()
			model.SetDocument(notPtMsg)
			rI, err := tran.Collection(CollRoomMessages).InsertOne(model)

			if err != nil {
				fmt.Printf("\nroom message error: %v\n", err)
				tran.Rollback()
				continue
			}

			message.Id = rI.InsertedID.(primitive.ObjectID)

			model2 := transactionmanager.NewModel()
			model2.SetFilter(bson.M{
				"_id": r.Id,
			})
			model2.SetUpdate(bson.M{
				"$set": bson.M{
					"last_message": message.Id,
				},
			})
			model2.SetRollbackUpdateWithOldData(func(i interface{}) bson.M {
				return bson.M{
					"$set": bson.M{
						"last_message": i.(map[string]any)["last_message"],
					},
				}
			})

			_, err = tran.Collection(CollRooms).FindOneAndUpdate(model2)
			if err != nil {
				fmt.Printf("\nchange room [last message] error: %v\n", err)
				tran.Rollback()
				continue
			}

			if err := tran.Err(); err != nil {
				fmt.Printf("\nroom message [tran] error: %v\n", err)
				tran.Rollback()
				continue
			}

			forM := *message
			if forM.CommentOf != nil && forM.CommentOf.Content.Product != nil {
				forM.CommentOf.Content.Product.Detail = nil
			}

			lT := ""
			if forM.Content.Text != nil {
				lT = *forM.Content.Text
			}

			r.ojoRoom.Emit("last-message", MobileChatClientRoomLastMessage{
				Type: forM.Type,
				Text: lT,
			}, forM.Room)
			r.ojoRoom.Emit("message", forM)
		case <-r.done:
			fmt.Printf("\nRoom session closed..\n")
			return
		}
	}
}

// @deprecated
// func (r *MobileChatRoom) dbToGoMessage(message MobileChatMessage) (*MobileChatMessage, error) {

// 	sender, err := r.service.Client(message.Sender, "customer")
// 	if err != nil {
// 		return nil, err
// 	}

// 	room, err := r.service.Room(message.Room)
// 	if err != nil {
// 		return nil, err
// 	}

// 	messageAsStruct := &MobileChatMessage{
// 		public:    message,
// 		Id:        message.Id,
// 		Sender:    sender,
// 		Room:      room,
// 		Content:   message.Content,
// 		CreatedAt: message.CreatedAt,
// 	}

// 	return messageAsStruct, nil
// }

// func (r *MobileChatRoom) Message(id primitive.ObjectID) *MobileChatMessage {
// 	r.mu.Lock()
// 	if m, ok := r.messages[id.Hex()]; ok {
// 		r.mu.Unlock()
// 		return m
// 	}

// 	r.mu.Unlock()
// 	return nil
// }

// @deprecated
// func (r *MobileChatRoom) load(lastMessage *primitive.ObjectID) error {
// 	coll := r.service.db.Collection("mobile_room_messages")
// 	ctx := context.Background()
// 	c, err := coll.Aggregate(ctx, bson.A{
// 		bson.M{
// 			"$match": bson.M{
// 				"room": r.Id,
// 			},
// 		},
// 		bson.M{
// 			"$sort": bson.M{
// 				"created_at": -1,
// 			},
// 		},
// 		bson.M{
// 			"$lookup": bson.M{
// 				"from":         CollRoomMessages,
// 				"localField":   "comment_of",
// 				"foreignField": "_id",
// 				"as":           "comment_of",
// 			},
// 		},
// 		bson.M{
// 			"$unwind": bson.M{
// 				"path":                       "$comment_of",
// 				"preserveNullAndEmptyArrays": true,
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	messages := map[string]*MobileChatMessage{}

// 	for c.Next(ctx) {
// 		var message MobileChatMessage
// 		err := c.Decode(&message)
// 		if err != nil {
// 			return err
// 		}

// 		m, err := r.dbToGoMessage(message)
// 		if err != nil {
// 			return err
// 		}
// 		if message.CommentOf != nil {
// 			cM, err := r.dbToGoMessage(*message.CommentOf)
// 			if err != nil {
// 				return err
// 			}
// 			m.CommentOf = cM
// 		}

// 		messages[m.Id.Hex()] = m
// 	}

// 	r.mu.Lock()
// 	r.messages = messages
// 	r.mu.Unlock()

// 	return nil
// }

func (r *MobileChatRoom) export() MobileChatRoomExported {
	clients := []primitive.ObjectID{}

	for _, c := range r.clients {
		clients = append(clients, c.Id)
	}

	return MobileChatRoomExported{
		Id:      r.Id,
		Clients: clients,
	}
}
