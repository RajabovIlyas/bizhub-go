package ojocronlisteners

import (
	"context"
	"fmt"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/ojocronservice"
	"github.com/devzatruk/bizhubBackend/ojologger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HandlePermissionStarted(job *ojocronservice.OjoCronJob) {
	logger := ojologger.LoggerService.Logger("AddOjoCronListeners()")
	log := logger.Group("handlePermissionStarted()")

	employeeObjId := job.Payload["employee_id"].(primitive.ObjectID)
	reason := job.Payload["reason"]
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	employeesColl := config.MI.DB.Collection(config.EMPLOYEES)
	updateResult, err := employeesColl.UpdateOne(ctx, bson.M{"_id": employeeObjId}, bson.M{
		"$set": bson.M{
			"reason": reason,
		},
	})
	if err != nil {
		log.Errorf("UpdateOne(employee): %v - %v", err, config.CANT_UPDATE)
		job.Failed()
		return
	}
	if updateResult.MatchedCount == 0 {
		log.Errorf("UpdateOne(employee): %v - %v", fmt.Sprintf("Employee %v not found.", employeeObjId), config.NOT_FOUND)
		job.Failed()
		return
	}
	log.Logf("Employee permission started.")
	job.Finish()
}
