package statisticsservice

import (
	"context"
	"errors"
	"time"

	"github.com/devzatruk/bizhubBackend/ojologger"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StatisticsService struct {
	date              time.Time
	coll              *mongo.Collection
	employeesColl     *mongo.Collection
	recentStatisticId primitive.ObjectID
	logger            ojologger.OjoLogger
	cron              *cron.Cron
	Writer            *StatisticsServiceWriter
	Reader            *StatisticsServiceReaderManager
}

func NewStatisticsService() *StatisticsService {
	service := &StatisticsService{}
	return service
}

func (s *StatisticsService) Init(coll *mongo.Collection, employeesColl *mongo.Collection) {
	s.logger = *ojologger.LoggerService.Logger("StatisticsService")

	s.Reader = &StatisticsServiceReaderManager{
		service: s,
		logger:  s.logger.Group("StatisticsServiceReaderManager"),
	}
	s.Writer = &StatisticsServiceWriter{
		service: s,
		logger:  s.logger.Group("StatisticsServiceWriter"),
		queue:   make(chan *StatisticWriterEvent, 1000),
	}

	go s.Writer.run()

	// date and collection configurations
	now := time.Now()
	s.date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	s.coll = coll
	s.employeesColl = employeesColl

	// getting default recent statistic id or active statistic id
	s.setDefaultRecentStatisticId()

	// cron job actions
	s.cron = cron.New()
	s.cron.AddFunc("@daily", func() {
		log := s.logger.Group("@Daily Cron job")
		now := time.Now()
		date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		log.Logf("daily cron job started at: %v", date)
		s.createNewStatisticDocument(date)
		log.Logf("daily cron job finished at: %v", time.Now())

	})

	s.cron.AddFunc("@hourly", func() {
		log := s.logger.Group("@Hourly Cron job")

		for i := 0; i < 3; i++ {
			currentHour := time.Now().Hour()
			withHour := s.Writer.withHourAsStringFunc(currentHour)

			oldResult := s.coll.FindOne(context.Background(), bson.M{
				"_id": s.recentStatisticId,
			})
			if err := oldResult.Err(); err != nil {
				log.Error(err)
				return
			}

			var oldStatisticHourly struct {
				Users   StatisticUsers   `bson:"users"`
				Sellers StatisticSellers `bson:"sellers"`
			}

			err := oldResult.Decode(&oldStatisticHourly)
			if err != nil {
				log.Error(err)
				return
			}

			_, err = s.coll.UpdateOne(context.Background(), bson.M{
				"_id": s.recentStatisticId,
			}, bson.M{
				"$set": bson.M{
					withHour("users_detail.%v"): StatisticDetailByHour[StatisticUsers]{
						Detail: StatisticUsers{
							All:     oldStatisticHourly.Users.All,
							Active:  oldStatisticHourly.Users.Active,
							Deleted: 0,
						},
						Hour: currentHour,
					},
					withHour("sellers_detail.%v"): StatisticDetailByHour[StatisticSellers]{
						Detail: StatisticSellers{
							All:     oldStatisticHourly.Sellers.All,
							Active:  oldStatisticHourly.Sellers.Active,
							Deleted: 0,
						},
						Hour: currentHour,
					},
				},
			})
			if err == nil {
				log.Log("successfuly created new hour activity object")
				return
			} else {
				log.Error(err)
			}
		}

		log.Error(errors.New("failed add hour fields to statistics"))

	})
	s.cron.Start()
}

func prepare24hourMap[K comparable, V any](f func(int) (K, V)) map[K]V {
	m := map[K]V{}
	for i := 0; i < 24; i++ {
		k, v := f(i)
		m[k] = v
	}
	return m
}

func (s *StatisticsService) createNewEmptyStatisticDocument(date time.Time) {
	log := s.logger.Group("createNewEmptyStatisticDocument()")

	statistic := Statistic{
		Date: date,
		Money: StatisticMoney{
			Total:     0,
			Deposited: 0,
			Withdrew:  0,
		},
		PublishedPosts:    0,
		PublishedProducts: 0,
		Expenses:          []StatisticExpense{},
		Users: StatisticActiveDifferenceWith[StatisticUsers]{
			Detail: StatisticUsers{
				All:     0,
				Active:  0,
				Deleted: 0,
			},
			ActiveDifference: StatisticDifference{
				Up:   0,
				Down: 0,
			},
		},
		Sellers: StatisticActiveDifferenceWith[StatisticSellers]{
			Detail: StatisticSellers{
				All:     0,
				Active:  0,
				Deleted: 0,
			},
			ActiveDifference: StatisticDifference{
				Up:   0,
				Down: 0,
			},
		},
		UsersDetail: prepare24hourMap(
			func(i int) (int, StatisticDetailByHour[StatisticUsers]) {
				return i, StatisticDetailByHour[StatisticUsers]{
					Detail: StatisticUsers{
						All:     0,
						Active:  0,
						Deleted: 0,
					},
					Hour: i,
				}
			},
		),
		SellersDetail: prepare24hourMap(
			func(i int) (int, StatisticDetailByHour[StatisticSellers]) {
				return i, StatisticDetailByHour[StatisticSellers]{
					Detail: StatisticSellers{
						All:     0,
						Active:  0,
						Deleted: 0,
					},
					Hour: i,
				}
			},
		),
	}

	r, err := s.coll.InsertOne(context.Background(), statistic)
	if err != nil {
		log.Error(err)
		return
	}

	s.recentStatisticId = r.InsertedID.(primitive.ObjectID)
	log.Logf("Date => %v | ObjectID => %v", date, s.recentStatisticId)
}

func (s *StatisticsService) createNewStatisticDocument(date time.Time) {
	log := s.logger.Group("createNewStatisticDocument()")

	oldStatisticR := s.coll.FindOne(context.Background(), bson.M{
		"_id": s.recentStatisticId,
	})
	if err := oldStatisticR.Err(); err != nil {
		log.Error(err)
		return
	}
	var oldStatistic Statistic
	err := oldStatisticR.Decode(&oldStatistic)
	if err != nil {
		log.Error(err)
		return
	}

	statistic := Statistic{
		Date: date,
		Money: StatisticMoney{
			Total:     oldStatistic.Money.Total,
			Deposited: 0,
			Withdrew:  0,
		},
		PublishedPosts:    0,
		PublishedProducts: 0,
		Expenses:          []StatisticExpense{},
		Users: StatisticActiveDifferenceWith[StatisticUsers]{
			Detail: StatisticUsers{All: oldStatistic.Users.Detail.All,
				Active:  oldStatistic.Users.Detail.Active,
				Deleted: 0,
			},
			ActiveDifference: StatisticDifference{
				Up:   0,
				Down: 0,
			},
		},
		Sellers: StatisticActiveDifferenceWith[StatisticSellers]{
			Detail: StatisticSellers{
				All:     oldStatistic.Sellers.Detail.All,
				Active:  oldStatistic.Users.Detail.Active,
				Deleted: 0,
			},
			ActiveDifference: StatisticDifference{
				Up:   0,
				Down: 0,
			},
		},
		UsersDetail: map[int]StatisticDetailByHour[StatisticUsers]{
			0: {
				Detail: StatisticUsers{
					All:     oldStatistic.Users.Detail.All,
					Active:  oldStatistic.Users.Detail.Active,
					Deleted: 0,
				},
				Hour: 0,
			},
		},
		SellersDetail: map[int]StatisticDetailByHour[StatisticSellers]{
			0: {
				Detail: StatisticSellers{
					All:     oldStatistic.Sellers.Detail.All,
					Active:  oldStatistic.Sellers.Detail.Active,
					Deleted: 0,
				},
				Hour: 0,
			},
		},
	}

	r, err := s.coll.InsertOne(context.Background(), statistic)
	if err != nil {
		log.Error(err)
		return
	}

	s.recentStatisticId = r.InsertedID.(primitive.ObjectID)
	s.date = date
	log.Logf("Date => %v | ObjectID => %v", date, s.recentStatisticId)
}

func (s *StatisticsService) setDefaultRecentStatisticId() {
	log := s.logger.Group("setDefaultRecentStatisticId()")
	statisticR := s.coll.FindOne(context.Background(), bson.M{
		"date": bson.M{
			"$lte": s.date,
		},
	}, options.FindOne().SetSort(bson.M{
		"date": -1,
	}))

	if err := statisticR.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			s.createNewEmptyStatisticDocument(s.date)
			return
		} else {
			// any error
		}
		panic(err)
	}

	var statistic struct {
		Id   primitive.ObjectID `bson:"_id"`
		Date time.Time          `bson:"date"`
	}

	err := statisticR.Decode(&statistic)
	if err != nil {
		panic(err)
	}

	s.recentStatisticId = statistic.Id

	if !statistic.Date.Equal(s.date) {
		s.createNewStatisticDocument(s.date)
	}

	log.Logf("recentStatisticId: %v", s.recentStatisticId)
}
