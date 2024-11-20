package mobilechatservice

import (
	"context"
	"time"

	"github.com/devzatruk/bizhubBackend/models"
	transactionmanager "github.com/devzatruk/bizhubBackend/transaction_manager"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// type MobileChatMessage struct {
// 	Id        primitive.ObjectID
// 	Sender    *MobileChatClient
// 	Room      *MobileChatRoom
// 	Content   MobileChatMessageContent
// 	CommentOf *MobileChatMessage
// 	CreatedAt time.Time
// 	public    MobileChatMessageExported
// 	Type      string
// }

type MobileChatMessageContentProduct struct {
	Id     primitive.ObjectID `json:"_id" bson:"_id" mapstructure:"_id"`
	Text   string             `json:"text" bson:"text" mapstructure:"text"`
	Detail *models.Product    `json:"detail" bson:"detail,omitempty" mapstructure:"detail"`
}

type MobileChatMessageContent struct {
	Text      *string                          `json:"text" bson:"text" mapstructure:"text"`
	Product   *MobileChatMessageContentProduct `json:"product" bson:"product" mapstructure:"product"`
	ImagePath *string                          `json:"image_path" bson:"image_path" mapstructure:"image_path"`
}

type MobileChatMessage struct {
	Id        primitive.ObjectID          `json:"_id" bson:"_id,omitempty" mapstructure:"-"`
	Sender    primitive.ObjectID          `json:"sender" bson:"sender" mapstructure:"-"`
	Room      primitive.ObjectID          `json:"room" bson:"room" mapstructure:"room"`
	Content   MobileChatMessageContent    `json:"content" bson:"content" mapstructure:"content,squash"`
	CommentOf *CommentOfMobileChatMessage `json:"comment_of" bson:"comment_of" mapstructure:"comment_of"`
	CreatedAt time.Time                   `json:"created_at" bson:"created_at" mapstructure:"-"`
	Type      string                      `json:"type" bson:"type" mapstructure:"type"`
}

type CommentOfMobileChatMessage struct {
	Id        primitive.ObjectID       `json:"_id" bson:"_id,omitempty" mapstructure:"-"`
	Content   MobileChatMessageContent `json:"content" bson:"content" mapstructure:"content,squash"`
	CreatedAt time.Time                `json:"-" bson:"created_at" mapstructure:"-"`
	Type      string                   `json:"type" bson:"type" mapstructure:"type"`
}

func (c *CommentOfMobileChatMessage) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if c == nil {
		return bsontype.Null, []byte{}, nil
	}

	return bson.MarshalValue(c.Id)
}

type MobileChatMessageAsCommetOf struct {
	Id        primitive.ObjectID       `bson:"_id,omitempty"`
	Sender    primitive.ObjectID       `bson:"sender"`
	Room      primitive.ObjectID       `bson:"room"`
	Content   MobileChatMessageContent `bson:"content"`
	CommentOf *primitive.ObjectID      `bson:"comment_of"`
	CreatedAt time.Time                `bson:"created_at"`
	Type      string                   `bson:"type"`
}

func (m *MobileChatRoom) DeleteMessage(messageId primitive.ObjectID) error {
	ctx := context.Background()

	tran := transactionmanager.NewTransaction(&ctx, m.service.db, 3)

	model := transactionmanager.NewModel()
	model.SetFilter(bson.M{
		"_id":  messageId,
		"room": m.Id,
	})
	_, err := tran.Collection(CollRoomMessages).FindOneAndDelete(model)
	if err != nil {
		tran.Rollback()
		return err
	}

	modelU := transactionmanager.NewModel()
	modelU.SetFilter(bson.M{
		"_id": m.Id,
	})
	modelU.SetUpdate(bson.M{
		"$set": bson.M{
			"last_message": nil,
		},
	})
	modelU.SetRollbackUpdateWithOldData(func(o interface{}) bson.M {
		return bson.M{
			"$set": bson.M{
				"last_message": (o.(map[string]any))["last_message"],
			},
		}
	})
	_, err = tran.Collection(CollRooms).FindOneAndUpdate(modelU)
	if err != nil {
		tran.Rollback()
		return err
	}

	if err := tran.Err(); err != nil {
		tran.Rollback()
		return err
	}

	m.ojoRoom.Emit("delete-message", m.Id.Hex(), messageId.Hex())
	m.ojoRoom.Emit("last-message", nil)

	return nil
}
