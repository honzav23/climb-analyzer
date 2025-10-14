package types

import (
	"github.com/tkrajina/gpxgo/gpx"
)

type ElevationProfilePlotData struct {
	Distance  float64 `json:"distance"`
	Elevation int     `json:"elevation"`
}

type PointCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type GpxItem struct {
	Point     gpx.GPXPoint
	Elevation gpx.NullableFloat64
}
type TripSummary struct {
	LengthKilometers float32                    `json:"lengthKilometers"`
	ElevationGain    int                        `json:"elevationGain"`
	ElevationProfile []ElevationProfilePlotData `json:"elevationProfile"`
	TripCoordinates  []PointCoordinates         `json:"tripCoordinates"`
}

type AnalysisResponse struct {
	TripSummary TripSummary `json:"tripSummary"`
	FoundClimbs []Climb     `json:"foundClimbs"`
}

type Climb struct {
	ElevationGain   int            `json:"elevationGain"`
	Length          float64        `json:"length"`
	AverageGradient float64        `json:"averageGradient"`
	ClimbSegments   []ClimbSegment `json:"climbSegments"`
	Start           float64        `json:"start"` // Distance from the beginning of the track where the climb starts
	End             float64        `json:"end"`   // Distance from the beginning of the track where the climb ends
}

func (c Climb) IsValidClimb() bool {
	return c.Length >= 500 && c.AverageGradient >= 3.0
}

func (c Climb) IsDescendTooBig() bool {
	return c.Length > 100 || c.ElevationGain < -25
}

type ClimbSegment struct {
	ElevationProfile []ElevationProfilePlotData `json:"elevationProfile"`
	AverageGradient  float64                    `json:"averageGradient"`
	SegmentLength    float64                    `json:"segmentLength"`
}

type ClimbGradientRange struct {
	GradientFrom float64
	GradientTo   float64
}
