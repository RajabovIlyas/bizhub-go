package v1

import (
	"fmt"
	"strconv"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/mobilechatservice"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UploadChatImageFile(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("UploadChatImageFile")

	// ext := strings.Split(file.Header["Content-Type"][0], "/")[1]

	newPath, err := helpers.SaveImageFile(c, "file", "chat")

	if err != nil {
		return c.JSON(errRes("c.SaveFile()", err, config.SERVER_ERROR))
	}

	return c.JSON(models.Response[any]{
		IsSuccess: true,
		Result:    newPath,
	})
}

func GetClientRoomMessages(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("GetClientRoomMessages()")
	roomObjId, err := primitive.ObjectIDFromHex(c.Params("roomId"))
	fmt.Printf("\nroomObjId: %v\n", roomObjId)
	if err != nil {
		return c.JSON(errRes("Params(roomId)", err, config.PARAM_NOT_PROVIDED))
	}

	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)

	user := c.Locals(config.CURRENT_USER).(map[string]any)
	clientId, err := primitive.ObjectIDFromHex(user["_id"].(string))
	if err != nil {
		return c.JSON(errRes("GetClientId()", err, config.CANT_DECODE))
	}

	chatClient, err := config.MobileChatService.Client(clientId)
	if err != nil {
		return c.JSON(errRes("service.Client()", err, config.NOT_FOUND))
	}
	room, err := chatClient.Room(roomObjId)
	if err != nil {
		return c.JSON(errRes("client.Room()", err, config.NOT_FOUND))
	}

	messages, err := room.Messages(page, limit, culture)
	if err != nil {
		return c.JSON(errRes("room.Messages()", err, config.DBQUERY_ERROR))
	}

	return c.JSON(models.Response[[]mobilechatservice.MobileChatMessage]{
		IsSuccess: true,
		Result:    messages,
	})
}

func GetClientRooms(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("GetClientRooms()")
	page, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}

	user := c.Locals(config.CURRENT_USER).(map[string]any)
	clientId, err := primitive.ObjectIDFromHex(user["_id"].(string))
	if err != nil {
		return c.JSON(errRes("GetClientId()", err, config.CANT_DECODE))
	}

	chatClient, err := config.MobileChatService.Client(clientId)
	if err != nil {
		return c.JSON(errRes("service.Client()", err, config.NOT_FOUND))
	}
	rooms, err := chatClient.Rooms(page, limit)
	if err != nil {
		return c.JSON(errRes("Rooms()", err, config.DBQUERY_ERROR))
	}

	return c.JSON(models.Response[[]mobilechatservice.MobileChatClientRoom]{
		IsSuccess: true,
		Result:    rooms,
	})
}
