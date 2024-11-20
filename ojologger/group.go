package ojologger

import (
	"fmt"
	"time"
)

type OjoLogGroup struct {
	name   string
	config *OjoLogConfig
	parent *OjoLogGroup
	logger *OjoLogger

	parentGroups []*OjoLogGroup
}

func (g *OjoLogGroup) getParentGroups() []*OjoLogGroup {
	groups := []*OjoLogGroup{}
	group := g
	for group.parent != nil {
		group = group.parent
		groups = append(groups, group)
	}

	// reverse groups slice[]
	for i, j := 0, len(groups)-1; i < j; i, j = i+1, j-1 {
		groups[i], groups[j] = groups[j], groups[i]
	}

	return groups
}

func (g *OjoLogGroup) Group(name string) *OjoLogGroup {
	return g.logger.group(name, g)
}

func (g *OjoLogGroup) SetConfig(config *OjoLogConfig) {
	g.config = config
}

func (g *OjoLogGroup) Log(str string) {
	g.logger.service.queue <- OjoLog{
		Type:      "string",
		Str:       str,
		createdAt: time.Now(),
		group:     g,
		config:    g.config,
	}
}

func (g *OjoLogGroup) Logf(format string, a ...any) {
	g.logger.service.queue <- OjoLog{
		Type:      "string",
		Str:       fmt.Sprintf(format, a...),
		createdAt: time.Now(),
		group:     g,
		config:    g.config,
	}

}

func (g *OjoLogGroup) Error(err error) {
	g.logger.service.queue <- OjoLog{
		Type:      "error",
		Error:     err,
		createdAt: time.Now(),
		group:     g,
		config:    g.config,
	}

}

func (g *OjoLogGroup) Errorf(format string, a ...any) {
	g.logger.service.queue <- OjoLog{
		Type:      "error",
		Error:     fmt.Errorf(format, a...),
		createdAt: time.Now(),
		group:     g,
		config:    g.config,
	}

}
