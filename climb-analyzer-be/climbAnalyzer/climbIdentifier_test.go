package climbAnalyzer

import (
	"math"
	"testing"

	"climb-analyzer-be/types"

	"github.com/tkrajina/gpxgo/gpx"
)

type elevationAtDistance func(float64) float64

func syntheticRoute(length, spacing float64, elevation elevationAtDistance, segmentID int) []types.GpxItem {
	items := make([]types.GpxItem, 0, int(length/spacing)+2)
	for distance := 0.0; distance < length; distance += spacing {
		items = append(items, syntheticPoint(distance, elevation(distance), segmentID))
	}
	items = append(items, syntheticPoint(length, elevation(length), segmentID))
	return items
}

func syntheticPoint(distance, elevation float64, segmentID int) types.GpxItem {
	nullableElevation := gpx.NewNullableFloat64(elevation)
	point := gpx.GPXPoint{Point: gpx.Point{
		Latitude:  0,
		Longitude: distance / 111_319.49,
		Elevation: *nullableElevation,
	}}
	return types.GpxItem{
		Point:     point,
		Elevation: *nullableElevation,
		SegmentID: segmentID,
	}
}

func TestIdentifyClimbsFindsClimbEndingAtTrackEnd(t *testing.T) {
	items := syntheticRoute(1_500, 10, func(distance float64) float64 {
		return 100 + distance*0.05
	}, 0)

	climbs := IdentifyClimbs(items)

	if len(climbs) != 1 {
		t.Fatalf("expected one climb, got %d", len(climbs))
	}
	assertNear(t, "length", climbs[0].Length, 1_500, 20)
	assertNear(t, "elevation gain", float64(climbs[0].ElevationGain), 75, 2)
	assertNear(t, "average gradient", climbs[0].AverageGradient, 5, 0.15)
	assertNear(t, "end", climbs[0].End, 1_500, 20)
}

func TestIdentifyClimbsIgnoresElevationNoiseOnFlatRoute(t *testing.T) {
	noise := []float64{-1.2, 0.8, -0.4, 1.1, 0.2, -0.7}
	items := syntheticRoute(2_000, 10, func(distance float64) float64 {
		return 100 + noise[int(distance/10)%len(noise)]
	}, 0)

	if climbs := IdentifyClimbs(items); len(climbs) != 0 {
		t.Fatalf("expected no climbs on a noisy flat route, got %d", len(climbs))
	}
}

func TestIdentifyClimbsKeepsShortDipInsideClimb(t *testing.T) {
	items := syntheticRoute(1_500, 10, func(distance float64) float64 {
		switch {
		case distance <= 600:
			return 100 + distance*0.05
		case distance <= 700:
			return 130 - (distance-600)*0.05
		default:
			return 125 + (distance-700)*0.05
		}
	}, 0)

	climbs := IdentifyClimbs(items)

	if len(climbs) != 1 {
		t.Fatalf("expected the short dip to remain inside one climb, got %d climbs", len(climbs))
	}
	assertNear(t, "length", climbs[0].Length, 1_500, 30)
	assertNear(t, "net elevation gain", float64(climbs[0].ElevationGain), 65, 3)
	assertNear(t, "average gradient", climbs[0].AverageGradient, 65.0/15.0, 0.2)
}

func TestIdentifyClimbsSplitsOnSignificantDescent(t *testing.T) {
	items := syntheticRoute(1_900, 10, func(distance float64) float64 {
		switch {
		case distance <= 800:
			return 100 + distance*0.05
		case distance <= 1_100:
			return 140 - (distance-800)*0.1
		default:
			return 110 + (distance-1_100)*0.05
		}
	}, 0)

	climbs := IdentifyClimbs(items)

	if len(climbs) != 2 {
		t.Fatalf("expected two climbs separated by a significant descent, got %d", len(climbs))
	}
	assertNear(t, "first climb end", climbs[0].End, 800, 40)
	assertNear(t, "second climb start", climbs[1].Start, 1_100, 60)
}

func TestIdentifyClimbsIsStableAcrossSamplingIntervals(t *testing.T) {
	elevation := func(distance float64) float64 { return 80 + distance*0.045 }
	dense := IdentifyClimbs(syntheticRoute(1_600, 5, elevation, 0))
	sparse := IdentifyClimbs(syntheticRoute(1_600, 37, elevation, 0))

	if len(dense) != 1 || len(sparse) != 1 {
		t.Fatalf("expected one climb for both sampling rates, got dense=%d sparse=%d", len(dense), len(sparse))
	}
	assertNear(t, "sampling-independent length", dense[0].Length, sparse[0].Length, 20)
	assertNear(t, "sampling-independent gradient", dense[0].AverageGradient, sparse[0].AverageGradient, 0.1)
}

func TestIdentifyClimbsDoesNotBridgeTrackSegments(t *testing.T) {
	first := syntheticRoute(300, 10, func(distance float64) float64 {
		return 100 + distance*0.05
	}, 0)
	second := syntheticRoute(300, 10, func(distance float64) float64 {
		return 115 + distance*0.05
	}, 1)
	items := append(first, second...)

	if climbs := IdentifyClimbs(items); len(climbs) != 0 {
		t.Fatalf("expected short track segments to be evaluated independently, got %d climbs", len(climbs))
	}
}

func TestIdentifyClimbsHandlesEmptyInput(t *testing.T) {
	if climbs := IdentifyClimbs(nil); len(climbs) != 0 {
		t.Fatalf("expected no climbs for empty input, got %d", len(climbs))
	}
}

func TestIdentifyClimbsRemovesFlatSegmentsFromEnd(t *testing.T) {
	items := syntheticRoute(1_900, 10, func(distance float64) float64 {
		return 100 + math.Min(distance, 1_500)*0.05
	}, 0)

	climbs := IdentifyClimbs(items)

	if len(climbs) != 1 {
		t.Fatalf("expected one climb, got %d", len(climbs))
	}
	lastSegment := climbs[0].ClimbSegments[len(climbs[0].ClimbSegments)-1]
	if math.Abs(lastSegment.AverageGradient) < 0.05 {
		t.Errorf("expected trailing 0.0%% segments to be removed, got %.4f%%", lastSegment.AverageGradient)
	}
	if climbs[0].End >= 1_900 {
		t.Errorf("expected climb end to move before the flat finish, got %.2f m", climbs[0].End)
	}
	if len(climbs[0].ClimbCoordinates) == 0 {
		t.Fatal("expected trimmed climb coordinates to remain populated")
	}
}

func assertNear(t *testing.T, name string, actual, expected, tolerance float64) {
	t.Helper()
	if math.Abs(actual-expected) > tolerance {
		t.Errorf("%s: expected %.2f +/- %.2f, got %.2f", name, expected, tolerance, actual)
	}
}
