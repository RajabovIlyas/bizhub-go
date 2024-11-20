package v1

import (
	"fmt"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/devzatruk/bizhubBackend/ojologger"
	statisticsservice "github.com/devzatruk/bizhubBackend/statisticsservice"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	logger = ojologger.LoggerService.Logger("StatisticsRoutes")
)

func UsersActivity(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UsersActivity()")
	type_ := c.Query("type", "today")

	reader, err := config.StatisticsService.Reader.New(type_)
	if err != nil {
		return c.JSON(errRes("NewReader()", err, ""))
	}

	activity, err := reader.UsersActivity()
	if err != nil {
		return c.JSON(errRes("GetActivity", err, ""))
	}

	return c.JSON(models.Response[statisticsservice.StatisticReportUsersActivity]{
		IsSuccess: true,
		Result:    *activity,
	})
}

func EmployeesActivity(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("EmployeesActivity()")
	type_ := c.Query("type", "today")

	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}

	reader, err := config.StatisticsService.Reader.New(type_)
	if err != nil {
		return c.JSON(errRes("NewReader()", err, ""))
	}

	activities, err := reader.EmployeesActivity(page, limit)
	if err != nil {
		return c.JSON(errRes("GetActivity", err, ""))
	}

	a := *activities

	for i := 0; i < len(a); i++ {
		if a[i].Job.Name == config.ADMIN_CHECKER {
			a[i].Note = fmt.Sprintf("%v checked", a[i].CompletedTasksCount)
		} else if a[i].Job.Name == config.CASHIER {
			a[i].Note = fmt.Sprintf("%v requests", a[i].CompletedTasksCount)
		}
	}

	return c.JSON(models.Response[[]statisticsservice.StatisticReportEmployeeActivity]{
		IsSuccess: true,
		Result:    *activities,
	})
}

func SellersActivity(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UsersActivity()")

	type_ := c.Query("type", "today")

	reader, err := config.StatisticsService.Reader.New(type_)
	if err != nil {
		return c.JSON(errRes("NewReader()", err, ""))
	}

	activity, err := reader.SellersActivity()
	if err != nil {
		return c.JSON(errRes("GetActivity", err, ""))
	}

	return c.JSON(models.Response[statisticsservice.StatisticReportSellersActivity]{
		IsSuccess: true,
		Result:    *activity,
	})
}

func PublishedProductsAndPosts(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UsersActivity()")

	type_ := c.Query("type", "today")

	reader, err := config.StatisticsService.Reader.New(type_)
	if err != nil {
		return c.JSON(errRes("NewReader()", err, ""))
	}

	activity, err := reader.PublishedProductsAndPosts()
	if err != nil {
		return c.JSON(errRes("GetActivity", err, ""))
	}

	return c.JSON(models.Response[statisticsservice.StatisticReportPublishedProductsAndPosts]{
		IsSuccess: true,
		Result:    *activity,
	})
}

func MoneyActivity(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UsersActivity()")

	type_ := c.Query("type", "today")

	reader, err := config.StatisticsService.Reader.New(type_)
	if err != nil {
		return c.JSON(errRes("NewReader()", err, ""))
	}

	activity, err := reader.MoneyActivity()
	if err != nil {
		return c.JSON(errRes("GetActivity", err, ""))
	}

	return c.JSON(models.Response[statisticsservice.StatisticReportMoneyActivity]{
		IsSuccess: true,
		Result:    *activity,
	})
}

func ExpensesActivity(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UsersActivity()")

	type_ := c.Query("type", "today")

	reader, err := config.StatisticsService.Reader.New(type_)
	if err != nil {
		return c.JSON(errRes("NewReader()", err, ""))
	}

	activity, err := reader.ExpensesActivity()
	if err != nil {
		return c.JSON(errRes("GetActivity", err, ""))
	}

	return c.JSON(models.Response[statisticsservice.StatisticReportExpensesActivity]{
		IsSuccess: true,
		Result:    *activity,
	})
}

// detail

func UsersActivityDetail(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UsersActivity()")

	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, ""))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, ""))
	}

	reader, err := config.StatisticsService.Reader.New("today")
	if err != nil {
		return c.JSON(errRes("NewReader()", err, ""))
	}

	detail, err := reader.UsersActivityDetail(page, limit)
	if err != nil {
		return c.JSON(errRes("GetActivityDetail", err, ""))
	}

	return c.JSON(models.Response[[]statisticsservice.StatisticReportUsersAndSellersActivityDetail]{
		IsSuccess: true,
		Result:    detail,
	})
}

