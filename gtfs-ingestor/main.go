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

const BARTTripUpdatesURL = "https://api.bart.gov/gtfsrt/tripupdate.aspx"
const BARTStaticGTFSURL = "https://www.bart.gov/dev/schedules/google_transit.zip"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := gtfs.NewHTTPClient(10 * time.Second)

	tripRouteMap, routeInfoMap, err := gtfs.LoadStaticIndexes(ctx, client, BARTStaticGTFSURL)
	if err != nil {
		log.Printf("failed to load static GTFS indexes: %v", err)
		tripRouteMap = nil
		routeInfoMap = nil
	} else {
		log.Printf("loaded %d trip-route mappings and %d routes", len(tripRouteMap), len(routeInfoMap))
	}

	interval := 10 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("starting GTFS-RT polling: interval=%s feed=%s", interval, BARTTripUpdatesURL)

	runPollCycle(ctx, client, tripRouteMap, routeInfoMap)

	for {
		select {
		case <-ctx.Done():
			log.Println("shutdown signal received, stopping ingestor")
			return
		case <-ticker.C:
			runPollCycle(ctx, client, tripRouteMap, routeInfoMap)
		}
	}
}

func runPollCycle(ctx context.Context, client *http.Client, tripRouteMap map[string]string, routeInfoMap map[string]gtfs.RouteInfo) {
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

	enrichRouteIDs(updates, tripRouteMap)

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

func enrichRouteIDs(updates []gtfs.ArrivalUpdate, tripRouteMap map[string]string) {
	if tripRouteMap == nil {
		return
	}

	for i := range updates {
		if updates[i].RouteID != "" {
			continue
		}
		updates[i].RouteID = tripRouteMap[updates[i].TripID]
	}
}
