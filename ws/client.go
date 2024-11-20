package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/savsgio/gotils/uuid"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512

	BeforeCloseConnection string = "before-close-connection"
	CloseConnection       string = "close-connection"
)

type OjoWebsocketListener func(...(any))

type OjoWebsocketClient struct {
	mu                    *sync.Mutex
	Id                    string
	conn                  *websocket.Conn
	emit                  chan OjoWebsocketMessage
	isAlive               bool
	done                  chan bool
	listeners             map[string]map[int]OjoWebsocketListener
	ojoWS                 *OjoWS
	rooms                 map[string]*OjoWebsocketRoom
	Broadcast             *OjoWebsocketBroadcast
	Locals                map[string]interface{}
	queueEventsFromClient map[string][][]any
}

type OjoWebsocketMessage struct {
	Event   string `json:"event"`
	Payload []any  `json:"payload"`
}

func NewClientId() string {
	uuid_ := uuid.V4()
	replacedString := strings.Replace(uuid_, "-", "", -1)

	return fmt.Sprintf("OJO%v", replacedString)
}

func funcEqual(a, b interface{}) bool {
	av := reflect.ValueOf(&a).Elem()
	bv := reflect.ValueOf(&b).Elem()
	return av.InterfaceData() == bv.InterfaceData()
}
func isFunc(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Func
}

func (c *OjoWebsocketClient) Get(key string) (interface{}, error) {
	if v, ok := c.Locals[key]; ok {
		return v, nil
	}
	return nil, errors.New("not found")
}

func (c *OjoWebsocketClient) Set(key string, value interface{}) {
	c.Locals[key] = value
}

func (c *OjoWebsocketClient) Join(name string) {

	fmt.Println("client.join():", name)
	if c.ojoWS.ContainsRoom(name) {
		c.mu.Lock()
		fmt.Println("client.join() room tapyldy")
		room := c.ojoWS.getRoom(name)
		room.join <- c
		c.mu.Unlock()
	} else {
		c.mu.Lock()
		room := c.ojoWS.Room(name)
		fmt.Println("new room:", room)
		room.join <- c
		fmt.Println("new room-a girdim")
		c.mu.Unlock()
	}

}
func (c *OjoWebsocketClient) Leave(name string) {
	if c.ojoWS.ContainsRoom(name) {
		room := c.ojoWS.getRoom(name)
		room.leave <- c

	}
}
func (c *OjoWebsocketClient) containsListenerEvent(name string) bool {
	if _, ok := c.listeners[name]; ok {
		return true
	}
	return false
}
func (c *OjoWebsocketClient) On(eventName string, listener OjoWebsocketListener) {
	c.mu.Lock()

	if c.containsListenerEvent(eventName) {
		c.listeners[eventName][len(c.listeners[eventName])] = listener
	} else {
		c.listeners[eventName] = map[int]OjoWebsocketListener{}
		c.listeners[eventName][0] = listener
	}

	c.mu.Unlock()

	if _, ok := c.queueEventsFromClient[eventName]; ok {
		for _, payload := range c.queueEventsFromClient[eventName] {
			c.emitToListeners(eventName, payload)
		}

		c.mu.Lock()
		delete(c.queueEventsFromClient, eventName)
		c.mu.Unlock()
	}
}
func (c *OjoWebsocketClient) Once(eventName string, listener OjoWebsocketListener) {
	if c.containsListenerEvent(eventName) {
		return
	}
	destroyFunc := func(payload ...(any)) {
	}
	destroyFunc = func(payload ...(any)) {
		c.Off(eventName, destroyFunc)
		listener(payload...)
	}

	c.On(eventName, destroyFunc)
}
func (c *OjoWebsocketClient) Off(eventName string, listener OjoWebsocketListener) {
	if c.containsListenerEvent(eventName) {
		for k, v := range c.listeners[eventName] {
			if funcEqual(v, listener) {
				c.mu.Lock()
				defer c.mu.Unlock()
				delete(c.listeners[eventName], k)
				return
			}
		}
	}
	return
}
func (c *OjoWebsocketClient) Emit(eventName string, payload ...(any)) {
	if !c.isAlive {
		return
	}

	m := OjoWebsocketMessage{
		Event:   eventName,
		Payload: payload,
	}
	c.emit <- m
	// if !ok {
	// 	fmt.Printf("\n[ojoclient] - [emit] - error - failed to send => %v: %payload\n", eventName, payload)

	// 	return;
	// }
	fmt.Printf("\n[ojoclient] - [emit] - %v: %payload\n", eventName, payload)
}

