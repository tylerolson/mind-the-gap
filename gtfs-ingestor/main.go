package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gtfs-ingestor/gtfs"
)

func main() {
	ctx := context.Background()
	client := gtfs.NewHTTPClient(10 * time.Second)

	body, err := gtfs.FetchTripUpdates(ctx, client, gtfs.BARTTripUpdatesURL)
	if err != nil {
		log.Fatalf("fetch failed: %v", err)
	}

	fmt.Printf("Fetched %d bytes from BART trip updates feed\n", len(body))

	updates, err := gtfs.DecodeArrivalUpdates(body)
	if err != nil {
		log.Fatalf("decode failed: %v", err)
	}

	printArrivalUpdates(updates)
}

func printArrivalUpdates(updates []gtfs.ArrivalUpdate) {
	if len(updates) == 0 {
		fmt.Println("No arrival updates in current feed snapshot")
		return
	}

	for _, update := range updates {
		arrival := time.Unix(update.ArrivalTS, 0).Local().Format(time.DateTime)

		fmt.Printf("Route: %s\n", update.RouteID)
		fmt.Printf("Trip: %s\n", update.TripID)
		fmt.Printf("Stop: %s\n", update.StopID)
		fmt.Printf("Arrival: %s\n", arrival)
		fmt.Printf("Delay: %+ds\n", update.DelaySec)
		fmt.Println("-----------------------")
	}
}
