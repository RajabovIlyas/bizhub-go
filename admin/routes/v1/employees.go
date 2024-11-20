package v1

import (
	"fmt"
	"os"
	"time"

	"github.com/devzatruk/bizhubBackend/admin/chat/v1/manager"
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/devzatruk/bizhubBackend/ws"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SetupAdminEmployeeRoutes(router fiber.Router) {
	employees := router.Group("/employees")
	employeesChatManager := manager.NewEmployeesChatManager()
	employeesChatManager.Init()
	// chatRoom := chat.NewEmployeeChatRoom()
	// go chatRoom.Run()

	employees.Get("/", middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN, config.EMPLOYEES_MANAGER}),
		controllers.GetAllEmployees)

	employees.Post("/",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.AddNewEmployee)

	employees.Use("/gepciler", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		// fmt.Printf("\n--- ws headers: %v\n", c.GetReqHeaders())
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	employees.Get("/gepciler", config.OjoWS.NewClient(func(ojows *ws.OjoWS, client *ws.OjoWebsocketClient) {
		client.On("secret", func(args ...(any)) {
			// fmt.Printf("\ngepciler clients count: %v\n", ojows.ClientsCount())
			secret, ok := args[0].(string)
			if !ok {
				client.Close()
				return
			}
			employee, err := helpers.ValidateToken(secret, os.Getenv(config.ACCT_PUBLIC_KEY))
			if err != nil {
				client.Close()
				return
			}

			employeeAsMap, ok := employee.(map[string]any)
			if !ok {
				client.Close()
				return
			}

			employeeObjId, err := primitive.ObjectIDFromHex(employeeAsMap["_id"].(string))
			if err != nil {
				client.Close()
				return
			}
			fullName := employeeAsMap["full_name"].(string)
			avatar := employeeAsMap["avatar"].(string)
			job := employeeAsMap["job"].(map[string]any)["name"].(string)
			client.Set("id", employeeObjId)
			client.Set("full_name", fullName)
			client.Set("avatar", avatar)
			client.Set("job", job)

			client.Join("gepciler")

			employeesChatClient := employeesChatManager.NewClient(client.Id, employeeObjId, fullName, avatar, job)

			client.Emit("clients", employeesChatManager.GetAllClientsWithoutMe(employeeObjId))
			client.Emit("messages", employeesChatManager.GetAllMessages())
			ojows.In("gepciler").Except(client.Id).Emit("client", employeesChatClient)

			client.On(ws.BeforeCloseConnection, func(args ...any) {
				ojows.In("gepciler").Except(client.Id).Emit("leave-client", employeeObjId)
			})

			client.On("close-connection", func(args ...any) {
				employeesChatManager.LeaveClient(employeeObjId)
			})

			client.On("message", func(args ...any) {
				fmt.Printf("\n***\n%v new message geldi\n***\n", client.Id)
				message, ok := args[0].(map[string]any)
				if !ok {
					// fmt.Println("On(message): message not provided")
					return
				}

				employeeId, err := client.Get("id")
				if err != nil {
					// fmt.Println("On(message): employee id not found")
					return
				}
				message_time, err := time.Parse(time.RFC3339, message["created_at"].(string))
				if err != nil {
					message_time = time.Now()
				}
				messageStruct := manager.EmployeesChatMessage{
					Id:        primitive.NewObjectID(),
					By:        employeeId.(primitive.ObjectID),
					Type:      message["type"].(string),
					Content:   message["content"].(any),
					Client:    &manager.EmployeesChatClient{},
					CreatedAt: message_time,
				}

				// fmt.Println("On(message): sending.. to gepciler room")

				employeesChatManager.NewMessage(&messageStruct)

				ojows.In("gepciler").Emit("message", messageStruct)
			})

			client.Emit("success", true)
		})
	}))

	employees.Post("/chat/upload", middlewares.DeSerializeEmployee, controllers.UploadEmployeesChatFile)

	// employees.Get("/chatroom", config.OjoWS.NewClient(func(ws *ws.OjoWS, client *ws.OjoWebsocketClient) {
	// 	client.Join("room1")

	// 	client.Emit("dine ozume!", client.Id)

	// 	ws.Emit(
	// 		"new client connected herkese",
	// 	)

	// 	client.On("private-message", func(args ...(any)) {
	// 		socketId := args[0].(string)
	// 		e := fmt.Sprintf("message-from-%v", client.Id)
	// 		payload := args[1].([]any)
	// 		ws.In(socketId).Emit(e, payload...)

	// 	})

	// 	client.On("room3-join", func(args ...(any)) {
	// 		client.Join("room3")
	// 	})
	// 	client.On("room2-join", func(args ...(any)) {
	// 		client.Join("room2")
	// 	})
	// 	client.On("close", func(args ...(any)) {
	// 		fmt.Println("hazir ocyarin..")
	// 		client.Close()
	// 	})
	// 	client.On("dinle", func(args ...(any)) {
	// 		ws.In("room1").In("room2").Except("room3").Emit(
	// 			"bisdim", "gonsym 2022-in highlander-ini alypdyr welin onema derdim yetikdi",
	// 		)
	// 	})

	// 	client.On("broadcast-et", func(args ...(any)) {
	// 		client.Broadcast.Emit("broadcast-event", client.Id, "tarapyndan..")
	// 	})

	// }))

	// employees.Get("/chathana", websocket.New(func(c *websocket.Conn) {
	// 	client := chat.NewEmployeeChatRoomClient(chatRoom, c)
	// 	<-client.Done
	// }))
	employees.Post("/:id/reasons",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER, config.ADMIN}),
		controllers.GivePermission)
	employees.Post("/:id/recruit",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER, config.ADMIN}),
		controllers.EditEmployeeInfo)
	employees.Post("/:id/dismiss",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER, config.ADMIN}),
		controllers.DismissEmployee)
	employees.Get("/:id/reasons",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER, config.ADMIN}),
		controllers.GetReasonsOfEmployee)
	employees.Post("/:id/notes",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER, config.ADMIN}),
		controllers.AddNote)
	employees.Get("/:id/notes",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER, config.ADMIN}),
		controllers.GetNotesOfEmployee)
	employees.Get("/:id/everyday_works",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER, config.ADMIN}),
		controllers.GetEverydayWorkOfEmployee)
	employees.Get("/:id/everyday_work/:workId/products",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN, config.EMPLOYEES_MANAGER}),
		controllers.GetEmployeeEverydayworkProducts)
	employees.Get("/:id/everyday_work/:workId/products/:productId",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.GetEverydayworkProductDetail)
	employees.Get("/:id/everyday_work/:workId/posts",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN, config.EMPLOYEES_MANAGER}),
		controllers.GetEmployeeEverydayworkPosts)
	employees.Get("/:id/everyday_work/:workId/posts/:postId",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.GetEverydayworkPostDetail)

	employees.Get("/:id/everyday_work/:workId/cashier_activities",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN, config.EMPLOYEES_MANAGER}),
		controllers.GetEmployeeEverydayworkCashierActivities)
	employees.Get("/:id",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN, config.EMPLOYEES_MANAGER}),
		controllers.GetEmployeeInfo)
	employees.Get("/:id/edit",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN, config.EMPLOYEES_MANAGER}),
		controllers.GetEmployeeInfoForEditing)
	employees.Put("/:id",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN, config.EMPLOYEES_MANAGER}),
		controllers.EditEmployeeInfo)
	// employees.Get("/image_man", controllers.ImageManipulation)
	// employees.Get("/by_manager", middlewares.DeSerializeEmployee,
	// 	middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER}),
	// 	controllers.GetEmployeesByManager)

	// employees.Get("/chat/:id", websocket.New(func(c *websocket.Conn) {
	// 	// c.Locals is added to the *websocket.Conn
	// 	fmt.Println(c.Locals("allowed"))  // true
	// 	fmt.Println(c.Params("id"))       // 123
	// 	fmt.Println(c.Query("v"))         // 1.0
	// 	fmt.Println(c.Cookies("session")) // ""

	// 	// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
	// 	var (
	// 		mt  int
	// 		msg []byte
	// 		err error
	// 	)
	// 	for {
	// 		if mt, msg, err = c.ReadMessage(); err != nil {
	// 			fmt.Println("read:", err, websocket.IsCloseError(err, websocket.CloseNoStatusReceived))
	// 			break
	// 		}
	// 		fmt.Printf("recv: %s\n", msg)

	// 		if err = c.WriteMessage(mt, msg); err != nil {
	// 			fmt.Println("write:", err)
	// 			break
	// 		}
	// 	}

	// }))

}
