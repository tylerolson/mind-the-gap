package gtfs

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

func ParseShapesMap(zipReader *zip.Reader) (map[string][]ShapePoint, error) {
	shapesFile, err := findShapesFile(zipReader.File)
	if err != nil {
		return nil, err
	}

	shapeReader, err := shapesFile.Open()
	if err != nil {
		return nil, fmt.Errorf("open shapes.txt: %w", err)
	}
	defer shapeReader.Close()

	csvReader := csv.NewReader(shapeReader)
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read shapes.txt headers: %w", err)
	}

	shapeIDIdx, latIdx, lonIdx, seqIdx, err := findShapesColumnIndexes(headers)
	if err != nil {
		return nil, err
	}

	shapes := make(map[string][]ShapePoint)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read shapes.txt row: %w", err)
		}

		if shapeIDIdx >= len(record) || latIdx >= len(record) || lonIdx >= len(record) || seqIdx >= len(record) {
			continue
		}

		shapeID := strings.TrimSpace(record[shapeIDIdx])
		if shapeID == "" {
			continue
		}

		lat, err := strconv.ParseFloat(strings.TrimSpace(record[latIdx]), 64)
		if err != nil {
			continue
		}

		lon, err := strconv.ParseFloat(strings.TrimSpace(record[lonIdx]), 64)
		if err != nil {
			continue
		}

		seq, err := strconv.Atoi(strings.TrimSpace(record[seqIdx]))
		if err != nil {
			continue
		}

		shapes[shapeID] = append(shapes[shapeID], ShapePoint{
			Lat:      lat,
			Lon:      lon,
			Sequence: seq,
		})
	}

	if len(shapes) == 0 {
		return nil, fmt.Errorf("no shape points found in shapes.txt")
	}

	for shapeID := range shapes {
		sort.Slice(shapes[shapeID], func(i, j int) bool {
			return shapes[shapeID][i].Sequence < shapes[shapeID][j].Sequence
		})
	}

	return shapes, nil
}

func findShapesFile(files []*zip.File) (*zip.File, error) {
	for _, file := range files {
		if strings.EqualFold(file.Name, "shapes.txt") || strings.HasSuffix(strings.ToLower(file.Name), "/shapes.txt") {
			return file, nil
		}
	}

	return nil, fmt.Errorf("shapes.txt not found in static GTFS zip")
}

func findShapesColumnIndexes(headers []string) (int, int, int, int, error) {
	shapeIDIdx := -1
	latIdx := -1
	lonIdx := -1
	seqIdx := -1

	for i, header := range headers {
		switch strings.TrimSpace(header) {
		case "shape_id":
			shapeIDIdx = i
		case "shape_pt_lat":
			latIdx = i
		case "shape_pt_lon":
			lonIdx = i
		case "shape_pt_sequence":
			seqIdx = i
		}
	}

	if shapeIDIdx == -1 || latIdx == -1 || lonIdx == -1 || seqIdx == -1 {
		return -1, -1, -1, -1, fmt.Errorf("required shape columns missing in shapes.txt")
	}

	return shapeIDIdx, latIdx, lonIdx, seqIdx, nil
}
