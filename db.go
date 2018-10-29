package main

import (
	"github.com/mongodb/mongo-go-driver/mongo"
	"golang.org/x/net/context"
)

var db *mongo.Database

func Connect() (bool, error) {
	connection, err := mongo.NewClient("mongodb://kjetilh:test123@ds145463.mlab.com:45463/paraglidig")

	connection.Connect(context.Background())
	if err != nil {
		return false, err
	}

	db = connection.Database("paragliding")
	return true, err
}


func addTrack(track IgcInfo) interface{} {
	collection := db.Collection("tracks")

	res, err := collection.InsertOne(context.Background(), &track)
	if err != nil {
		return nil
	}

	return res.InsertedID
}