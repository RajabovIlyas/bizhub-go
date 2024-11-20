package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmployeeChatRoomClientProfile struct {
	Id       primitive.ObjectID `json:"_id"`
	Avatar   string             `json:"avatar"`
	FullName string             `json:"full_name"`
	Job      models.Job         `json:"job"`
	Token    string             `json:"-"` // access token
}

type EmployeeChatRoomClient struct {
	auth    bool
	Profile *EmployeeChatRoomClientProfile
	room    *EmployeeChatRoom
	conn    *websocket.Conn
	send    chan ChatMessage
	Done    chan struct{}
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newLine = []byte("|*|")
)

func (c *EmployeeChatRoomClient) errorMessage(name string, data interface{}) ChatMessage {
	return ChatMessage{
		Type: "error", Content: ChatErrorMessageContent{
			Name: name,
			Data: data,
		},
	}
}

func (c *EmployeeChatRoomClient) Authenticate(authJsonMessage ChatMessage) error {
	fmt.Println("authenticate function-da")
	if authJsonMessage.Type == "secret" {
		fmt.Println("authenticate function-da secket tapyldy")
		token := authJsonMessage.Content.(string)

		employee, err := helpers.ValidateToken(token, os.Getenv(config.ACCT_PUBLIC_KEY))
		if err != nil {
			return err
		}
		employeeAsMap := employee.(map[string]any)

		empId, err := primitive.ObjectIDFromHex(employeeAsMap["_id"].(string))
		if err != nil {
			return err
		}

		err = c.room.IsNotClientInRoom(empId)
		if err != nil {
			return err
		}

		emplJobAsMap := employeeAsMap["job"].(map[string]any)

		empJob := models.Job{
			Name:        emplJobAsMap["name"].(string),
			DisplayName: emplJobAsMap["display_name"].(string),
		}

		c.Profile = &EmployeeChatRoomClientProfile{
			Id:       empId,
			FullName: employeeAsMap["full_name"].(string),
			Avatar:   employeeAsMap["avatar"].(string),
			Job:      empJob,
			Token:    token,
		}
		c.auth = true
		c.room.register <- c

		fmt.Printf("\nchat room client token: %v\nchat room client Id: %v\n", token, empId)
		fmt.Printf("\nchat room client profile: %v\n", c.Profile)

		return nil
	}

	return errors.New("auth secret not found")
}

func (c *EmployeeChatRoomClient) readPump() {
	defer c.Close()
	// c.conn.SetReadLimit()
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) {
				fmt.Printf("chat room error: %v\n", err)
			}
			fmt.Printf("\nchat room read message error: %v\n", err)
			return
		}

		var jsonMessage ChatMessage
		err = json.Unmarshal(message, &jsonMessage)
		if err != nil {
			fmt.Printf("chat room message decode err: %v\n", err)
			c.send <- c.errorMessage("error-message-decode", -1)
			continue
		}

		if c.auth == false {
			err := c.Authenticate(jsonMessage)
			if err != nil {
				fmt.Printf("chat room auth err: %v\n", err)
				if errors.Is(err, jwt.ErrTokenExpired) {
					c.send <- c.errorMessage("secret-expired", true)
					continue
				} else {
					return
				}
			}
			continue
		}

		jsonMessage.Id = primitive.NewObjectID()
		jsonMessage.By = c.Profile.Id

		c.room.broadcast <- jsonMessage

	}
}

func (c *EmployeeChatRoomClient) Close() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("chat room closing error [force]:", err)
		}
	}()

	c.room.unregister <- c
	c.Done <- struct{}{}
	if c.conn != nil {
		fmt.Println("current websocket conn =>", c.conn)
		c.conn.Close()
		c.conn = nil
	}

	return

}

func (c *EmployeeChatRoomClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				fmt.Printf("chat room next writer error: %v\n", err)
				return
			}

			jsonByteArray, err := json.Marshal(message)

			if err != nil {
				fmt.Printf("chat room json marshal error: %v\n", err)

				return
			}
			w.Write(jsonByteArray)

			// for i := 0; i < len(c.send); i++ {

			// 	newData, err := json.Marshal(<-c.send)

			// 	if err != nil {
			// 		fmt.Printf("chat room send message decode error: %v\n", err)
			// 		continue
			// 	}

			// 	w.Write(newLine)
			// 	w.Write(newData)
			// }

			if err := w.Close(); err != nil {
				fmt.Printf("chat room send message close error: %v\n", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("chat room write ping message error: %v\n", err)
				return
			}
		}
	}
}

func NewEmployeeChatRoomClient(room *EmployeeChatRoom, conn *websocket.Conn) *EmployeeChatRoomClient {
	client := &EmployeeChatRoomClient{
		room: room,
		conn: conn,
		send: make(chan ChatMessage),
		Done: make(chan struct{}),
		auth: false,
	}

	go client.writePump()
	go client.readPump()

	// client.Done <- struct{}{}

	return client

}
