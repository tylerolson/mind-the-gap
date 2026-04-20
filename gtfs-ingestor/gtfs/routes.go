package gtfs

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// LoadStaticIndexes fetches the GTFS zip from the given URL, parses trips.txt, routes.txt, and shapes.txt,
// and returns trip metadata, route metadata, and shape geometry.
// It returns an error if any step of the process fails, including fetching the zip, parsing the files, or if required data is missing.
func LoadStaticIndexes(ctx context.Context, client *http.Client, zipURL string) (map[string]TripInfo, map[string]RouteInfo, map[string][]ShapePoint, error) {
	zipBytes, err := FetchStaticGTFSZip(ctx, client, zipURL)
	if err != nil {
		return nil, nil, nil, err
	}

	if len(zipBytes) == 0 {
		return nil, nil, nil, fmt.Errorf("empty static GTFS zip payload")
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open static GTFS zip: %w", err)
	}

	tripInfoMap, err := ParseTripInfoMap(zipReader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse trips.txt: %w", err)
	}

	routeInfoMap, err := ParseRouteInfoMap(zipReader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse routes.txt: %w", err)
	}

	shapesMap, err := ParseShapesMap(zipReader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse shapes.txt: %w", err)
	}

	return tripInfoMap, routeInfoMap, shapesMap, nil
}

func ParseRouteInfoMap(zipReader *zip.Reader) (map[string]RouteInfo, error) {
	routesFile, err := findRoutesFile(zipReader.File)
	if err != nil {
		return nil, err
	}

	routesFileReader, err := routesFile.Open()
	if err != nil {
		return nil, fmt.Errorf("open routes.txt: %w", err)
	}
	defer routesFileReader.Close()

	csvReader := csv.NewReader(routesFileReader)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read routes.txt headers: %w", err)
	}

	routeIDIdx, routeShortNameIdx, routeLongNameIdx, routeColorIdx, routeTextColorIdx, err := findRoutesColumnIndexes(headers)
	if err != nil {
		return nil, err
	}

	routes := make(map[string]RouteInfo)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read routes.txt row: %w", err)
		}

		if routeIDIdx >= len(record) {
			continue
		}

		routeID := strings.TrimSpace(record[routeIDIdx])
		if routeID == "" {
			continue
		}

		routes[routeID] = RouteInfo{
			RouteID:        routeID,
			RouteShortName: getCSVValue(record, routeShortNameIdx),
			RouteLongName:  getCSVValue(record, routeLongNameIdx),
			RouteColor:     normalizeHexColor(getCSVValue(record, routeColorIdx)),
			TextColor:      normalizeHexColor(getCSVValue(record, routeTextColorIdx)),
		}
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("no route metadata found in routes.txt")
	}

	return routes, nil
}

func findRoutesFile(files []*zip.File) (*zip.File, error) {
	for _, file := range files {
		if strings.EqualFold(file.Name, "routes.txt") || strings.HasSuffix(strings.ToLower(file.Name), "/routes.txt") {
			return file, nil
		}
	}

	return nil, fmt.Errorf("routes.txt not found in static GTFS zip")
}

func findRoutesColumnIndexes(headers []string) (int, int, int, int, int, error) {
	routeIDIdx := -1
	routeShortNameIdx := -1
	routeLongNameIdx := -1
	routeColorIdx := -1
	routeTextColorIdx := -1

	for i, header := range headers {
		switch strings.TrimSpace(header) {
		case "route_id":
			routeIDIdx = i
		case "route_short_name":
			routeShortNameIdx = i
		case "route_long_name":
			routeLongNameIdx = i
		case "route_color":
			routeColorIdx = i
		case "route_text_color":
			routeTextColorIdx = i
		}
	}

	if routeIDIdx == -1 {
		return -1, -1, -1, -1, -1, fmt.Errorf("required column route_id missing in routes.txt")
	}

	return routeIDIdx, routeShortNameIdx, routeLongNameIdx, routeColorIdx, routeTextColorIdx, nil
}

func getCSVValue(record []string, idx int) string {
	if idx < 0 || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func normalizeHexColor(color string) string {
	color = strings.TrimSpace(color)
	if color == "" {
		return ""
	}
	if strings.HasPrefix(color, "#") {
		return color
	}
	return "#" + color
}
