package checkertaskservice

import (
	"time"

	"github.com/devzatruk/bizhubBackend/models"
	"github.com/devzatruk/bizhubBackend/ws"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CheckerTaskCheckingListItem struct {
	TaskId    primitive.ObjectID `json:"task_id"`
	CheckerId primitive.ObjectID `json:"checker_id"`
	CreatedAt time.Time          `json:"created_at"`
}

type CheckerTaskServiceRealtimeAnnouncer struct {
	ws      *ws.OjoWS
	service *CheckerTaskService
}

func (a *CheckerTaskServiceRealtimeAnnouncer) task(t models.NewTask) {
	a.ws.In("checkers").Emit("task", t)
}

func (a *CheckerTaskServiceRealtimeAnnouncer) checking(i *CheckerTaskCheckingListItem) {
	a.ws.In("checkers").Emit("check-task", i)
}
func (a *CheckerTaskServiceRealtimeAnnouncer) stopChecking(i *CheckerTaskCheckingListItem) {
	i_ := *i
	a.ws.In("checkers").Emit("uncheck-task", i_)
}

func (a *CheckerTaskServiceRealtimeAnnouncer) deleteTask(taskId primitive.ObjectID) {
	a.ws.In("checkers").Emit("delete-task", taskId)
}
