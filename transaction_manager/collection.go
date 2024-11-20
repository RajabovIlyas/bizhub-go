package transactionmanager

import (
	"fmt"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionManagerCollection struct {
	Name       string
	collection *mongo.Collection
	manager    *TransactionManager
}

func (c *TransactionManagerCollection) InsertOne(model *TransactionModel,

// document interface{}, opts ...*options.InsertOneOptions
) (*mongo.InsertOneResult, error) {

	insertResult, err := c.collection.InsertOne(*c.manager.ctx,
		model.insertDocument) // , (model.Options).([]*options.InsertOneOptions)...)

	if err == nil {
		action := (&TransactionAction{Collection: c}).
			SetType(TransactionActionType{Insert: TransactionActionTypeDetail{One: true}}).
			SetInsertedIDs([]interface{}{insertResult.InsertedID})
		c.manager.newAction(action)
	} else {
		c.manager.setErr(err)
	}

	return insertResult, err
}
func (c *TransactionManagerCollection) FindOneAndUpdate(
	model *TransactionModel,
	// filter interface{}, update interface{}, rollbackUpdate interface{}, opts ...*options.FindOneAndUpdateOptions
) (*mongo.SingleResult, error) {
	singleResult := c.collection.FindOneAndUpdate(*c.manager.ctx,
		model.filter, model.update)
	// (model.Options).([]*options.FindOneAndUpdateOptions)...)
	err := singleResult.Err()

	if err == nil {
		action := (&TransactionAction{Collection: c}).
			SetType(TransactionActionType{Update: TransactionActionTypeDetail{One: true}}).
			SetFilter(model.filter).
			SetUpdate(model.update).
			SetRollbackUpdateWithOldData(model.rollbackUpdateWithOldData)

		var oldDocument bson.M
		err := singleResult.Decode(&oldDocument)
		if err != nil {
			action.Rollback()
			return singleResult, err
		}

		action.SetOriginalDatas(&[]interface{}{oldDocument})

		// SetOriginalDatas(&[]singleResult{})

		// RollBack etmek ucin SetOriginalDatas() ulansak gowy bolar!!!
		c.manager.newAction(action)
	} else {
		c.manager.setErr(err)
	}
	return singleResult, err
}
func (c *TransactionManagerCollection) UpdateOne(model *TransactionModel,

// filter interface{}, update interface{}, rollbackUpdate interface{}, opts ...*options.UpdateOptions
) (*mongo.UpdateResult, error) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Printf("\nInside UpdateOne RECOVER - ERROR: %v\n", err)
		}
	}()
	updateResult, err := c.collection.UpdateOne(*c.manager.ctx, model.filter, model.update) //, (model.Options).([]*options.UpdateOptions)...)
	// if can't find any matching document, UpdateOne() doesn't return error, so
	if updateResult.MatchedCount == 0 && err == nil {
		err = errors.New("No matching document found.")
	}

	if err == nil {
		action := (&TransactionAction{Collection: c}).
			SetType(TransactionActionType{Update: TransactionActionTypeDetail{One: true}}).
			SetFilter(model.filter).
			SetUpdate(model.update).
			SetRollbackUpdate(model.rollbackUpdate)
		c.manager.newAction(action)
	} else {
		c.manager.setErr(err)
	}
	return updateResult, err
}

// Always use FindAndDeleteOne() even when you need DeleteOne() only.
func (c *TransactionManagerCollection) FindOneAndDelete(model *TransactionModel,

// filter interface{}, opts ...*options.FindOneAndDeleteOptions
) (*mongo.SingleResult, error) {

	findAndDeleteResult := c.collection.FindOneAndDelete(*c.manager.ctx,
		model.filter) // , (model.Options).([]*options.FindOneAndDeleteOptions)...)
	err := findAndDeleteResult.Err()

	var data interface{}

	if err == nil {
		err = findAndDeleteResult.Decode(&data)
	}

	if err == nil {
		action := (&TransactionAction{Collection: c}).
			SetType(TransactionActionType{Delete: TransactionActionTypeDetail{One: true}}).
			SetFilter(model.filter).
			SetOriginalDatas(&[]interface{}{data})
		c.manager.newAction(action)
	} else {
		c.manager.setErr(err)
	}
	return findAndDeleteResult, err
}

func (c *TransactionManagerCollection) UpdateMany(model *TransactionModel,

// filter interface{}, update interface{}, rollbackUpdate interface{}, opts ...*options.UpdateOptions
) (*mongo.UpdateResult, error) {
	// TODO: debug etmeli
	updateManyResult, err := c.collection.UpdateMany(*c.manager.ctx,
		model.filter, model.update) // , (model.Options).([]*options.UpdateOptions)...)
	if updateManyResult.MatchedCount == 0 && err == nil {
		err = errors.New("No matching documents found.")
	}
	if err == nil {
		action := (&TransactionAction{Collection: c}).
			SetType(TransactionActionType{Update: TransactionActionTypeDetail{Many: true}}).
			SetFilter(model.filter).
			SetUpdate(model.update).
			SetRollbackUpdate(model.rollbackUpdate)
		c.manager.newAction(action)
	} else {
		c.manager.setErr(err)
	}

	return updateManyResult, err
}

func (c *TransactionManagerCollection) InsertMany(model *TransactionModel,

// document []interface{}, opts ...*options.InsertManyOptions
) (*mongo.InsertManyResult, error) {

	fmt.Printf("\ntransaction_manager: Before InsertMany()...\n")
	result, err := c.collection.InsertMany(*c.manager.ctx,
		model.insertManyDocuments) // , (model.Options).([]*options.InsertManyOptions)...)

	fmt.Printf("\ntransaction_manager: After InsertMany()...\n")
	action := (&TransactionAction{Collection: c}).
		SetType(TransactionActionType{Insert: TransactionActionTypeDetail{Many: true}}).
		SetInsertedIDs(result.InsertedIDs)

	fmt.Printf("\ntransaction_manager: After action:= ...\n")

	if err == nil && len(result.InsertedIDs) == len(model.insertManyDocuments) {
		fmt.Printf("\ntransaction_manager: inside if .. c.manager.newAction()...\n")
		c.manager.newAction(action)
		fmt.Printf("\ntransaction_manager: after c.manager.newAction()...\n")
	} else {
		fmt.Printf("\ntransaction_manager: before action.Rollback()...\n")
		action.Rollback()
		if err == nil {
			err = fmt.Errorf("Error: SomeInsertsFailed.")
		}
		c.manager.setErr(err)
		fmt.Printf("\ntransaction_manager: after action.Rollback()...\n")
	}
	fmt.Printf("\nInsertMany error: %v\n", err) // err.Error() diysen Error: nil pointer kellani agyrdyar!!
	return result, err
}
