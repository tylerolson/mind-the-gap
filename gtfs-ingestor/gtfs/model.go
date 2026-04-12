package gtfs

type ArrivalUpdate struct {
	TripID    string
	RouteID   string
	StopID    string
	ArrivalTS int64
	DelaySec  int32
}
