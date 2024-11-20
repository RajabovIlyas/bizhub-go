package ojologger

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

type OjoLogGroup struct {
	Name                  string
	root                  *OjoLogger
	parent                *OjoLogGroup
	parentAsString        string
	parentAsStringForFile string
}

func (g *OjoLogGroup) name() string {
	title := color.New(color.BgHiYellow, color.FgBlack)
	return title.Sprintf(" %v ", g.Name)
}

func (g *OjoLogGroup) Group(name string) *OjoLogGroup {
	group := &OjoLogGroup{
		root:   g.root,
		parent: g,
		Name:   name,
	}

	if config.Enabled {
		group.prepareParentString()
	}

	return group
}

func (g *OjoLogGroup) prepareParentString() {
	names := []*OjoLogGroup{g}

	p_ := g

	for p_.parent != nil {
		names = append(names, p_)
		p_ = p_.parent
	}

	names_ := []string{}
	namesForFile_ := []string{}
	for i := len(names) - 1; i >= 0; i-- {
		namesForFile_ = append(namesForFile_, (*names[i]).Name)
		names_ = append(names_, (*names[i]).name())
	}

	g.parentAsString = strings.Join(names_, "")
	g.parentAsStringForFile = strings.Join(namesForFile_, " | ")
}

func (g *OjoLogGroup) prepareLog(content string, err bool) string {
	// by
	root := g.root.by() // true

	// content
	if err {
		contentP := color.New(color.FgRed)
		content = contentP.Sprint(content)
	}

	data := fmt.Sprintf(" %v \n", content) // true

	// result

	result := root + g.parentAsString + data
	return result
}

func zeroLeftForDateTime(v int) string {
	v_ := strconv.Itoa(v)
	if len(v_) < 2 {
		return fmt.Sprintf("%v%v", strings.Repeat("0", 2-len(v_)), v_)
	}

	return v_
}

func (g *OjoLogGroup) prepareLogFile(content string, err bool) string {
	// result
	now := time.Now()
	y, m, d := now.Date()
	h, min, sec := now.Hour(), now.Minute(), now.Second()
	time_ := fmt.Sprintf("%v/%v/%v %v:%v:%v", zeroLeftForDateTime(d), zeroLeftForDateTime(int(m)), y, zeroLeftForDateTime(h), zeroLeftForDateTime(min), zeroLeftForDateTime(sec))

	if err {
		content = fmt.Sprintf("❌ - %v", content)
	}

	result := fmt.Sprintf("%v %v - %v\n", g.parentAsStringForFile, time_, content)
	return result
}

func (g *OjoLogGroup) log(content string, err bool) {
	if !config.Enabled {
		return
	}

	g.root.mu.Lock()

	log := g.prepareLog(content, err)

	fmt.Print(log)
	if config.UseFile {
		logForFile := g.prepareLogFile(content, err)
		g.root.logToFile(logForFile)
	}

	g.root.mu.Unlock()
}
func (g *OjoLogGroup) Log(content string) {
	g.log(content, false)
}

func (g *OjoLogGroup) Logf(format string, a ...interface{}) {
	if !config.Enabled {
		return
	}

	g.Log(fmt.Sprintf(format, a...))
}

func (g *OjoLogGroup) Error(err error) {
	if !config.Enabled {
		return
	}
	g.log(strings.ReplaceAll(err.Error(), "\n", " "), true)
}

func (g *OjoLogGroup) Errorf(format string, a ...any) {
	if !config.Enabled {
		return
	}

	g.log(strings.ReplaceAll(fmt.Errorf(format, a...).Error(), "\n", " "), true)
}
