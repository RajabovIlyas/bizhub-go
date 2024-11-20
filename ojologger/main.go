package ojologger

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
)

type OjoLoggerServiceConfig struct {
	Enabled      bool
	LogToFile    bool
	LogToConsole bool
	LogsFolder   string
}

type OjoLoggerService struct {
	loggers map[string]*OjoLogger
	mu      *sync.Mutex
	queue   chan OjoLog
	config  *OjoLoggerServiceConfig
}

func NewOjoLoggerService() *OjoLoggerService {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	service := OjoLoggerService{
		loggers: map[string]*OjoLogger{},
		mu:      &sync.Mutex{},
		queue:   make(chan OjoLog, 1000),
		config: &OjoLoggerServiceConfig{
			Enabled:      true,
			LogToFile:    true,
			LogToConsole: true,
			LogsFolder:   path.Join(root, "./logs"),
		},
	}

	go service.run()

	return &service
}

func (s *OjoLoggerService) SetConfig(config *OjoLoggerServiceConfig) {
	if config.LogsFolder == "default" {
		config.LogsFolder = s.config.LogsFolder
	}
	s.config = config
}

func (s *OjoLoggerService) Logger(name string) *OjoLogger {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger_, ok := s.loggers[name]
	if ok {
		return logger_
	}

	logger := &OjoLogger{
		name:    name,
		service: s,
		groups:  map[string]*OjoLogGroup{},
		mu:      &sync.Mutex{},
	}
	logger.init()

	s.loggers[logger.name] = logger

	return logger
}

func (s *OjoLoggerService) run() {
	errorColorFmt := color.New(color.FgRed, color.BgHiWhite)
	loggerColorFmt := color.New(color.BgBlue, color.FgHiWhite, color.Italic)
	groupsColorFmt := color.New(color.BgHiYellow, color.FgBlack)

	for {
		select {
		case log, ok := <-s.queue:
			if !ok {
				continue
			}

			if s.config.Enabled == false {
				continue
			}

			groups := []*OjoLogGroup{log.group}
			groups = append(groups, log.group.parentGroups...)

			if s.config.LogToConsole == true && log.config.LogToConsole == true {
				loggerName := loggerColorFmt.Sprintf(" %v ", log.group.logger.name)
				groupsStr := ""
				for i, group := range groups {
					groupsStr = fmt.Sprintf("%v%v", groupsStr, groupsColorFmt.Sprintf(" %v ", group.name))
					if i != len(groups)-1 { //! dine sorde!!
						groupsStr = fmt.Sprintf("%v%v", groupsStr, color.New(color.FgRed, color.BgHiYellow, color.Bold).Sprint("/"))
					}
				}
				// groupsStr = groupsColorFmt.Sprint(groupsStr)

				content := log.Str
				if log.Type == "error" {
					content = errorColorFmt.Sprint(strings.ReplaceAll(log.Error.Error(), "\n", ", "))
				}

				fmt.Printf("%v%v %v\n", loggerName, groupsStr, content)
			}

			if s.config.LogToFile == true && log.config.LogToFile == true && log.group.logger.file != nil {
				// loggerName := log.group.logger.name
				groupsStr := ""
				for i, group := range groups {
					groupsStr = fmt.Sprintf("%v %v ", groupsStr, group.name)
					if i != len(groups)-1 { //! dine sorde!!
						groupsStr = fmt.Sprintf("%v/", groupsStr)
					}
				}

				content := log.Str
				if log.Type == "error" {
					content = fmt.Sprintf("❌ - %v", strings.ReplaceAll(log.Error.Error(), "\n", ", "))
				}

				y, m, d := log.createdAt.Date()
				h, min, sec := log.createdAt.Clock()
				createdAtStr := fmt.Sprintf("%v/%v/%v %v:%v:%v", zeroLeftForDateTime(d), zeroLeftForDateTime(int(m)), y, zeroLeftForDateTime(h), zeroLeftForDateTime(min), zeroLeftForDateTime(sec))
				log.group.logger.file.WriteString(fmt.Sprintf("%v |%v- %v\n", createdAtStr, groupsStr, content))
			}

		}
	}
}

func zeroLeftForDateTime(v int) string {
	v_ := strconv.Itoa(v)
	if len(v_) < 2 {
		return fmt.Sprintf("%v%v", strings.Repeat("0", 2-len(v_)), v_)
	}

	return v_
}
