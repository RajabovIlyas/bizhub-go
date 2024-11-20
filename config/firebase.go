package config

import (
	"context"
	"fmt"
	"os"
	"path"

	firebase "firebase.google.com/go/v4"
	"github.com/devzatruk/bizhubBackend/ojologger"
	"google.golang.org/api/option"
)

func SetupFirebase() *firebase.App {
	LoadEnv()
	logger := ojologger.LoggerService.Logger("Firebase")
	log := logger.Group("setup")
	currentPath, err := os.Getwd()
	if err != nil {
		log.Errorf("current path error: %v", err)
		panic(fmt.Sprintf("firebase current path error : %v", err))
	}

	opt := option.WithCredentialsFile(path.Join(currentPath, "config", "firebase-config.json"))
	app, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		log.Errorf("NewApp() error: %v", err)
		panic(fmt.Sprintf("firebase error : %v", err.Error()))
	}

	NotificationManager.SetFirebase(app)

	return app
}

var (
	FirebaseApp = SetupFirebase()
)
