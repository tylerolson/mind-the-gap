package gtfs

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

func ParseTripRouteMap(zipReader *zip.Reader) (map[string]string, error) {
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

	tripIDIdx, routeIDIdx, err := findTripsColumnIndexes(headers)
	if err != nil {
		return nil, err
	}

	tripRouteMap := make(map[string]string)
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

		tripRouteMap[tripID] = routeID
	}

	if len(tripRouteMap) == 0 {
		return nil, fmt.Errorf("no trip-route mappings found in trips.txt")
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

func findTripsColumnIndexes(headers []string) (int, int, error) {
	tripIDIdx := -1
	routeIDIdx := -1

	for i, header := range headers {
		switch strings.TrimSpace(header) {
		case "trip_id":
			tripIDIdx = i
		case "route_id":
			routeIDIdx = i
		}
	}

	if tripIDIdx == -1 || routeIDIdx == -1 {
		return -1, -1, fmt.Errorf("required columns trip_id/route_id missing in trips.txt")
	}

	return tripIDIdx, routeIDIdx, nil
}
