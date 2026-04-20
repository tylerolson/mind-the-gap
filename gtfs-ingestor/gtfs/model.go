package gtfs

type RouteInfo struct {
	RouteID        string
	RouteShortName string
	RouteLongName  string
	RouteColor     string
	TextColor      string
}

type TripInfo struct {
	TripID  string
	RouteID string
	ShapeID string
}

type ShapePoint struct {
	Lat      float64
	Lon      float64
	Sequence int
}

type ArrivalUpdate struct {
	TripID    string `json:"trip_id"`
	RouteID   string `json:"route_id"`
	ShapeID   string `json:"shape_id"`
	StopID    string `json:"stop_id"`
	ArrivalTS int64  `json:"arrival_ts"`
	DelaySec  int32  `json:"delay_sec"`
}
