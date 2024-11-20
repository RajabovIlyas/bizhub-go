package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	controllers "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/devzatruk/bizhubBackend/mobilechatservice"
	"github.com/devzatruk/bizhubBackend/ojologger"
	"github.com/devzatruk/bizhubBackend/ws"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SetupV1ChatRoutes(router fiber.Router) {
	chat := router.Group("/chat")
	chat.Get("/rooms", middlewares.DeSerializeCustomer, controllers.GetClientRooms)
	chat.Get("/rooms/:roomId/messages", middlewares.DeSerializeCustomer, controllers.GetClientRoomMessages)
	chat.Post("/upload", middlewares.DeSerializeCustomer, controllers.UploadChatImageFile)
	chat.Use("/realtime", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	chat.Get("/realtime", config.OjoWS.NewClient(func(ws *ws.OjoWS, client *ws.OjoWebsocketClient) {
		logger := ojologger.LoggerService.Logger("MobileChatClient")

		client.On("secret", func(data ...(any)) {
			log := logger.Group("realtime()")
			log.Log("New client connected!")
			log.SetConfig(&ojologger.OjoLogConfig{
				LogToFile:    false,
				LogToConsole: true,
			})
			accessToken := data[0].(string)
			log.Logf("\nmobileChatClient - accessToken - %v\n", accessToken)

			customer, err := helpers.ValidateToken(accessToken, os.Getenv(config.ACCT_PUBLIC_KEY))
			if err != nil {
				log.Error(err)
				client.Close()
				return
			}

			customerAsMap := customer.(map[string]any)

			id := customerAsMap["_id"].(string)
			clientObjId, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				log.Error(err)
				client.Close()
				return
			}

			chatClient, err := config.MobileChatService.Client(clientObjId)
			if err != nil {
				log.Error(err)
				client.Close()
				return
			}
			chatClient.ActivateRealtime(client.Join)

			client.On("send-message", func(data ...(any)) {
				if len(data) == 0 {
					return
				}

				message := data[0].(map[string]any)
				var roomInfo any

				if len(data) > 1 {
					roomInfo = data[1]
				}

				fmt.Printf("\n[message] - message - %+v\n", message)
				var roomMessage mobilechatservice.MobileChatMessage
				m, err := json.Marshal(message)
				if err != nil {
					log.Errorf("\nmobileChatClient - [message.Encode()] - error - %v\n", err)
					return
				}

				err = json.Unmarshal(m, &roomMessage)
				if err != nil {
					log.Errorf("\nmobileChatClient - [message.Decode()] - error - %v\n", err)
					return
				}

				room, err := chatClient.Room(roomMessage.Room)

				if err != nil {
					log.Error(err)
					if errors.Is(err, mobilechatservice.RoomNotFound) {
						if roomInfo != nil {
							otherObjId, err := primitive.ObjectIDFromHex(roomInfo.(string))
							if err != nil {
								log.Errorf("[Parse(otherObjId)] %v", err)
								return
							}
							log.Logf("[otherObjId]: %v", otherObjId)
							if len(data) > 2 {
								if data[2] == "seller" {
									otherObjId, err = config.MobileChatService.GetClientIdBySellerId(otherObjId)
									if err != nil {
										log.Errorf("[GetClientIdBySellerId()] %v", err)
										return
									}
								}
								log.Logf("[otherObjId] - changed : %v | type: %v", otherObjId, data[2])
							}

							if room_, err := chatClient.GetRoomByClients([]primitive.ObjectID{
								chatClient.Id,
								otherObjId,
							}); err == nil {
								room = room_
								roomMessage.Room = room_.Id
							} else {
								if errors.Is(err, mobilechatservice.RoomNotActive) {
									log.Logf("room activating..")
									room_, err := chatClient.ActivateRoom(room_.Id, client.Join, client.Emit)

									if err != nil {
										log.Error(err)
										return
									}

									room = room_
									log.Logf("room activated!")
								} else {
									log.Logf("otherObjId: %v", otherObjId)

									room_, err = config.MobileChatService.CreateRoom([]primitive.ObjectID{
										chatClient.Id,
										otherObjId,
									})
									if err != nil {
										log.Errorf("CreateRoom() - error - %v", err)
										return
									}

									room = room_
									roomMessage.Room = room_.Id
									client.Emit("new-room", true)
								}
							}
						}
					} else if errors.Is(err, mobilechatservice.RoomNotActive) {
						log.Logf("room activating..")
						room_, err := chatClient.ActivateRoom(roomMessage.Room, client.Join, client.Emit)

						if err != nil {
							log.Error(err)
							return
						}

						room = room_
						log.Logf("room activated!")
					} else {
						return
					}
				}

				roomMessage.Sender = clientObjId
				roomMessage.CreatedAt = time.Now()

				log.Logf("\nmobileChatClient - [message] - %+v\n", roomMessage)

				err = room.SendMessage(roomMessage)
				if err != nil {
					log.Errorf("\nmobileChatClient - [message.Send()] - error - %v\n", err)

					return
				}
			})

			client.On("delete-message", func(data ...(any)) {
				if len(data) == 0 {
					return
				}
				message := data[0].(map[string]any)
				roomId, err := primitive.ObjectIDFromHex(message["room"].(string))
				if err != nil {
					return
				}

				messageId, err := primitive.ObjectIDFromHex(message["message_id"].(string))
				if err != nil {
					return
				}

				room, err := chatClient.Room(roomId)
				if err != nil {
					return
				}

				err = room.DeleteMessage(messageId)
				if err != nil {
					return
				}

			})

			client.On("delete-room", func(data ...any) {
				if len(data) == 0 {
					return
				}

				roomId, err := primitive.ObjectIDFromHex((data[0]).(string))
				if err != nil {
					log.Error(err)
					return
				}
				err = chatClient.DeleteRoom(roomId, client.Leave)
				if err != nil {
					log.Error(err)
					return
				}

				client.Emit("delete-room", roomId.Hex())
			})
		})

	}))

}
