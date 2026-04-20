package gtfs

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

func ParseTripInfoMap(zipReader *zip.Reader) (map[string]TripInfo, error) {
	tripsFile, err := findTripsFile(zipReader.File)
	if err != nil {
		return nil, err
	}

	fileReader, err := tripsFile.Open()
	if err != nil {
		return nil, fmt.Errorf("open trips.txt: %w", err)
	}
	defer fileReader.Close()

	csvReader := csv.NewReader(fileReader)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read trips.txt headers: %w", err)
	}

	tripIDIdx, routeIDIdx, shapeIDIdx, err := findTripsColumnIndexes(headers)
	if err != nil {
		return nil, err
	}

	tripInfoMap := make(map[string]TripInfo)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read trips.txt row: %w", err)
		}

		if tripIDIdx >= len(record) || routeIDIdx >= len(record) {
			continue
		}

		tripID := strings.TrimSpace(record[tripIDIdx])
		routeID := strings.TrimSpace(record[routeIDIdx])
		if tripID == "" || routeID == "" {
			continue
		}

		shapeID := ""
		if shapeIDIdx >= 0 && shapeIDIdx < len(record) {
			shapeID = strings.TrimSpace(record[shapeIDIdx])
		}

		tripInfoMap[tripID] = TripInfo{
			TripID:  tripID,
			RouteID: routeID,
			ShapeID: shapeID,
		}
	}

	if len(tripInfoMap) == 0 {
		return nil, fmt.Errorf("no trip metadata found in trips.txt")
	}

	return tripInfoMap, nil
}

func ParseTripRouteMap(zipReader *zip.Reader) (map[string]string, error) {
	tripInfoMap, err := ParseTripInfoMap(zipReader)
	if err != nil {
		return nil, err
	}

	tripRouteMap := make(map[string]string)
	for tripID, info := range tripInfoMap {
		tripRouteMap[tripID] = info.RouteID
	}

	return tripRouteMap, nil
}

func findTripsFile(files []*zip.File) (*zip.File, error) {
	for _, file := range files {
		if strings.EqualFold(file.Name, "trips.txt") || strings.HasSuffix(strings.ToLower(file.Name), "/trips.txt") {
			return file, nil
		}
	}

	return nil, fmt.Errorf("trips.txt not found in static GTFS zip")
}

func findTripsColumnIndexes(headers []string) (int, int, int, error) {
	tripIDIdx := -1
	routeIDIdx := -1
	shapeIDIdx := -1

	for i, header := range headers {
		switch strings.TrimSpace(header) {
		case "trip_id":
			tripIDIdx = i
		case "route_id":
			routeIDIdx = i
		case "shape_id":
			shapeIDIdx = i
		}
	}

	if tripIDIdx == -1 || routeIDIdx == -1 {
		return -1, -1, -1, fmt.Errorf("required columns trip_id/route_id missing in trips.txt")
	}

	return tripIDIdx, routeIDIdx, shapeIDIdx, nil
}
