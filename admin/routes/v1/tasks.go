package v1

import (
	"os"

	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/devzatruk/bizhubBackend/ojologger"
	"github.com/devzatruk/bizhubBackend/ws"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SetupAdminTaskRoutes(router fiber.Router) {
	tasks := router.Group("/tasks")
	tasks.Get("/",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN_CHECKER}),
		controllers.GetTasks)

	tasks.Post("/:id/confirm",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN_CHECKER}),
		controllers.ConfirmTask,
	)
	tasks.Post("/:id/reject",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN_CHECKER}),
		controllers.RejectTask,
	)

	tasks.Use("/realtime", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		// fmt.Printf("\n--- ws headers: %v\n", c.GetReqHeaders())
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	tasks.Get("/realtime", config.OjoWS.NewClient(func(ojows *ws.OjoWS, client *ws.OjoWebsocketClient) {
		log := ojologger.LoggerService.Logger("AdminChecker.Tasks").Group("/realtime")
		client.On("secret", func(args ...(any)) {
			secret, ok := args[0].(string)
			if !ok {
				log.Errorf("Args not provided.")
				client.Close()
				return
			}
			employee, err := helpers.ValidateToken(secret, os.Getenv(config.ACCT_PUBLIC_KEY))
			if err != nil {
				log.Errorf("ValidateToken() error: %v", err)
				client.Close()
				return
			}
			employeeAsMap, ok := employee.(map[string]any)
			if !ok {
				log.Errorf("Employee map invalid.")
				client.Close()
				return
			}
			employeeObjId, err := primitive.ObjectIDFromHex(employeeAsMap["_id"].(string))
			if err != nil {
				log.Errorf("Employee ID error: %v", err)
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
			client.Join("checkers")
			checkingList := config.CheckerTaskService.CheckingList.All()
			client.Emit("checking-list", checkingList)

			client.On("close-connection", func(args ...any) {
				list := config.CheckerTaskService.CheckingList.List(employeeObjId)
				for _, i := range list {
					config.CheckerTaskService.CheckingList.Remove(i.TaskId, employeeObjId)
				}
			})

			client.On("check-task", func(args ...(any)) {
				taskIdAsString := args[0].(string)
				taskObjId, err := primitive.ObjectIDFromHex(taskIdAsString)
				if err != nil {
					log.Errorf("Task ID error: %v", err)
					return
				}

				config.CheckerTaskService.CheckingList.Add(taskObjId, employeeObjId)
			})

			client.On("uncheck-task", func(args ...(any)) {
				taskIdAsString := args[0].(string)
				taskObjId, err := primitive.ObjectIDFromHex(taskIdAsString)
				if err != nil {
					log.Errorf("Task ID error: %v", err)
					return
				}

				config.CheckerTaskService.CheckingList.Remove(taskObjId, employeeObjId)
			})

			client.Emit("success", true)

		})
	}))

	tasks.Get("/:id",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN_CHECKER}),
		controllers.GetTaskDetail,
	)
}
