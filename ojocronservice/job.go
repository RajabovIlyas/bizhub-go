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
	Group        interface{}            `bson:"group"`
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
	if j.Status == "removed" {
		return
	}

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
	if j.Status == "removed" {
		return
	}

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

type OjoCronJobModel struct {
	runAt        time.Time
	listenerName string
	payload      map[string]interface{}
	group        interface{}
}

func (m *OjoCronJobModel) RunAt(at time.Time) *OjoCronJobModel {
	m.runAt = at
	return m
}

func (m *OjoCronJobModel) Group(g interface{}) *OjoCronJobModel {
	m.group = g
	return m
}

func (m *OjoCronJobModel) ListenerName(name string) *OjoCronJobModel {
	m.listenerName = name
	return m
}

func (m *OjoCronJobModel) Payload(p map[string]interface{}) *OjoCronJobModel {
	m.payload = p
	return m
}

func NewOjoCronJobModel() *OjoCronJobModel {
	return &OjoCronJobModel{}
}

// new job

func (s *OjoCronService) NewJob(model *OjoCronJobModel) error {
	job := &OjoCronJob{
		service:      s,
		RunAt:        model.runAt,
		ListenerName: model.listenerName,
		Payload:      model.payload,
		Status:       "active",
		RetryCount:   0,
		Group:        model.group,
	}

	err := job.save()
	return err
}
