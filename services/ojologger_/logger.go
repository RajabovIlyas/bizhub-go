package ojologger

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron"
)

type OjoLogConfig struct {
	LogToFile    bool
	LogToConsole bool
}

type OjoLog struct {
	Type      string
	Error     error
	Str       string
	createdAt time.Time

	group  *OjoLogGroup
	config *OjoLogConfig
}

type OjoLogger struct {
	name    string
	file    *os.File
	service *OjoLoggerService
	groups  map[string]*OjoLogGroup
	date    time.Time
	mu      *sync.Mutex
	cron    *cron.Cron
}

func (l *OjoLogger) init() {
	now := time.Now()
	l.date = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if l.service.config.LogToFile == true {
		l.openLogFile()
	}

	l.cron = cron.New()
	l.cron.AddFunc("@daily", func() {
		l.openLogFile()
	})
	l.cron.Start()
}

func (l *OjoLogger) openLogFile() {
	folder := path.Join(l.service.config.LogsFolder, strings.ToLower(strings.ReplaceAll(l.name, " ", "_")))

	err := os.Mkdir(folder, 0777)
	if err != nil {
		if !os.IsExist(err) {
			return
		}
	}

	now := time.Now()
	filename := fmt.Sprintf("%v_%v_%v.ojo.log", now.Day(), int(now.Month()), now.Year())
	filenamePath := path.Join(folder, filename)

	file, err := os.OpenFile(filenamePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}

	l.file = file

	info, err := file.Stat()
	if err != nil {
		return
	}

	if info.Size() == 0 {
		now := time.Now()
		y, m, d := now.Date()
		h, min, sec := now.Clock()
		createdAt := fmt.Sprintf("%v/%v/%v %v:%v:%v", zeroLeftForDateTime(d), zeroLeftForDateTime(int(m)), y, zeroLeftForDateTime(h), zeroLeftForDateTime(min), zeroLeftForDateTime(sec))

		l.file.WriteString("**************************************\n")
		l.file.WriteString(fmt.Sprintf("  Name: %v\n", l.name))
		l.file.WriteString(fmt.Sprintf("  Created at: %v\n", createdAt))
		l.file.WriteString("**************************************\n\n")
	}

}

// any

func (l *OjoLogger) Group(name string) *OjoLogGroup {
	return l.group(name, nil)
}

func (l *OjoLogger) group(name string, parent *OjoLogGroup) *OjoLogGroup {
	group_, ok := l.groups[name]
	if ok {
		return group_
	}

	group := &OjoLogGroup{
		name:         name,
		parent:       parent,
		logger:       l,
		parentGroups: []*OjoLogGroup{},
		config: &OjoLogConfig{
			LogToFile:    true,
			LogToConsole: true,
		},
	}

	group.parentGroups = group.getParentGroups()

	l.mu.Lock()
	l.groups[group.name] = group
	l.mu.Unlock()

	return group
}
