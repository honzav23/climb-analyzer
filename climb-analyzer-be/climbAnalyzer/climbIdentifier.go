package climbAnalyzer

import (
	"math"
	"sort"

	"climb-analyzer-be/types"
)

const (
	climbResampleDistance       = 10.0
	climbSmoothingRadius        = 25.0
	climbDetectionWindow        = 150.0
	climbStartGradient          = 2.0
	climbMaxDescent             = 25.0
	climbMaxDistanceWithoutGain = 250.0
	climbEndGradient            = 0.2
	climbSegmentLength          = 200.0
	climbMinLastSegmentLength   = 50.0
	climbDisplayedFlatGradient  = 0.05
)

type climbPoint struct {
	distance        float64
	latitude        float64
	longitude       float64
	rawElevation    float64
	smoothElevation float64
}

// IdentifyClimbs finds climbs without changing the response DTO. Each GPX track
// segment is processed independently so a gap in a recording cannot become a
// climb. Elevation is resampled by distance and smoothed before gradients are
// calculated, which makes detection much less dependent on recording frequency.
func IdentifyClimbs(gpxItems []types.GpxItem) []types.Climb {
	if len(gpxItems) < 2 {
		return []types.Climb{}
	}

	climbs := make([]types.Climb, 0)
	distanceOffset := 0.0

	for segmentStart := 0; segmentStart < len(gpxItems); {
		segmentEnd := segmentStart + 1
		for segmentEnd < len(gpxItems) && gpxItems[segmentEnd].SegmentID == gpxItems[segmentStart].SegmentID {
			segmentEnd++
		}

		points, segmentLength := prepareClimbPoints(gpxItems[segmentStart:segmentEnd], distanceOffset)
		climbs = append(climbs, detectClimbs(points)...)
		distanceOffset += segmentLength
		segmentStart = segmentEnd
	}

	return trimFlatClimbEnds(climbs)
}

// trimFlatClimbEnds removes trailing segments whose gradient would be rendered
// as 0.0%. It then synchronizes each affected climb's measurements, endpoint,
// profile segments, and coordinates with the shortened route section.
func trimFlatClimbEnds(climbs []types.Climb) []types.Climb {
	for i := range climbs {
		segments := climbs[i].ClimbSegments
		lastNonFlat := len(segments) - 1
		for lastNonFlat >= 0 && math.Abs(segments[lastNonFlat].AverageGradient) < climbDisplayedFlatGradient {
			lastNonFlat--
		}
		if lastNonFlat < 0 || lastNonFlat == len(segments)-1 {
			continue
		}

		segments = segments[:lastNonFlat+1]
		firstProfile := segments[0].ElevationProfile
		lastProfile := segments[lastNonFlat].ElevationProfile
		if len(firstProfile) == 0 || len(lastProfile) == 0 {
			continue
		}

		lastPoint := lastProfile[len(lastProfile)-1]
		elevationGain := lastPoint.Elevation - firstProfile[0].Elevation
		climbs[i].ClimbSegments = segments
		climbs[i].Length = lastPoint.Distance
		climbs[i].End = climbs[i].Start + lastPoint.Distance
		climbs[i].ElevationGain = elevationGain
		climbs[i].AverageGradient = calculateGradientPercent(float64(elevationGain), climbs[i].Length)
		climbs[i].ClimbCoordinates = mergeClimbSegmentCoordinates(segments)
	}
	return climbs
}

// mergeClimbSegmentCoordinates rebuilds a climb route from its retained
// segments while omitting the duplicated boundary point shared by neighbours.
func mergeClimbSegmentCoordinates(segments []types.ClimbSegment) []types.PointCoordinates {
	coordinates := make([]types.PointCoordinates, 0)
	for i, segment := range segments {
		segmentCoordinates := segment.SegmentCoordinates
		if i > 0 && len(segmentCoordinates) > 0 {
			segmentCoordinates = segmentCoordinates[1:]
		}
		coordinates = append(coordinates, segmentCoordinates...)
	}
	return coordinates
}

