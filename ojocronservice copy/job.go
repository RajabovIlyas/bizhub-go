package ojocronservice

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OjoCronJob struct {
	Id           primitive.ObjectID     `bson:"_id,omitempty"`
	RunAt        time.Time              `bson:"run_at"`
	RetryCount   int64                  `bson:"retry_count"`
	ListenerName string                 `bson:"listener"`
	Payload      map[string]interface{} `bson:"payload"`
	service      *OjoCronService        `bson:"-"`
	Status       string                 `bson:"status"`
}

func (j *OjoCronJob) init(service *OjoCronService) {
	j.service = service
	j.Status = "inited"
}

func (j *OjoCronJob) run() {
	j.service.ongoingJobAction(j, true)

	listener, err := j.service.getListener(j.ListenerName)
	if err != nil {
		j.Failed()
		return
	}

	go listener(j)
}

func (j *OjoCronJob) Failed() {
	j.Status = "failed"
	j.service.failedJobChan <- j

	j.service.coll.UpdateOne(context.Background(), bson.M{
		"_id": j.Id,
	}, bson.M{
		"$set": bson.M{
			"status": "failed",
		},
	})
}

func (j *OjoCronJob) Finish() {
	err := j.remove()
	if err != nil {
		j.Failed()
	}
}

func (j *OjoCronJob) Retry() error {
	_, err := j.service.coll.UpdateOne(context.Background(), bson.M{
		"_id": j.Id,
	}, bson.M{
		"$inc": bson.M{
			"retry_count": 1,
		},
	})
	if err != nil {
		return err
	}

	j.service.ongoingJobAction(j, false)

	return nil
}

func (j *OjoCronJob) save() error {
	log := j.service.logger.Group("cron job save()")
	log.Logf("job saving: %v", j)
	ctx := context.Background()
	_, err := j.service.coll.InsertOne(ctx, *j)

	return err
}

func (j *OjoCronJob) remove() error {
	ctx := context.Background()
	_, err := j.service.coll.DeleteOne(ctx, bson.M{
		"_id": j.Id,
	})
	return err
}

// new job model

type ojoCronJobModel struct {
	runAt        time.Time
	listenerName string
	payload      map[string]interface{}
}

func (m *ojoCronJobModel) RunAt(at time.Time) *ojoCronJobModel {
	m.runAt = at
	return m
}

func (m *ojoCronJobModel) ListenerName(name string) *ojoCronJobModel {
	m.listenerName = name
	return m
}

func (m *ojoCronJobModel) Payload(p map[string]interface{}) *ojoCronJobModel {
	m.payload = p
	return m
}

func NewOjoCronJobModel() *ojoCronJobModel {
	return &ojoCronJobModel{}
}

// new job

func (s *OjoCronService) NewJob(model *ojoCronJobModel) error {
	job := &OjoCronJob{
		service:      s,
		RunAt:        model.runAt,
		ListenerName: model.listenerName,
		Payload:      model.payload,
		Status:       "active",
		RetryCount:   0,
	}

	err := job.save()

	return err
}
