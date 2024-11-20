package checkertaskservice

import (
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CheckerTaskCheckingListManager struct {
	list    map[string]*CheckerTaskCheckingListItem
	mu      *sync.Mutex
	service *CheckerTaskService
}

func (m *CheckerTaskCheckingListManager) removeByTaskId(taskId primitive.ObjectID) {
	if m.IsTaskChecking(taskId) {
		m.mu.Lock()
		delete(m.list, taskId.Hex())
		m.mu.Unlock()
	}
}

func (m *CheckerTaskCheckingListManager) All() []*CheckerTaskCheckingListItem {
	list := []*CheckerTaskCheckingListItem{}
	for _, v := range m.list {
		list = append(list, v)
	}

	return list
}

func (m *CheckerTaskCheckingListManager) List(checkerId primitive.ObjectID) []*CheckerTaskCheckingListItem {
	list := []*CheckerTaskCheckingListItem{}

	for _, v := range m.list {
		if v.CheckerId == checkerId {
			list = append(list, v)
		}
	}

	return list
}

func (m *CheckerTaskCheckingListManager) Add(taskId primitive.ObjectID, checkerId primitive.ObjectID) {
	if m.IsTaskChecking(taskId) {
		return
	}

	item := &CheckerTaskCheckingListItem{
		TaskId:    taskId,
		CheckerId: checkerId,
		CreatedAt: time.Now(),
	}
	m.mu.Lock()
	m.list[taskId.Hex()] = item
	m.mu.Unlock()

	m.service.announcer.checking(item)
}

func (m *CheckerTaskCheckingListManager) IsTaskChecking(taskId primitive.ObjectID) bool {
	m.mu.Lock()
	_, ok := m.list[taskId.Hex()]
	m.mu.Unlock()
	return ok
}

func (m *CheckerTaskCheckingListManager) Remove(taskId primitive.ObjectID, checkerId primitive.ObjectID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	i, ok := m.list[taskId.Hex()]
	if !ok || i.CheckerId != checkerId {
		return
	}

	delete(m.list, taskId.Hex())

	m.service.announcer.stopChecking(i)
}