// prepareClimbPoints produces regularly spaced samples and interpolates missing
// elevations. It returns the original horizontal segment length even when the
// segment has insufficient elevation data.
func prepareClimbPoints(items []types.GpxItem, distanceOffset float64) ([]climbPoint, float64) {
	if len(items) < 2 {
		return nil, 0
	}

	distances := make([]float64, len(items))
	elevations := make([]float64, len(items))
	validElevationCount := 0

	for i := range items {
		if i > 0 {
			distance := items[i-1].Point.Distance2D(&items[i].Point)
			if !math.IsNaN(distance) && !math.IsInf(distance, 0) && distance > 0 {
				distances[i] = distances[i-1] + distance
			} else {
				distances[i] = distances[i-1]
			}
		}

		elevations[i] = math.NaN()
		if items[i].Elevation.NotNull() {
			elevation := items[i].Elevation.Value()
			if !math.IsNaN(elevation) && !math.IsInf(elevation, 0) {
				elevations[i] = elevation
				validElevationCount++
			}
		}
	}

	segmentLength := distances[len(distances)-1]
	if segmentLength == 0 || validElevationCount == 0 {
		return nil, segmentLength
	}

	interpolateMissingElevations(elevations, distances)
	points := resampleClimbPoints(items, distances, elevations, distanceOffset)
	smoothClimbElevations(points)
	return points, segmentLength
}

// interpolateMissingElevations fills absent elevation samples in place. Missing
// values between known samples are linearly interpolated by travelled distance;
// leading and trailing gaps use the nearest known elevation.
func interpolateMissingElevations(elevations, distances []float64) {
	firstValid := -1
	for i, elevation := range elevations {
		if !math.IsNaN(elevation) {
			firstValid = i
			break
		}
	}
	if firstValid == -1 {
		return
	}

	for i := 0; i < firstValid; i++ {
		elevations[i] = elevations[firstValid]
	}

	previousValid := firstValid
	for i := firstValid + 1; i < len(elevations); i++ {
		if math.IsNaN(elevations[i]) {
			continue
		}

		distanceSpan := distances[i] - distances[previousValid]
		for missing := previousValid + 1; missing < i; missing++ {
			ratio := 0.0
			if distanceSpan > 0 {
				ratio = (distances[missing] - distances[previousValid]) / distanceSpan
			}
			elevations[missing] = elevations[previousValid] + ratio*(elevations[i]-elevations[previousValid])
		}
		previousValid = i
	}

	for i := previousValid + 1; i < len(elevations); i++ {
		elevations[i] = elevations[previousValid]
	}
}

// resampleClimbPoints converts irregularly spaced GPX points into samples at a
// fixed distance interval. Coordinates and elevations are linearly interpolated,
// and distanceOffset keeps positions relative to the complete route.
func resampleClimbPoints(items []types.GpxItem, distances, elevations []float64, distanceOffset float64) []climbPoint {
	segmentLength := distances[len(distances)-1]
	targets := make([]float64, 0, int(segmentLength/climbResampleDistance)+2)
	for target := 0.0; target < segmentLength; target += climbResampleDistance {
		targets = append(targets, target)
	}
	if len(targets) == 0 || targets[len(targets)-1] != segmentLength {
		targets = append(targets, segmentLength)
	}

	points := make([]climbPoint, 0, len(targets))
	right := 1
	for _, target := range targets {
		for right < len(distances)-1 && distances[right] < target {
			right++
		}
		left := right - 1
		span := distances[right] - distances[left]
		ratio := 0.0
		if span > 0 {
			ratio = (target - distances[left]) / span
		}

		elevation := elevations[left] + ratio*(elevations[right]-elevations[left])
		points = append(points, climbPoint{
			distance:     distanceOffset + target,
			latitude:     items[left].Point.Latitude + ratio*(items[right].Point.Latitude-items[left].Point.Latitude),
			longitude:    items[left].Point.Longitude + ratio*(items[right].Point.Longitude-items[left].Point.Longitude),
			rawElevation: elevation,
		})
	}
	return points
}

