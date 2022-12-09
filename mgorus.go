package mgorus

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type hooker struct {
	c *mongo.Collection
}

type M bson.M

func NewHooker(mgoUrl, db, collection string) (*hooker, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mgoUrl))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return &hooker{c: client.Database(db).Collection(collection)}, nil
}

func NewHookerFromCollection(collection *mongo.Collection) *hooker {
	return &hooker{c: collection}
}

// func NewHookerWithAuth(mgoUrl, db, collection, user, pass string) (*hooker, error) {
// 	session, err := mgo.Dial(mgoUrl)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := session.DB(db).Login(user, pass); err != nil {
// 		return nil, fmt.Errorf("Failed to login to mongodb: %v", err)
// 	}

// 	return &hooker{c: session.DB(db).C(collection)}, nil
// }

// func NewHookerWithAuthDb(mgoUrl, authdb, db, collection, user, pass string) (*hooker, error) {
// 	session, err := mgo.Dial(mgoUrl)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := session.DB(authdb).Login(user, pass); err != nil {
// 		return nil, fmt.Errorf("Failed to login to mongodb: %v", err)
// 	}

// 	return &hooker{c: session.DB(db).C(collection)}, nil
// }

func (h *hooker) Fire(entry *logrus.Entry) error {
	data := make(logrus.Fields)
	data["Level"] = entry.Level.String()
	data["Time"] = entry.Time
	data["Message"] = entry.Message

	for k, v := range entry.Data {
		if errData, isError := v.(error); logrus.ErrorKey == k && v != nil && isError {
			data[k] = errData.Error()
		} else {
			data[k] = v
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, mgoErr := h.c.InsertOne(ctx, M(data))

	if mgoErr != nil {
		return fmt.Errorf("failed to send log entry to mongodb: %v", mgoErr)
	}

	return nil
}

func (h *hooker) Levels() []logrus.Level {
	return logrus.AllLevels
}
