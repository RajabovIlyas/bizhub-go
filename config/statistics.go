package config

import "github.com/devzatruk/bizhubBackend/statisticsservice"

var (
	StatisticsService = statisticsservice.NewStatisticsService()
)