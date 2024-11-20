package checkertaskservice

import (
	"sync"

	"github.com/devzatruk/bizhubBackend/ws"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CheckerTaskService struct {
	Writer       *CheckerTaskServiceWriter
	announcer    *CheckerTaskServiceRealtimeAnnouncer
	CheckingList *CheckerTaskCheckingListManager
	mu           *sync.Mutex
	coll         *mongo.Collection
}

func NewCheckerTaskService() *CheckerTaskService {
	service := &CheckerTaskService{
		mu: &sync.Mutex{},
	}
	service.CheckingList = &CheckerTaskCheckingListManager{
		service: service,
		list:    map[string]*CheckerTaskCheckingListItem{},
		mu:      &sync.Mutex{},
	}

	return service
}

func (s *CheckerTaskService) Init(coll *mongo.Collection, ws *ws.OjoWS) {

	s.coll = coll
	s.announcer = &CheckerTaskServiceRealtimeAnnouncer{
		ws:      ws,
		service: s,
	}
	s.Writer = &CheckerTaskServiceWriter{
		queue:   make(chan *CheckerTaskWithRetry, configWriterQueueMaxCapacity),
		service: s,
	}

	go s.Writer.run()

	//TODO: run..
}

// TODO: eger AddNewProduct() bolsa AutoPost yayratmaly!!! Eger EditProduct() bolsa etmesin
func (s *CheckerTaskService) Confirm(taskId primitive.ObjectID, checkerId primitive.ObjectID) {
	s.CheckingList.Remove(taskId, checkerId)
	s.announcer.deleteTask(taskId)
}

// func (s *CheckerTaskService) Task(task *CheckerTask) {
// 	s.announcer.task(task)
// }

func (s *CheckerTaskService) RemoveTask(taskId primitive.ObjectID) {
	s.announcer.deleteTask(taskId)
	s.CheckingList.removeByTaskId(taskId)
}
