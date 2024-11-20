package transactionmanager

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

type TransactionActionTypeDetail struct {
	Many bool
	One  bool
}

type TransactionActionType struct {
	Insert TransactionActionTypeDetail
	Update TransactionActionTypeDetail
	Delete TransactionActionTypeDetail
}

type TransactionAction struct {
	Collection     *TransactionManagerCollection
	Type           TransactionActionType
	Filter         interface{}
	Update         interface{}
	RollbackUpdate interface{}
	RollbackUpdateWithOldData func(interface{}) bson.M
	OriginalDatas  *[]interface{}
	InsertedIDs    []interface{}
	UpdatedIDs     []interface{}
	DeletedIDs     []interface{}
}

func (a *TransactionAction) SetType(t TransactionActionType) *TransactionAction {
	a.Type = t
	return a
}

func (a *TransactionAction) SetUpdatedIDs(ids []interface{}) *TransactionAction {
	a.UpdatedIDs = ids
	return a
}

func (a *TransactionAction) SetInsertedIDs(ids []interface{}) *TransactionAction {
	a.InsertedIDs = ids
	return a
}

func (a *TransactionAction) SetFilter(filter interface{}) *TransactionAction {
	a.Filter = filter
	return a
}

func (a *TransactionAction) SetRollbackUpdate(update interface{}) *TransactionAction {
	a.RollbackUpdate = update
	return a
}

func (a *TransactionAction) SetRollbackUpdateWithOldData(update func(interface{}) bson.M) *TransactionAction {
	a.RollbackUpdateWithOldData = update
	return a
}

func (a *TransactionAction) SetUpdate(update interface{}) *TransactionAction {
	a.Update = update
	return a
}

func (a *TransactionAction) SetOriginalDatas(datas *[]interface{}) *TransactionAction {
	a.OriginalDatas = datas
	return a
}

func (a *TransactionAction) Rollback() {

	ctx := context.TODO()

	var err error

	for retry := 0; retry < a.Collection.manager.retryCount; retry++ {

		if a.Type.Insert.One {
			_, err = a.Collection.collection.DeleteOne(ctx, bson.M{
				"_id": a.InsertedIDs[0],
			})
			if err == nil {
				break
			}
		} else if a.Type.Insert.Many {
			_, err = a.Collection.collection.DeleteMany(ctx, bson.M{
				"_id": bson.M{
					"$in": a.InsertedIDs,
				},
			})
			if err == nil {
				break
			}
		} else if a.Type.Update.One {
			// REPLACE DOCUMENT ETSEK DAHA DOGRU BOLAR OYDYAN???
			updateBson:= a.RollbackUpdate
			if (a.RollbackUpdateWithOldData != nil){
				updateBson = a.RollbackUpdateWithOldData((*a.OriginalDatas)[0]);
			}

			_, err = a.Collection.collection.UpdateOne(ctx, a.Filter, updateBson)
			if err == nil {
				break
			}
		} else if a.Type.Update.Many {
			_, err = a.Collection.collection.UpdateMany(ctx, a.Filter, a.RollbackUpdate)
			if err == nil {
				break
			}
		} else if a.Type.Delete.One {
			data := (*a.OriginalDatas)[0]
			_, err = a.Collection.collection.InsertOne(ctx, data)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		a.Collection.manager.setErr(err)
	}
}
