package ojologger

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type OjoLogger struct {
	name string
	mu   *sync.Mutex
	file *os.File
}

func New(name string) *OjoLogger {

	logger := &OjoLogger{
		name: name,
		mu:   &sync.Mutex{},
	}

	if config.UseFile && config.Enabled {
		logger.openLogFile()
	}

	return logger
}

func (l *OjoLogger) openLogFile() {
	now := time.Now()
	filename := fmt.Sprintf("%v_%v_%v.ojo.log", now.Day(), int(now.Month()), now.Year())

	folder := strings.ReplaceAll(l.name, " ", "_")

	folderPath := path.Join(config.LogFolderPath, folder)

	err := os.Mkdir(folderPath, 0777)
	if err != nil {
		if os.IsExist(err) {
			fmt.Printf("\nfound log folder of `%v` logger\n", l.name)
		} else {
			fmt.Printf("\nfailed create log folder of `%v` logger\n", l.name)
			return
		}
	}
	filenamePath := path.Join(folderPath, filename)
	file, err := os.OpenFile(filenamePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("\nfailed open log file of `%v` logger\n", l.name)
		return
	}

	b := make([]byte, 1)
	isNew, err := file.Read(b)
	fmt.Println(b)

	if isNew == 0 {
		file.WriteString("---------------------------\n")
		file.WriteString(fmt.Sprintf(" Name: %v\n", l.name))
		file.WriteString(fmt.Sprintf(" Created at: %v/%v/%v %v:%v:%v\n", now.Day(), int(now.Month()), now.Year(), now.Hour(), now.Minute(), now.Second()))
		file.WriteString("---------------------------\n")
	}

	l.file = file
}

func (l *OjoLogger) logToFile(str string) {
	if l.file == nil {
		return
	}

	l.file.WriteString(str)
}

func (l *OjoLogger) Group(name string) *OjoLogGroup {
	group := &OjoLogGroup{
		Name:   name,
		root:   l,
		parent: nil,
	}
	group.prepareParentString()

	return group
}

func (l *OjoLogger) by() string {
	by := color.New(color.FgHiWhite, color.BgBlue)
	return by.Sprintf(" %v ", l.name)
}
