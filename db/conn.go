// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	. "github.com/polaris1119/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var MasterDB *mongo.Database
var mongoClient *mongo.Client
var once sync.Once

func init() {
	mongoConfig, err := ConfigFile.GetSection("mongodb")
	if err != nil {
		fmt.Println("get mongodb config error:", err)
		return
	}

	if err = initEngine(mongoConfig); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}
}

var (
	ConnectDBErr = errors.New("connect db error")
	UseDBErr     = errors.New("use db error")
)

func TestDB() error {
	mongoConfig, err := ConfigFile.GetSection("mongodb")
	if err != nil {
		fmt.Println("get mongodb config error:", err)
		return err
	}

	uri := buildURI(mongoConfig)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println("connect mongodb error:", err)
		return ConnectDBErr
	}
	defer client.Disconnect(ctx)

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		fmt.Println("ping mongodb error:", err)
		return ConnectDBErr
	}

	return Init()
}

func Init() error {
	mongoConfig, err := ConfigFile.GetSection("mongodb")
	if err != nil {
		fmt.Println("get mongodb config error:", err)
		return err
	}

	if err = initEngine(mongoConfig); err != nil {
		fmt.Println("mongodb is not open:", err)
		return err
	}

	return nil
}

func buildURI(mongoConfig map[string]string) string {
	host := mongoConfig["host"]
	port := mongoConfig["port"]
	user := mongoConfig["user"]
	password := mongoConfig["password"]

	if user != "" && password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s", user, password, host, port)
	}
	return fmt.Sprintf("mongodb://%s:%s", host, port)
}

func initEngine(mongoConfig map[string]string) error {
	uri := buildURI(mongoConfig)
	dbname := mongoConfig["dbname"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	MasterDB = mongoClient.Database(dbname)
	return nil
}

func GetCollection(name string) *mongo.Collection {
	return MasterDB.Collection(name)
}

func GetClient() *mongo.Client {
	return mongoClient
}

type counter struct {
	ID  string `bson:"_id"`
	Seq int    `bson:"seq"`
}

func NextID(collectionName string) (int, error) {
	coll := MasterDB.Collection("counters")
	filter := bson.M{"_id": collectionName}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var result counter
	err := coll.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("NextID for %s error: %w", collectionName, err)
	}
	return result.Seq, nil
}

func SetNextID(collectionName string, val int) error {
	coll := MasterDB.Collection("counters")
	filter := bson.M{"_id": collectionName}
	update := bson.M{"$set": bson.M{"seq": val}}
	opts := options.Update().SetUpsert(true)

	_, err := coll.UpdateOne(context.Background(), filter, update, opts)
	return err
}
