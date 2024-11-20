package everydayworkservice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/devzatruk/bizhubBackend/ojologger"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EverydayWorkTasks struct {
	Products          []primitive.ObjectID `json:"products" bson:"products"`
	Posts             []primitive.ObjectID `json:"posts" bson:"posts"`
	SellerProfiles    []primitive.ObjectID `json:"seller_profiles" bson:"seller_profiles"`
	Notifications     []primitive.ObjectID `json:"notifications" bson:"notifications"`
	Auctions          []primitive.ObjectID `json:"auctions" bson:"auctions"`
	CashierActivities []primitive.ObjectID `json:"cashier_activities" bson:"cashier_activities"`
}

type EverydayWork struct {
	Id                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	EmployeeId          primitive.ObjectID `json:"employee_id" bson:"employee_id"`
	Date                time.Time          `json:"date" bson:"date"`
	CompletedTasksCount int64              `json:"completed_tasks_count" bson:"completed_tasks_count"`
	Tasks               EverydayWorkTasks  `json:"tasks" bson:"tasks"`
	Note                string             `json:"note" bson:"note"`
}

type EverydayWorkWithoutTasks struct {
	Id                  primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	EmployeeId          primitive.ObjectID `json:"employee_id" bson:"employee_id"`
	Date                time.Time          `json:"date" bson:"date"`
	CompletedTasksCount int64              `json:"completed_tasks_count" bson:"completed_tasks_count"`
	Note                string             `json:"note" bson:"note"`
}

type EverydayWorkService struct {
	coll          *mongo.Collection
	employeesColl *mongo.Collection
	cron          *cron.Cron
	logger        *ojologger.OjoLogger
	employees     map[string]*EverydayWorkServiceClient
	mu            *sync.Mutex
}

func NewEverydayWorkService() *EverydayWorkService {
	service := &EverydayWorkService{
		logger:    ojologger.LoggerService.Logger("EverydayWorkService"),
		mu:        &sync.Mutex{},
		employees: map[string]*EverydayWorkServiceClient{},
	}

	return service
}

func (s *EverydayWorkService) Init(employeesColl *mongo.Collection, coll *mongo.Collection) {
	s.employeesColl = employeesColl
	s.coll = coll

	log := s.logger.Group("Init()")
	log.Log("service has started..")

	s.cron = cron.New()
	s.cron.AddFunc("@daily", s.createDailyDocuments)

	s.cron.Start()

	// init employees
	s.initEmployees()
}

func (s *EverydayWorkService) initEmployees() {
	ctx := context.Background()
	cursor, err := s.employeesColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"exited_on": nil,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	for cursor.Next(ctx) {
		var e struct {
			Id primitive.ObjectID `bson:"_id"`
		}
		err := cursor.Decode(&e)
		if err != nil {
			panic(err) // TODO: panic(production-da) ulanmalymy ya basga cozgut gerekmi?
		}

		s.Of(e.Id)
	}

	if err := cursor.Err(); err != nil {
		panic(err)
	}
}

func (s *EverydayWorkService) Of(employeeId primitive.ObjectID) *EverydayWorkServiceClient {
	s.mu.Lock()
	defer s.mu.Unlock()
	client, ok := s.employees[employeeId.Hex()]
	if !ok {
		client := &EverydayWorkServiceClient{
			employeeId: employeeId,
			service:    s,
			logger:     s.logger.Group(fmt.Sprintf("Client[%v]", employeeId.Hex())),
			queue:      make(chan *EverydayWorkTask, 1000),
		}
		client.Reader = &EverydayWorkServiceReader{
			client: client,
		}

		go client.run()

		s.employees[employeeId.Hex()] = client
		return client
	}
	return client
}

func (s *EverydayWorkService) getEmployees() ([]primitive.ObjectID, error) {
	ids := []primitive.ObjectID{}

	ctx := context.Background()
	cursor, err := s.employeesColl.Aggregate(ctx, bson.A{
		bson.M{
			"exited_on": nil,
		},
	})
	if err != nil {
		return []primitive.ObjectID{}, err
	}

	for cursor.Next(ctx) {
		var row struct {
			Id primitive.ObjectID `bson:"_id"`
		}
		err := cursor.Decode(&row)
		if err != nil {
			return []primitive.ObjectID{}, err
		}

		ids = append(ids, row.Id)
	}

	if err := cursor.Err(); err != nil {
		return []primitive.ObjectID{}, err
	}

	return ids, nil
}

func (s *EverydayWorkService) getOnlyDate(date time.Time) time.Time {
	y, m, d := date.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (s *EverydayWorkService) createDailyDocuments() {
	date := s.getOnlyDate(time.Now())

	log := s.logger.Group("createDailyDocuments()")
	ctx := context.Background()
	employees, err := s.getEmployees()
	if err != nil {
		log.Error(err)
		panic(err)
	}

	batch := bson.A{}
	employeesMap := map[primitive.ObjectID]primitive.ObjectID{}

	for _, employeeId := range employees {
		s.mu.Lock()
		client, ok := s.employees[employeeId.Hex()]
		s.mu.Unlock()
		if ok && client.date.Equal(date) {
			continue
		}

		workId := primitive.NewObjectID()
		row := EverydayWork{
			Id:         workId,
			EmployeeId: employeeId,
			Date:       date,
			Tasks: EverydayWorkTasks{
				Products:          []primitive.ObjectID{},
				Posts:             []primitive.ObjectID{},
				SellerProfiles:    []primitive.ObjectID{},
				Notifications:     []primitive.ObjectID{},
				Auctions:          []primitive.ObjectID{},
				CashierActivities: []primitive.ObjectID{},
			},
		}

		batch = append(batch, row)
		employeesMap[employeeId] = workId

	}

	_, err = s.coll.InsertMany(ctx, batch)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	// s.mu.Lock() // TODO: su yerde lock() gerekmi?
	for k, v := range employeesMap {
		s.Of(k).workId = &v
	}
	// s.mu.Unlock()
}
