package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gtfs-ingestor/gtfs"
)

const (
	BARTTripUpdatesURL = "https://api.bart.gov/gtfsrt/tripupdate.aspx"
	BARTStaticGTFSURL  = "https://www.bart.gov/dev/schedules/google_transit.zip"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := gtfs.NewHTTPClient(10 * time.Second)

	tripInfoMap, routeInfoMap, shapesMap, err := gtfs.LoadStaticIndexes(ctx, client, BARTStaticGTFSURL)
	if err != nil {
		log.Fatalf("failed to load static GTFS indexes: %v", err)
	} else {
		log.Printf("loaded %d trip metadata rows, %d routes, and %d shapes", len(tripInfoMap), len(routeInfoMap), len(shapesMap))
	}

	store := gtfs.NewStore()

	apiServer := NewAPIServer("8080", store)
	go func() {
		if err := apiServer.ListenAndServe(); err != nil {
			log.Fatalf("API server error: %v", err)
		}
	}()

	interval := 10 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("starting GTFS-RT polling: interval=%s feed=%s", interval, BARTTripUpdatesURL)

	runPollCycle(ctx, client, tripInfoMap, routeInfoMap, store)

	for {
		select {
		case <-ctx.Done():
			log.Println("shutdown signal received, stopping ingestor")
			return
		case <-ticker.C:
			runPollCycle(ctx, client, tripInfoMap, routeInfoMap, store)
		}
	}
}

func runPollCycle(ctx context.Context, client *http.Client, tripInfoMap map[string]gtfs.TripInfo, routeInfoMap map[string]gtfs.RouteInfo, store *gtfs.Store) {
	body, err := gtfs.FetchTripUpdates(ctx, client, BARTTripUpdatesURL)
	if err != nil {
		log.Printf("fetch failed: %v", err)
		return
	}

	log.Printf("fetched %d bytes", len(body))

	updates, err := gtfs.DecodeArrivalUpdates(body)
	if err != nil {
		log.Printf("decode failed: %v", err)
		return
	}

	enrichTripMetadata(updates, tripInfoMap)

	for _, update := range updates {
		store.UpdateArrival(&update)
	}

	printArrivalUpdates(updates, routeInfoMap)
}

func printArrivalUpdates(updates []gtfs.ArrivalUpdate, routeInfoMap map[string]gtfs.RouteInfo) {
	if len(updates) == 0 {
		fmt.Println("No arrival updates in current feed snapshot")
		return
	}

	for _, update := range updates {
		arrival := time.Unix(update.ArrivalTS, 0).Local().Format(time.DateTime)
		routeID := update.RouteID
		if routeID == "" {
			routeID = "UNKNOWN"
		}
		routeDisplay := routeID
		if routeInfo, ok := routeInfoMap[update.RouteID]; ok {
			if routeInfo.RouteShortName != "" {
				routeDisplay = routeInfo.RouteShortName
			} else if routeInfo.RouteLongName != "" {
				routeDisplay = routeInfo.RouteLongName
			}
		}

		fmt.Printf("Route: %s\n", routeDisplay)
		fmt.Printf("RouteID: %s\n", routeID)
		if update.ShapeID != "" {
			fmt.Printf("ShapeID: %s\n", update.ShapeID)
		}
		fmt.Printf("Trip: %s\n", update.TripID)
		fmt.Printf("Stop: %s\n", update.StopID)
		fmt.Printf("Arrival: %s\n", arrival)
		fmt.Printf("Delay: %+ds\n", update.DelaySec)
		if routeInfo, ok := routeInfoMap[update.RouteID]; ok && routeInfo.RouteColor != "" {
			fmt.Printf("RouteColor: %s\n", routeInfo.RouteColor)
		}
		fmt.Println("-----------------------")
	}
}

// enrichTripMetadata fills in missing RouteID and ShapeID fields in the arrival updates using the provided trip metadata map.
func enrichTripMetadata(updates []gtfs.ArrivalUpdate, tripInfoMap map[string]gtfs.TripInfo) {
	if tripInfoMap == nil {
		return
	}

	for i := range updates {
		tripInfo, ok := tripInfoMap[updates[i].TripID]
		if !ok {
			continue
		}

		// Fill in missing RouteID and ShapeID from the trip metadata if not already present in the update
		if updates[i].RouteID == "" {
			updates[i].RouteID = tripInfo.RouteID
		}

		if updates[i].ShapeID == "" {
			updates[i].ShapeID = tripInfo.ShapeID
		}
	}
}
