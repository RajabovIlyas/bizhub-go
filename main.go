package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/admin"
	"github.com/devzatruk/bizhubBackend/config"
	ojocronlisteners "github.com/devzatruk/bizhubBackend/config/ojocron_listeners"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	notificationmanager "github.com/devzatruk/bizhubBackend/notification_manager"
	"github.com/devzatruk/bizhubBackend/ojocronservice"
	"github.com/devzatruk/bizhubBackend/ojologger"
	"github.com/devzatruk/bizhubBackend/seeders"

	"github.com/devzatruk/bizhubBackend/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	config.LoadEnv()
	ojologger.LoggerService.SetConfig(&ojologger.OjoLoggerServiceConfig{
		Enabled:      true,
		LogToFile:    true,
		LogToConsole: true,
		LogsFolder:   "default",
	})
	is_prefork, err := strconv.ParseBool(os.Getenv("SERVER_PREFORK"))
	if err != nil {
		panic(fmt.Sprintf("Error in .env file: %v", err))
	}

	is_startup_message_disabled, err := strconv.ParseBool(os.Getenv("SERVER_DISABLE_STARTUP_MESSAGE"))
	if err != nil {
		panic(fmt.Sprintf("Error in .env file: %v", err))
	}

	app := fiber.New(
		fiber.Config{
			Prefork:               is_prefork,
			DisableStartupMessage: is_startup_message_disabled,
		},
	)
	app.Use(cors.New())
	log_file, err := os.OpenFile("./bizhub_logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer log_file.Close()
	app.Use(logger.New(logger.Config{
		Output: log_file,
		Format: "${time}-[${ip}:${port}]-[${status}:${latency}] - ${method} ${path}\n",
	}))
	// app.Use(limiter.New(
	// 	limiter.Config{
	// 		LimitReached: func(c *fiber.Ctx) error {
	// 			return c.JSON(models.ErrorResponse(config.TOO_MANY_REQUESTS))
	// 		},
	// 		Max:        20,
	// 		Expiration: 10 * time.Second,
	// 	},
	// ))
	app.Static("/cdn", "./public")
	config.ConnectDB()
	config.MobileChatService.Init(config.MI.DB, config.OjoWS)
	config.CheckerTaskService.Init(config.MI.DB.Collection("tasks"), config.OjoWS)
	config.StatisticsService.Init(config.MI.DB.Collection("statistics"), config.MI.DB.Collection("employees"))
	config.OjoCronService.Init(config.MI.DB.Collection("ojocron_jobs"))
	ojocronlisteners.AddOjoCronListeners()

	config.EverydayWorkService.Init(config.MI.DB.Collection(config.EMPLOYEES), config.MI.DB.Collection(config.EVERYDAYWORK))

	defer config.CloseDB()
	config.NotificationManager.SetDatabase(config.MI.DB)
	routes.SetupApiRoutes(app)
	admin.SetupAdminRoutes(app)
	app.Get("/links", func(c *fiber.Ctx) error {
		return c.SendFile("./public/files/html/dynamic_links.html")
	})
	app.Get("/logger/errors/please", func(c *fiber.Ctx) error {
		is_active := c.Query("active")
		if is_active == "true" {
			ojologger.LoggerService.SetConfig(&ojologger.OjoLoggerServiceConfig{
				Enabled:      true,
				LogsFolder:   "default",
				LogToFile:    true,
				LogToConsole: true,
			})
		} else if is_active == "false" {
			ojologger.LoggerService.SetConfig(&ojologger.OjoLoggerServiceConfig{
				Enabled:      false,
				LogsFolder:   "default",
				LogToFile:    false,
				LogToConsole: false,
			})

		}
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    "logger obeys you!",
		})
	})
	app.Get("/faker-seed", func(c *fiber.Ctx) error {
		seeder := seeders.NewSeeder(config.MI.DB)
		seeder.CreateCustomer(5)
		seeder.PrintCustomers()
		// seeder.CreateCity(5)
		// seeder.PrintCities()
		// n, err := seeder.SaveCities()
		// if err != nil {
		// 	fmt.Printf("\nError saving cities: %v\n", err)
		// } else {
		// 	fmt.Printf("\nSaved %v cities.\n", n)
		// }
		return c.Status(200).SendString("/faker-seed")
	})

	app.Get("/send-notification", func(c *fiber.Ctx) error {
		title := c.Query("t", "title")
		description := c.Query("d", "description")
		event := &notificationmanager.NotificationEvent{
			Title:       title,
			Description: description,
			ClientIds:   []primitive.ObjectID{},
			ClientType: notificationmanager.NotificationEventClientType{
				All: true,
			},
			RetryCount: 0,
		}

		config.NotificationManager.AddNotificationEvent(event)

		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    "success",
		})
	})
	app.Get("/add-notification-token", func(c *fiber.Ctx) error {
		errRes := helpers.ErrorResponse("add-notification-token")
		token := c.Query("token")
		if len(token) <= 0 {
			return c.JSON(errRes("Query(token)", errors.New("Token not found."), config.NOT_FOUND))
		}

		var clientId *primitive.ObjectID

		clientId_, err := primitive.ObjectIDFromHex(c.Query("client_id"))
		if err == nil {
			clientId = &clientId_
		}

		var clientType *string

		clientType_ := c.Query("client_type")
		if len(clientType_) > 0 {
			clientType = &clientType_
		}

		os := c.Query("os", "android")

		nToken := notificationmanager.NotificationToken{
			Token:      token,
			ClientId:   clientId,
			ClientType: clientType,
			OS:         os,
		}

		config.NotificationManager.SaveNotificationToken(nToken)

		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    "inserted",
		})
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		log := ojologger.LoggerService.Logger("main").Group("/test")
		log.Log("salam logger!")
		log.Errorf("Bu sacma error!")
		return c.Status(200).JSON(fiber.Map{
			"id":      os.Getpid(),
			"message": "welcome",
		})
	})
	app.Get("/timeout", func(ctx *fiber.Ctx) error {
		fmt.Printf("\n/timeout after 50milliseconds...\n")
		c, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		type Response struct {
			Data []int `json:"data"`
		}
		res := Response{}
		ch := make(chan Response)
		go func(c chan<- Response) {
			fmt.Printf("\ngoroutine: expensiveOperation started...\n")
			time.Sleep(time.Duration(time.Millisecond * 30))
			fmt.Printf("\ngoroutine: awaking after sleeping 30 milliseconds...\n")
			a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
			rand.Seed(time.Now().UnixMilli())
			rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
			fmt.Printf("\ngoroutine: shuffled array: %v\n", a)
			ch <- Response{a}
		}(ch)
		select {
		case <-c.Done():
			fmt.Printf("\nexpensiveOperation took too long. timeout!\n")
			break
		case res = <-ch:
			fmt.Printf("\nnot timeout yet! response returned from expensiveOperation.\n")
			break
		}
		return ctx.JSON(models.Response[Response]{
			IsSuccess: true,
			Result:    res,
		})
	})
	app.Get("auto-post-test", func(c *fiber.Ctx) error {
		d := int64(5)
		dType := "day"
		prodId, _ := primitive.ObjectIDFromHex("62cf9f30c48e57fb6702a74b")
		err := ojocronlisteners.ScheduleAutoPostAddDiscountTimes(d, dType, prodId)
		if err != nil {
			return c.JSON(models.Response[string]{
				IsSuccess: false,
				Result:    fmt.Sprintf("ScheduleAutoPostAddDiscountTimes()-%v", err),
			})
		}
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    "auto-post-test successfull.",
		})
	})
	app.Get("/ayyrcron", func(c *fiber.Ctx) error {
		errRes := helpers.ErrorResponse("ayyrcron()")
		a, err := strconv.Atoi(c.Query("a"))
		if err != nil {
			return c.JSON(errRes("convert()", err, ""))
		}
		b, err := strconv.Atoi(c.Query("b"))
		if err != nil {
			return c.JSON(errRes("convert()", err, ""))
		}

		// new model
		jobmodel := ojocronservice.NewOjoCronJobModel()
		jobmodel.ListenerName("ayyr").
			Payload(map[string]interface{}{
				"a": a,
				"b": b,
			}).
			RunAt(time.Now().Add(time.Second * 30))

		// new job
		err = config.OjoCronService.NewJob(jobmodel)
		if err != nil {
			return c.JSON(errRes("NewJob()", err, ""))
		}

		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    "success",
		})
	})

	app.Listen(":3000")
}
