package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/proximax-storage/go-xpx-chain-sdk/sdk"
)

var (
	start   int
	end     int
	baseUrl string
	client  *sdk.Client
)

type Block struct {
	Height    sdk.Height
	Timestamp time.Time
}

func init() {
	startHeightPtr := flag.Int("start", 0, "Start block height")
	endHeightPtr := flag.Int("end", 0, "End block height (inclusive)")
	urlPtr := flag.String("url", "", "Sirius Chain REST Server URL")
	flag.Parse()

	if *startHeightPtr == 0 || *endHeightPtr == 0 || *urlPtr == "" {
		log.Fatal("Missing required flags")
	}

	if *startHeightPtr >= *endHeightPtr {
		log.Fatal("Make sure 'start' is smaller than 'end'")
	}

	start = *startHeightPtr
	end = *endHeightPtr
	baseUrl = *urlPtr
}

func main() {
	conf, err := sdk.NewConfig(context.Background(), []string{baseUrl})
	if err != nil {
		panic(err)
	}

	client = sdk.NewClient(nil, conf)

	filename := "block_timestamp_diff.csv"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating CSV file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Height", "Time Diff (Seconds)"}
	if err := writer.Write(header); err != nil {
		fmt.Println("Error writing CSV header:", err)
		return
	}

	for i := start; i <= end; i++ {
		blockA, err := getBlock(i)
		if err != nil {
			fmt.Printf("Error retrieving block %d: %v\n", i, err)
			continue
		}

		blockB, err := getBlock(i - 1)
		if err != nil {
			fmt.Printf("Error retrieving block %d: %v\n", i-1, err)
			continue
		}
		// fmt.Println(blockA.Timestamp)
		// fmt.Println(blockB.Timestamp)

		timeDiff := blockA.Timestamp.Sub(blockB.Timestamp).Seconds()
		fmt.Printf("Between block %d and %d: %.2f (s)\n", i, i-1, timeDiff)

		record := []string{fmt.Sprintf("%d", i), fmt.Sprintf("%.2f", timeDiff)}
		if err := writer.Write(record); err != nil {
			fmt.Println("Error writing CSV record:", err)
			return
		}
	}
	fmt.Printf("Created csv file: %s\n", filename)
}

func getBlock(height int) (*Block, error) {
	info, err := client.Blockchain.GetBlockByHeight(context.Background(), sdk.Height(height))
	if err != nil {
		return nil, err
	}

	return &Block{
		Height:    info.Height,
		Timestamp: info.Timestamp.UTC(),
	}, nil
}