func (c *OjoWebsocketClient) emitWithResponse(eventName string, payload ...(any)) error {
	//!!! bug !!!

	f := payload[len(payload)-1]
	if !isFunc(f) {
		return errors.New("response callback not provided")
	}
	if f == nil {
		return errors.New("payload not provided")
	}

	responseEventName := fmt.Sprintf("**%v:response**", eventName)

	payload_ := payload[0 : len(payload)-1]
	c.Emit(eventName, payload_...)

	fmt.Printf("callback function type: %T", f)

	c.Once(responseEventName, f.(func(...any)))
	return nil
}
func (c *OjoWebsocketClient) Close() {
	c.done <- true
	fmt.Printf("\n[client] - connection closed!\n")
}
func (c *OjoWebsocketClient) closeConnection() {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Println("\nc.Close() recover error:", err)
	// 	}
	// }()
	fmt.Printf("\n[client] - connection closing..\n")
	if c.isAlive == false {
		return
	}
	c.mu.Lock()
	c.isAlive = false
	c.mu.Unlock()

	c.emitToListeners(BeforeCloseConnection, []any{})

	fmt.Println("[close] c.conn:", c)

	fmt.Println("[close] c.conn != null :", c.conn != nil)

	if c.conn != nil {

		for _, room := range c.rooms {
			room.leave <- c
		}
		c.ojoWS.leave <- c

		// fmt.Printf("\nc.rooms: %v\nc.conn: %v", c.rooms, c.conn)

		//* son gosuldy
		// c.conn.WriteMessage(websocket.CloseMessage, []byte{})
		// c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		close(c.emit) // c.emit channel closed
		
		err := c.conn.Close()
		if err != nil {
			fmt.Printf("\n[close] - error - %v\n", err)
		}
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		// fmt.Println("[close] sonuna yetdi")
	}

	c.emitToListeners(CloseConnection, []any{})

	return
}
func (c *OjoWebsocketClient) emitPump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case message, ok := <-c.emit:
			if !ok {
				// c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			data, err := json.Marshal(message)
			if err != nil {
				fmt.Println("ojo socket message marshal error:", err)
				continue
			}

			err = c.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println("ojo socket message send error:", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			fmt.Printf("\nojo socket ping sent!\n")
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Println("ojo socket ping error:", err)
				return
			}
		}
	}
}
func (c *OjoWebsocketClient) emitToListeners(eventName string, payload []any) {
	if c.containsListenerEvent(eventName) {

		fmt.Printf("\nevent:%v\nlistenersCount:%v\n", eventName, len(c.listeners[eventName]))
		for _, listener := range c.listeners[eventName] {
			listener(payload...)
		}
	} else {
		c.mu.Lock()
		if _, ok := c.queueEventsFromClient[eventName]; ok {
			c.queueEventsFromClient[eventName] = append(c.queueEventsFromClient[eventName], payload)
		} else {
			c.queueEventsFromClient[eventName] = [][]any{payload}
		}
		c.mu.Unlock()
	}
}
func (c *OjoWebsocketClient) readPump() {
	// defer c.Close()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error {
		fmt.Printf("\nojo websocket client pong message: %v\n", appData)
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		if c.conn == nil {
			return
		}
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) {
				fmt.Println("ojo socket `websocket.isCloseError` error:", err)
			}
			fmt.Println("ojo socket `conn.readMessage` error:", err)
			return
		}

		var MessageAsStruct OjoWebsocketMessage
		err = json.Unmarshal(message, &MessageAsStruct)
		if err != nil {
			fmt.Println("ojo socket message unmarshal error:", err)
			return
		}

		fmt.Printf("\nojo socket gelen message: %v\n", MessageAsStruct)

		fmt.Printf("\n********\nEvent: %v\nPayload: %v\n********\n", MessageAsStruct.Event, MessageAsStruct.Payload)

		c.emitToListeners(MessageAsStruct.Event, MessageAsStruct.Payload)
	}
}
