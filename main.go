package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	dbUrl       string
	startHeight int
	endHeight   int
)

const (
	dbName   = "catapult"
	collectionName = "blocks"
)

type BlockData struct {
	Block Block `bson:"block"`
	Meta  Meta  `bson:"meta"`
	size  int
}

type Meta struct {
	NumTransactions int `bson:"numTransactions"`
}

type Block struct {
	Height int32 `bson:"height"`
}

func init() {
	startHeightPtr := flag.Int("start", 0, "Start block height (inclusive)")
	endHeightPtr := flag.Int("end", 0, "End block height (inclusive)")
	urlPtr := flag.String("url", "", "Database URL")
	flag.Parse()

	if *startHeightPtr == 0 || *endHeightPtr == 0 || *urlPtr == "" {
		log.Fatal("Missing required flags")
	}

	startHeight = *startHeightPtr
	endHeight = *endHeightPtr
	dbUrl = *urlPtr
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbUrl))
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatalf("Failed to disconnect database: %v", err)
		}
	}()

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	collection := client.Database(dbName).Collection(collectionName)
	filter := bson.M{
		"block.height": bson.M{
			"$gte": startHeight,
			"$lte": endHeight,
		},
	}

	sort := bson.D{{Key: "block.height", Value: 1}}
	opts := options.Find().SetSort(sort)

	cur, err := collection.Find(ctx, filter, opts)
	if err != nil {
		log.Fatal("Failed to fetch data from database:", err)
	}

	var blocksData []BlockData
	var maxFetchedHeight int32 // Track the highest height fetched from database

	for cur.Next(context.Background()) {
		var result bson.M
		if err := cur.Decode(&result); err != nil {
			log.Fatal(err)
		}

		bsonBytes, err := bson.Marshal(result)
		if err != nil {
			log.Fatal(err)
		}

		documentSize := len(bsonBytes)

		var blockData BlockData
		if err := cur.Decode(&blockData); err != nil {
			log.Fatal(err)
		}

		blockData.size = documentSize
		blocksData = append(blocksData, blockData)
		
		log.Printf("Fetching blocks for height %d...", blockData.Block.Height)

        if blockData.Block.Height > maxFetchedHeight {
            maxFetchedHeight = blockData.Block.Height
        }
	}

	fileName := fmt.Sprintf("blocks-%v-%v.csv", startHeight, maxFetchedHeight)
	if err := writeToCSV(blocksData, fileName); err != nil {
		log.Fatal(err)
	}

	log.Printf("Data saved: %s", fileName)
}

func writeToCSV(data []BlockData, filename string) (error) {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Height", "NumTransactions", "BlockSize (bytes)"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, d := range data {
		row := []string{fmt.Sprintf("%d", d.Block.Height), fmt.Sprintf("%d", d.Meta.NumTransactions), fmt.Sprintf("%d", d.size)}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
