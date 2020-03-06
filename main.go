package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"

	"go.uber.org/zap"
)

type server struct {
	nc *nats.Conn
}

type CLog struct {
	Tag           string      `json:"tag" bson:"tag"`
	Log           interface{} `json:"log" bson:"log"`
	ContainerID   string      `json:"container_id" bson:"container_id"`
	ContainerName string      `json:"container_name" bson:"container_name"`
	Source        string      `json:"source" bson:"source"`
}

var sugar *zap.SugaredLogger

func init() {

	// do something here to set environment depending on an environment variable

	rawJSON := []byte(`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["stdout", "/tmp/logs"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar = logger.Sugar()
	sugar.Debug("Init app")

}

// client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

func main() {
	var s server
	var err error
	uri := os.Getenv("NATS_URI")
	sugar.Debug(uri)
	for i := 0; i < 5; i++ {
		nc, err := nats.Connect(uri)
		if err == nil {
			s.nc = nc
			break
		}

		sugar.Debug("Waiting before connecting to NATS at:", uri)
		time.Sleep(1 * time.Second)
	}
	//defer s.nc.Close()
	if err != nil {
		sugar.Fatal("Error establishing connection to NATS:", err)
	}

	sugar.Debug("@Connected to NATS at:", s.nc.ConnectedUrl())

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://root:ykk@localhost:27017"))
	if err != nil {
		sugar.Panic("mongo.Connect() ERROR:", err)
	}
	collection := client.Database("logging").Collection("logs")

	// Subscribe
	s.nc.Subscribe("logger.>", func(m *nats.Msg) {
		print("Log comed")

		var bulkLog [][]json.RawMessage
		json.Unmarshal([]byte(string(m.Data)), &bulkLog)

		// ? Loglar
		for _, singleLog := range bulkLog {
			var logum CLog

			json.Unmarshal([]byte(string(singleLog[1])), &logum)
			var bdoc interface{}
			logstr := fmt.Sprintf("%v", logum.Log)
			err = bson.UnmarshalJSON([]byte(logstr), &bdoc)
			if err != nil {
				panic(err)
			}
			logum.Log = bdoc

			_, err := collection.InsertOne(context.TODO(), logum)

			if err != nil {
				sugar.Fatal(err)
				//sugar.Debug(res)
			}
			context.TODO().Done()
			// id := res.InsertedID
			// sugar.Debug(id)
		}

	})
	s.nc.Flush()
	if err := s.nc.LastError(); err != nil {
		sugar.Fatal(err)
	}

	runtime.Goexit()

}
