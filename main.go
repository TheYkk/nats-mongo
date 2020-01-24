package main

import (
	"context"
	"encoding/json"
	"os"
	"runtime"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

type server struct {
	nc *nats.Conn
}

type CLog struct {
	Tag           string `json:"tag" bson:"tag"`
	Log           string `json:"log" bson:"log"`
	ContainerID   string `json:"container_id" bson:"container_id"`
	ContainerName string `json:"container_name" bson:"container_name"`
	Source        string `json:"source" bson:"source"`
}

type ResponseNats struct {
	Date   uint64
	Loglar CLog
}

func init() {

	// do something here to set environment depending on an environment variable
	log.SetLevel(log.TraceLevel)
	if os.Getenv("APP_ENV") == "production" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		// The TextFormatter is default, you don't actually have to do this.
		log.SetFormatter(&log.TextFormatter{})
	}

}

// client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

func main() {
	var s server
	var err error
	uri := os.Getenv("NATS_URI")
	log.Debug(uri)
	for i := 0; i < 5; i++ {
		nc, err := nats.Connect(uri)
		if err == nil {
			s.nc = nc
			break
		}

		log.Debug("Waiting before connecting to NATS at:", uri)
		time.Sleep(1 * time.Second)
	}
	//defer s.nc.Close()
	if err != nil {
		log.Fatal("Error establishing connection to NATS:", err)
	}

	log.Debug("@Connected to NATS at:", s.nc.ConnectedUrl())

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	collection := client.Database("logging").Collection("logs")

	// Subscribe
	s.nc.Subscribe("logger.>", func(m *nats.Msg) {
		//		log.Printf("%s: %s", m.Subject, m.Data)

		// insert data to mongo
		var arr []string
		_ = json.Unmarshal(m.Data, &arr)
		for _, dat := range arr {

			log.Debug(dat)
			res, err := collection.InsertOne(ctx, bson.D{})
			if err != nil {
				log.Debug(err)
			}
			id := res.InsertedID
			log.Debug(id)
		}

	})
	s.nc.Flush()
	if err := s.nc.LastError(); err != nil {
		log.Fatal(err)
	}

	runtime.Goexit()

}
