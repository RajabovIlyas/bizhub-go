package transactionmanager

import "go.mongodb.org/mongo-driver/bson"

type TransactionModel struct {
	filter                    bson.M //interface{}
	update                    bson.M // interface{}
	rollbackUpdate            bson.M // interface{}
	rollbackUpdateWithOldData func(interface{}) bson.M
	insertDocument            interface{}
	insertManyDocuments       []interface{}
	options                   interface{}
}

func NewModel() *TransactionModel {
	return &TransactionModel{
		options: []any{},
	}
}
func (t *TransactionModel) SetFilter(filter bson.M) *TransactionModel {
	t.filter = filter
	return t
}
func (t *TransactionModel) SetUpdate(update bson.M) *TransactionModel {
	t.update = update
	return t
}
func (t *TransactionModel) SetRollbackUpdate(rollbackUpdate bson.M) *TransactionModel {
	t.rollbackUpdate = rollbackUpdate
	return t
}

func (t *TransactionModel) SetRollbackUpdateWithOldData(f func(interface{}) bson.M) *TransactionModel {
	t.rollbackUpdateWithOldData = f
	return t
}

func (t *TransactionModel) SetDocument(document interface{}) *TransactionModel {
	t.insertDocument = document
	return t
}
func (t *TransactionModel) SetManyDocuments(documents []interface{}) *TransactionModel {
	t.insertManyDocuments = documents
	return t
}
func (t *TransactionModel) SetOptions(options interface{}) *TransactionModel {
	t.options = options
	return t
}
