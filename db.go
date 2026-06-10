package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	mongoClient *mongo.Client
	userCol     *mongo.Collection
)

func ConnectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := "mongodb://localhost:27017" // change to your URI (Atlas, Docker, etc.)

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Failed to create MongoDB client:", err)
	}

	// Ping to confirm the connection is alive
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	mongoClient = client
	userCol = client.Database("webserver_go").Collection("webserver_go") // connect to the "webserver_go" collection inside "webserver_go" database
	fmt.Println("Connected to MongoDB!")
}

// getNextID fetches the current counter and increments it atomically in MongoDB
func getNextID() (int, error) {
	counterCol := mongoClient.Database("webserver_go").Collection("counters") // separate collection to store the counter

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": "userid"}          // document that tracks the user ID counter
	update := bson.M{"$inc": bson.M{"seq": 1}} // increment the counter by 1 each time
	opts := options.FindOneAndUpdate().SetUpsert(true). // create the counter document if it doesn't exist
		SetReturnDocument(options.After) // return the updated document after increment

	var result struct {
		Seq int `bson:"seq"`
	}
	err := counterCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, err
	}
	return result.Seq, nil // return the new incremented ID
}

func DisconnectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := mongoClient.Disconnect(ctx); err != nil {
		log.Println("Error disconnecting MongoDB:", err)
	}
}