// smoothClimbElevations reduces short GPS elevation spikes in two passes. A
// distance-based median filter removes outliers, then a moving average produces
// the elevation used for climb boundaries and gradient calculations.
func smoothClimbElevations(points []climbPoint) {
	if len(points) == 0 {
		return
	}

	medianElevations := make([]float64, len(points))
	window := make([]float64, 0, 7)
	left := 0
	right := 0
	for i := range points {
		for left < i && points[i].distance-points[left].distance > climbSmoothingRadius {
			left++
		}
		if right < i {
			right = i
		}
		for right+1 < len(points) && points[right+1].distance-points[i].distance <= climbSmoothingRadius {
			right++
		}

		window = window[:0]
		for j := left; j <= right; j++ {
			window = append(window, points[j].rawElevation)
		}
		sort.Float64s(window)
		middle := len(window) / 2
		if len(window)%2 == 0 {
			medianElevations[i] = (window[middle-1] + window[middle]) / 2
		} else {
			medianElevations[i] = window[middle]
		}
	}

	left = 0
	right = 0
	for i := range points {
		for left < i && points[i].distance-points[left].distance > climbSmoothingRadius {
			left++
		}
		if right < i {
			right = i
		}
		for right+1 < len(points) && points[right+1].distance-points[i].distance <= climbSmoothingRadius {
			right++
		}

		sum := 0.0
		for j := left; j <= right; j++ {
			sum += medianElevations[j]
		}
		points[i].smoothElevation = sum / float64(right-left+1)
	}
}

// detectClimbs scans one prepared GPX segment using a rolling gradient window.
// It starts a candidate after sustained climbing, tracks its highest point, and
// finishes it after a meaningful descent or a sufficiently long loss of gain.
func detectClimbs(points []climbPoint) []types.Climb {
	if len(points) < 2 {
		return nil
	}

	climbs := make([]types.Climb, 0)
	active := false
	climbStart := -1
	climbPeak := -1
	windowStart := 0

	for i := 1; i < len(points); i++ {
		for windowStart+1 < i && points[i].distance-points[windowStart+1].distance >= climbDetectionWindow {
			windowStart++
		}
		windowLength := points[i].distance - points[windowStart].distance
		windowGradient := calculateGradientPercent(points[i].smoothElevation-points[windowStart].smoothElevation, windowLength)

		if !active {
			if windowLength >= climbDetectionWindow && windowGradient >= climbStartGradient {
				active = true
				climbStart = lowestElevationIndex(points, windowStart, i)
				climbPeak = highestElevationIndex(points, climbStart, i)
			}
			continue
		}

		if points[i].smoothElevation >= points[climbPeak].smoothElevation {
			climbPeak = i
		}

		descentFromPeak := points[climbPeak].smoothElevation - points[i].smoothElevation
		distanceWithoutGain := points[i].distance - points[climbPeak].distance
		if descentFromPeak >= climbMaxDescent ||
			(distanceWithoutGain >= climbMaxDistanceWithoutGain && windowGradient <= climbEndGradient) {
			if climb, valid := buildClimb(points, climbStart, climbPeak); valid {
				climbs = append(climbs, climb)
			}
			active = false
			climbStart = -1
			climbPeak = -1
		}
	}

	if active {
		if climb, valid := buildClimb(points, climbStart, climbPeak); valid {
			climbs = append(climbs, climb)
		}
	}
	return climbs
}

// lowestElevationIndex returns the position of the lowest smoothed elevation in
// the inclusive range. It refines a detected candidate to the bottom of a climb.
func lowestElevationIndex(points []climbPoint, start, end int) int {
	lowest := start
	for i := start + 1; i <= end; i++ {
		if points[i].smoothElevation < points[lowest].smoothElevation {
			lowest = i
		}
	}
	return lowest
}

