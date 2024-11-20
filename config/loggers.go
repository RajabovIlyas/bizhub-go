package config

import "github.com/devzatruk/bizhubBackend/ojologger"

var (
	AuthV1Logger     = ojologger.LoggerService.Logger("Auth v1")
	ProductsV1Logger = ojologger.LoggerService.Logger("Products v1")
)