func SellersActivityDetail(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UsersActivity()")

	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, ""))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, ""))
	}

	reader, err := config.StatisticsService.Reader.New("today")
	if err != nil {
		return c.JSON(errRes("NewReader()", err, ""))
	}

	detail, err := reader.SellersActivityDetail(page, limit)
	if err != nil {
		return c.JSON(errRes("GetActivityDetail", err, ""))
	}

	return c.JSON(models.Response[[]statisticsservice.StatisticReportUsersAndSellersActivityDetail]{
		IsSuccess: true,
		Result:    detail,
	})
}

func PublishedProductsAndPostsDetail(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UsersActivity()")

	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, ""))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, ""))
	}

	reader, err := config.StatisticsService.Reader.New("today")
	if err != nil {
		return c.JSON(errRes("NewReader()", err, ""))
	}

	detail, err := reader.PublishedProductsAndPostsDetail(page, limit)
	if err != nil {
		return c.JSON(errRes("GetActivityDetail", err, ""))
	}

	return c.JSON(models.Response[[]statisticsservice.StatisticReportPublishedProductsAndPostsDetail]{
		IsSuccess: true,
		Result:    detail,
	})
}

func MoneyActivityDetail(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UsersActivity()")

	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, ""))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, ""))
	}

	reader, err := config.StatisticsService.Reader.New("today")
	if err != nil {
		return c.JSON(errRes("NewReader()", err, ""))
	}

	detail, err := reader.MoneyActivityDetail(page, limit)
	if err != nil {
		return c.JSON(errRes("GetActivityDetail", err, ""))
	}

	return c.JSON(models.Response[[]statisticsservice.StatisticReportMoneyActivityDetail]{
		IsSuccess: true,
		Result:    detail,
	})
}

// run

func RunStatistic(c *fiber.Ctx) error {
	log := logger.Group("RunStatistic()")

	startedAt := time.Now()

	var body struct {
		Type    string         `json:"type"`
		Payload map[string]any `json:"payload"`
	}
	err := c.BodyParser(&body)
	if err != nil {
		return c.JSON(models.ErrorResponse(err))
	}
	log.Logf("body: %v", body)

	switch body.Type {
	case "new_user":
		config.StatisticsService.Writer.NewUser()
	case "new_active_user":
		config.StatisticsService.Writer.NewActiveUser()
	case "new_inactive_user":
		config.StatisticsService.Writer.NewInactiveUser()
	case "new_deleted_user":
		config.StatisticsService.Writer.NewDeletedUser()
	case "new_seller":
		config.StatisticsService.Writer.NewSeller()
	case "new_active_seller":
		config.StatisticsService.Writer.NewActiveSeller()
	case "new_inactive_seller":
		config.StatisticsService.Writer.NewInactiveSeller()
	case "new_deleted_seller":
		config.StatisticsService.Writer.NewDeletedSeller()
	case "published_post":
		config.StatisticsService.Writer.NewPublishedPost()
	case "published_product":
		config.StatisticsService.Writer.NewPublishedProduct()
	case "money_deposited":
		config.StatisticsService.Writer.MoneyDeposited(body.Payload["amount"].(float64))
	case "money_withdrew":
		config.StatisticsService.Writer.MoneyWithdrew(body.Payload["amount"].(float64))
	case "new_expense":
		log.Log("ay isledimow :)")
		var exp statisticsservice.StatisticExpense
		err := mapstructure.Decode(body.Payload, &exp)
		if err != nil {
			return c.JSON(models.ErrorResponse(err))
		}
		exp.Date = time.Now()
		config.StatisticsService.Writer.NewExpense(exp)
	case "remove_expense":
		expObjId, err := primitive.ObjectIDFromHex(body.Payload["expense_id"].(string))
		if err != nil {
			return c.JSON(models.ErrorResponse(err))
		}
		config.StatisticsService.Writer.RemoveExpense(expObjId)
	}

	finishedAt := time.Now()

	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    finishedAt.Sub(startedAt).String(),
	})
}