// highestElevationIndex returns the last position of the highest smoothed
// elevation in the inclusive range, allowing plateaus to end at their far edge.
func highestElevationIndex(points []climbPoint, start, end int) int {
	highest := start
	for i := start + 1; i <= end; i++ {
		if points[i].smoothElevation >= points[highest].smoothElevation {
			highest = i
		}
	}
	return highest
}

// buildClimb converts a candidate range into the existing Climb DTO. Its length,
// elevation gain, and average gradient use the smoothed endpoints; invalid
// candidates are rejected before profiles and coordinates are allocated.
func buildClimb(points []climbPoint, start, end int) (types.Climb, bool) {
	if start < 0 || end <= start || end >= len(points) {
		return types.Climb{}, false
	}

	length := points[end].distance - points[start].distance
	elevationGain := points[end].smoothElevation - points[start].smoothElevation
	climb := types.Climb{
		ElevationGain:   int(math.Round(elevationGain)),
		Length:          length,
		AverageGradient: calculateGradientPercent(elevationGain, length),
		Start:           points[start].distance,
		End:             points[end].distance,
	}
	if !climb.IsValidClimb() {
		return types.Climb{}, false
	}

	climb.ClimbSegments = buildClimbSegments(points, start, end)
	climb.ClimbCoordinates = make([]types.PointCoordinates, 0, end-start+1)
	for i := start; i <= end; i++ {
		climb.ClimbCoordinates = append(climb.ClimbCoordinates, types.PointCoordinates{
			Latitude:  points[i].latitude,
			Longitude: points[i].longitude,
		})
	}
	return climb, true
}

// buildClimbSegments divides a climb into stable distance-based sections for the
// coloured elevation chart. A very short remainder is merged into the preceding
// section, and all profile distances remain relative to the climb start.
func buildClimbSegments(points []climbPoint, climbStart, climbEnd int) []types.ClimbSegment {
	segments := make([]types.ClimbSegment, 0, int((points[climbEnd].distance-points[climbStart].distance)/climbSegmentLength)+1)
	segmentStart := climbStart

	for i := climbStart + 1; i <= climbEnd; i++ {
		segmentLength := points[i].distance - points[segmentStart].distance
		remainingLength := points[climbEnd].distance - points[i].distance
		shouldSplit := i == climbEnd || (segmentLength >= climbSegmentLength && remainingLength >= climbMinLastSegmentLength)
		if !shouldSplit {
			continue
		}

		elevationChange := points[i].smoothElevation - points[segmentStart].smoothElevation
		profile := make([]types.ElevationProfilePlotData, 0, i-segmentStart+1)
		coordinates := make([]types.PointCoordinates, 0, i-segmentStart+1)
		for j := segmentStart; j <= i; j++ {
			profile = append(profile, types.ElevationProfilePlotData{
				Distance:  points[j].distance - points[climbStart].distance,
				Elevation: int(math.Round(points[j].smoothElevation)),
			})
			coordinates = append(coordinates, types.PointCoordinates{
				Latitude:  points[j].latitude,
				Longitude: points[j].longitude,
			})
		}

		segments = append(segments, types.ClimbSegment{
			ElevationProfile:   profile,
			AverageGradient:    calculateGradientPercent(elevationChange, segmentLength),
			SegmentLength:      segmentLength,
			SegmentCoordinates: coordinates,
		})
		segmentStart = i
	}

	return segments
}

// calculateGradientPercent converts vertical change over horizontal distance to
// a percentage grade. Non-positive distances return zero to avoid invalid math.
func calculateGradientPercent(elevationDiff, distance float64) float64 {
	if distance <= 0 {
		return 0
	}
	return elevationDiff / distance * 100
}
