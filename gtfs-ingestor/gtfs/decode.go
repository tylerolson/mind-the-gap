package gtfs

import (
	"fmt"

	gtfsrt "github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

func DecodeArrivalUpdates(payload []byte) ([]ArrivalUpdate, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty protobuf payload")
	}

	var feed gtfsrt.FeedMessage
	if err := proto.Unmarshal(payload, &feed); err != nil {
		return nil, fmt.Errorf("unmarshal gtfs-rt feed: %w", err)
	}

	updates := make([]ArrivalUpdate, 0)

	for _, entity := range feed.GetEntity() {
		tripUpdate := entity.GetTripUpdate()
		if tripUpdate == nil {
			continue
		}

		trip := tripUpdate.GetTrip()
		tripID := trip.GetTripId()
		routeID := trip.GetRouteId()

		for _, stopTimeUpdate := range tripUpdate.GetStopTimeUpdate() {
			arrivalTS, delaySec, ok := getEventTimeAndDelay(stopTimeUpdate)
			if !ok {
				continue
			}

			updates = append(updates, ArrivalUpdate{
				TripID:    tripID,
				RouteID:   routeID,
				StopID:    stopTimeUpdate.GetStopId(),
				ArrivalTS: arrivalTS,
				DelaySec:  delaySec,
			})
		}
	}

	return updates, nil
}

func getEventTimeAndDelay(stopTimeUpdate *gtfsrt.TripUpdate_StopTimeUpdate) (int64, int32, bool) {
	arrival := stopTimeUpdate.GetArrival()
	if arrival != nil && arrival.GetTime() > 0 {
		return arrival.GetTime(), arrival.GetDelay(), true
	}

	departure := stopTimeUpdate.GetDeparture()
	if departure != nil && departure.GetTime() > 0 {
		return departure.GetTime(), departure.GetDelay(), true
	}

	return 0, 0, false
}
