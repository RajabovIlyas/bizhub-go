package ojocronservice

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/devzatruk/bizhubBackend/ojologger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ongoingJobAction struct {
	Id     string
	Job    *OjoCronJob
	Action string
}

type OjoCronListener func(*OjoCronJob)

type OjoCronService struct {
	coll           *mongo.Collection
	logger         *ojologger.OjoLogger
	ongoingJobs    map[string]*OjoCronJob
	failedJobs     map[string]*OjoCronJob
	listeners      map[string]OjoCronListener
	ongoingJobChan chan *ongoingJobAction
	failedJobChan  chan *OjoCronJob
	mu             *sync.Mutex
}

func NewOjoCronService() *OjoCronService {
	service := &OjoCronService{
		logger:         ojologger.LoggerService.Logger("OjoCronService"),
		ongoingJobs:    map[string]*OjoCronJob{},
		failedJobs:     map[string]*OjoCronJob{},
		listeners:      map[string]OjoCronListener{},
		ongoingJobChan: make(chan *ongoingJobAction, 1000),
		failedJobChan:  make(chan *OjoCronJob, 1000),
		mu:             &sync.Mutex{},
	}
	return service
}

func (s *OjoCronService) Init(coll *mongo.Collection) *OjoCronService {
	s.coll = coll
	s.run()

	return s
}

func (s *OjoCronService) run() {
	go s.channelCron()
	go s.failedJobsCron()
	go s.expiredJobsCron()
}

/*
 eger `finish` ya-da `failed` function-i isletmesen sol job islanok!
*/
func (s *OjoCronService) On(listenerName string, listener OjoCronListener) {
	log := s.logger.Group("AddListener()")
	s.mu.Lock()
	s.listeners[listenerName] = listener
	s.mu.Unlock()
	log.Logf("`%v` listener added", listenerName)
}

func (s *OjoCronService) ongoingJobAction(j *OjoCronJob, action bool) {
	action_ := "add"
	if action == false {
		action_ = "remove"
	}

	j.service.ongoingJobChan <- &ongoingJobAction{
		Id:     j.Id.Hex(),
		Job:    j,
		Action: action_,
	}
}

func (s *OjoCronService) getListener(name string) (OjoCronListener, error) {
	listener, ok := s.listeners[name]
	if !ok {
		return nil, fmt.Errorf("Listener [%v] not found", name)
	}

	return listener, nil
}

func (s *OjoCronService) containsJobInOngoingList(Id string) bool {
	_, ok := s.ongoingJobs[Id]
	return ok
}

func (s *OjoCronService) getOngoingIds() bson.A {
	ids := bson.A{}

	for _, v := range s.ongoingJobs {
		ids = append(ids, v.Id)
	}

	return ids
}

func (s *OjoCronService) getExpiredJobs(date time.Time) ([]*OjoCronJob, error) {
	ctx := context.Background()

	ongoingIds := s.getOngoingIds()

	cursor, err := s.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": bson.M{
					"$nin": ongoingIds,
				},
				"run_at": bson.M{
					"$lt": date,
				},
				"status": "active",
			},
		},
		bson.M{
			"$limit": 100,
		},
	})
	if err != nil {
		return []*OjoCronJob{}, err
	}

	var jobs []*OjoCronJob

	for cursor.Next(ctx) {
		var job OjoCronJob
		err := cursor.Decode(&job)
		if err != nil {
			return []*OjoCronJob{}, err
		}

		jobs = append(jobs, &job)
	}

	if jobs == nil {
		jobs = []*OjoCronJob{}
	}

	return jobs, nil
}

func (s *OjoCronService) expiredJobsCron() {
	log := s.logger.Group("expiredJobsCron()")
	tick_every, err := time.ParseDuration(os.Getenv("OJOCRON_EXPIRES_TICKER"))
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(tick_every) // .env.CRON_TICKER
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			t := time.Now()
			log.Logf("getting expired jobs %v", t)
			jobs, err := s.getExpiredJobs(t)
			if err != nil {
				log.Error(err)
				continue
			}

			for _, job := range jobs {
				job.init(s)
				job.run()
			}
		}
	}
}

func (s *OjoCronService) failedJobsCron() {
	log := s.logger.Group("expiredJobsCron()")
	tick_every, err := time.ParseDuration(os.Getenv("OJOCRON_FAILED_TICKER"))
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(tick_every)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			if len(s.failedJobs) == 0 {
				continue
			}

			log.Log("Removing failed jobs")
			faileds := s.failedJobs

			for _, job := range faileds {
				err := job.remove()
				if err == nil {
					s.mu.Lock()
					delete(s.failedJobs, job.Id.Hex())
					s.mu.Unlock()
				}
			}

			log.Log("Removed failed jobs")

		}
	}
}

func (s *OjoCronService) channelCron() {
	log := s.logger.Group("channelCron()")
	log.Logf("Started at: %v", time.Now())

	for {
		select {
		case job, ok := <-s.failedJobChan:
			if !ok {
				continue
			}
			s.mu.Lock()
			s.failedJobs[job.Id.Hex()] = job
			s.mu.Unlock()

		case action, ok := <-s.ongoingJobChan:
			if !ok {
				continue
			}
			contains := s.containsJobInOngoingList(action.Id)
			if !contains && action.Action == "add" {
				s.mu.Lock()
				s.ongoingJobs[action.Id] = action.Job
				s.mu.Unlock()
			} else if contains && action.Action == "delete" {
				s.mu.Lock()
				delete(s.ongoingJobs, action.Id)
				s.mu.Unlock()
			}

		}
	}
}
