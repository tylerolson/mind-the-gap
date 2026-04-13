package gtfs

type RouteInfo struct {
	RouteID        string
	RouteShortName string
	RouteLongName  string
	RouteColor     string
	TextColor      string
}

type ArrivalUpdate struct {
	TripID    string
	RouteID   string
	StopID    string
	ArrivalTS int64
	DelaySec  int32
}
