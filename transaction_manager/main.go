package transactionmanager

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionManager struct {
	database       *mongo.Database
	successActions []*TransactionAction
	ctx            *context.Context
	retryCount     int
	err            error
}

func NewTransaction(ctx *context.Context, database *mongo.Database, retryCount int) *TransactionManager {
	return &TransactionManager{
		database:       database,
		successActions: []*TransactionAction{},
		ctx:            ctx,
		retryCount:     retryCount,
	}
}

func (t *TransactionManager) Collection(name string) *TransactionManagerCollection {
	return &TransactionManagerCollection{
		Name:       name,
		collection: t.database.Collection(name),
		manager:    t,
	}
}

func (t *TransactionManager) Err() error {
	return t.err
}

func (t *TransactionManager) setErr(err error) {
	t.err = err
}

func (t *TransactionManager) newAction(action *TransactionAction) {
	t.successActions = append(t.successActions, action)
}

func (t *TransactionManager) Rollback() error {
	fmt.Printf("\nTransaction Rollback Started...\n")
	for i := len(t.successActions) - 1; i >= 0; i-- {
		t.successActions[i].Rollback()
	}

	t.successActions = []*TransactionAction{}
	fmt.Printf("\nTransaction Rollback Ended...\n")
	return t.err
}
