package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoInstance struct {
	Client *mongo.Client
	DB     *mongo.Database
}

var MI *MongoInstance

func ConnectDB() {
	LoadEnv()
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal("Error initializing db: ", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal("Error connecting to db: ", err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("Error cannot ping db: ", err)
	}
	fmt.Println("Connected to db...")
	MI = &MongoInstance{
		Client: client,
		DB:     client.Database(os.Getenv("DB")),
	}
}
func CloseDB() {
	err := MI.Client.Disconnect(context.Background())
	if err != nil {
		fmt.Printf("\nError in DB.Disconnect(): %v\n", err.Error())
	}
}
