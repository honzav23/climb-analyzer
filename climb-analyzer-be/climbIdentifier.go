package main

import (
	"climb-analyzer-be/types"
)

func calculateGradientPercent(elevationDiff, distance float64) float64 {
	if distance == 0 {
		return 0.0
	}
	return (elevationDiff / distance) * 100
}

// Finds the index of which range the gradient belongs to
func findRangeIndex(gradient float64, ranges []types.ClimbGradientRange) int {
	for index, gradientRange := range ranges {
		if index == 0 {
			if gradient < gradientRange.GradientTo {
				return index
			}
		} else if index == len(ranges)-1 {
			if gradient >= gradientRange.GradientFrom {
				return index
			}
		} else {
			if gradient >= gradientRange.GradientFrom && gradient < gradientRange.GradientTo {
				return index
			}
		}
	}
	return -1
}

// Checks if the current gradient is in the same range as the average gradient
func checkIfClimbGradientsInRange(averageGradient, currentGradient float64) bool {
	flat := types.ClimbGradientRange{GradientFrom: 0.0, GradientTo: 3.0}
	moderate := types.ClimbGradientRange{GradientFrom: 3.0, GradientTo: 7.0}
	steep := types.ClimbGradientRange{GradientFrom: 7.0, GradientTo: 10.0}
	verySteep := types.ClimbGradientRange{GradientFrom: 10.0, GradientTo: 100.0}

	gradientRanges := []types.ClimbGradientRange{flat, moderate, steep, verySteep}
	averageGradientRangeIndex := findRangeIndex(averageGradient, gradientRanges)
	currentGradientRangeIndex := findRangeIndex(currentGradient, gradientRanges)

	return averageGradientRangeIndex == currentGradientRangeIndex
}

func getDistancesAndElevationsGainsFromBeginning(gpxItems []types.GpxItem) ([]float64, []int) {
	distances := make([]float64, len(gpxItems))
	elevationGains := make([]int, len(gpxItems))
	distances[0] = 0
	elevationGains[0] = 0

	for i := 1; i < len(gpxItems); i++ {
		distances[i] = distances[i-1] + gpxItems[i-1].Point.Distance3D(&gpxItems[i].Point)
		elevationDiff := gpxItems[i].Elevation.Value() - gpxItems[i-1].Elevation.Value()
		if elevationDiff > 0 {
			elevationGains[i] = elevationGains[i-1] + int(elevationDiff)
		} else {
			elevationGains[i] = elevationGains[i-1]
		}
	}
	return distances, elevationGains
}

func IdentifyClimbs(gpxItems []types.GpxItem) []types.Climb {
	distances, elevationGains := getDistancesAndElevationsGainsFromBeginning(gpxItems)
	var (
		climbs               = []types.Climb{}
		climbStartEndIndices = [][2]int{}
		currentClimb         types.Climb
		currentDescent       types.Climb
		distance             = float64(0)
		climbStarted         = false
		descentStarted       = false
		climbStartingIndex   = -1
		climbPeakIndex       = -1
		climbEndingIndex     = -1
	)

	for i := 0; i < len(gpxItems)-1; i++ {
		currentElevation := gpxItems[i].Elevation.Value()
		nextElevation := gpxItems[i+1].Elevation.Value()

		elevationDiff := nextElevation - currentElevation
		distanceBetweenPoints := distances[i+1] - distances[i]
		if distanceBetweenPoints == 0.0 {
			continue
		}
		gradientPercent := calculateGradientPercent(elevationDiff, distanceBetweenPoints)

		// Start of a climb
		if gradientPercent > 0.0 && !climbStarted {
			climbStarted = true
			climbStartingIndex = i
			currentClimb = types.Climb{
				Start:         distance,
				ElevationGain: int(elevationDiff),
				Length:        distanceBetweenPoints,
			}

		} else if climbStarted && gradientPercent >= 0.0 {
			// Continuing a climb

			if descentStarted {
				// Add the descend to the climb
				descentStarted = false
				climbStarted = true

				if currentDescent.Length > 0 {
					currentClimb.Length += currentDescent.Length
				}
				currentDescent = types.Climb{}
			}
			currentClimb.ElevationGain += int(elevationDiff)
			currentClimb.Length += distanceBetweenPoints

		} else if climbStarted && gradientPercent < 0.0 { // Descend found
			// New descend start
			if !descentStarted {
				climbPeakIndex = i
				descentStarted = true
				currentDescent.ElevationGain += int(elevationDiff)
				currentDescent.Length += distanceBetweenPoints
			} else { // Continuing descend
				currentDescent.ElevationGain += int(elevationDiff)
				currentDescent.Length += distanceBetweenPoints
			}

			if currentDescent.IsDescendTooBig() {
				// End the climb if the descend is significant enough
				climbEndingIndex = climbPeakIndex

				finishClimb(&currentClimb, &currentDescent, &climbStarted, &descentStarted, distance)
				validateClimb(currentClimb, &climbStartEndIndices, climbStartingIndex, climbEndingIndex, &climbs, gpxItems, distances, elevationGains)
				climbStartingIndex = -1
				climbEndingIndex = -1

			}
		} else if i+1 == len(gpxItems)-1 && climbStarted { // End of the track reached
			climbEndingIndex = i + 1

			finishClimb(&currentClimb, &currentDescent, &climbStarted, &descentStarted, distance)
			validateClimb(currentClimb, &climbStartEndIndices, climbStartingIndex, climbEndingIndex, &climbs, gpxItems, distances, elevationGains)
		}
		distance += distanceBetweenPoints
	}

	for i := range climbs {
		climbs[i].ClimbSegments = getClimbSegments(distances, gpxItems, climbStartEndIndices[i][0], climbStartEndIndices[i][1])
	}
	return climbs
}

func validateClimb(currentClimb types.Climb, climbStartEndIndices *[][2]int, climbStartingIndex int, climbEndingIndex int, climbs *[]types.Climb, gpxItems []types.GpxItem, distances []float64, elevationGains []int) {
	if currentClimb.IsValidClimb() {
		*climbStartEndIndices = append(*climbStartEndIndices, [2]int{climbStartingIndex, climbEndingIndex})
		*climbs = append(*climbs, currentClimb)
	} else {
		start, end := findHiddenClimbs(climbs, &currentClimb, gpxItems, climbStartingIndex, climbEndingIndex, distances, elevationGains)
		if start != -1 && end != -1 {
			*climbStartEndIndices = append(*climbStartEndIndices, [2]int{start, end})
		}
	}
}

func finishClimb(currentClimb *types.Climb, currentDescent *types.Climb, climbStarted *bool, descentStarted *bool, distance float64) {
	// Finish the climb

	*climbStarted = false
	*descentStarted = false
	currentClimb.End = distance

	if currentClimb.Length > 0 {
		currentClimb.AverageGradient = calculateGradientPercent(float64(currentClimb.ElevationGain), currentClimb.Length)
		//currentClimb.Length += currentDescent.Length
	}
	*currentDescent = types.Climb{}
}

func gatherHiddenClimbs(climbPoints []types.GpxItem, distances []float64, elevationGains []int, climbStartingIndex int) (int, int) {
	bestLength := float64(-1)
	bestGradient := float64(-1)
	bestStartingIndex := -1
	bestEndingIndex := -1

	for i := 0; i < len(climbPoints)-1; i++ {
		for j := i + 1; j < len(climbPoints); j++ {
			length := distances[climbStartingIndex+j] - distances[climbStartingIndex+i]
			if length < 500 {
				continue
			}
			elevationGain := elevationGains[climbStartingIndex+j] - elevationGains[climbStartingIndex+i]
			gradient := calculateGradientPercent(float64(elevationGain), length)
			if gradient < 3.0 {
				continue
			}
			if length > bestLength {
				bestLength = length
				bestGradient = gradient
				bestStartingIndex = i
				bestEndingIndex = j
			} else if length == bestLength {
				if gradient > bestGradient {
					bestGradient = gradient
					bestStartingIndex = i
					bestEndingIndex = j
				}
			}
		}
	}
	return bestStartingIndex, bestEndingIndex
}

func getClimbSegments(distances []float64, gpxItems []types.GpxItem, climbStartingIndex int, climbEndingIndex int) []types.ClimbSegment {
	climbSegments := []types.ClimbSegment{}
	if climbEndingIndex < climbStartingIndex {
		return climbSegments
	}

	segmentStartIdx := climbStartingIndex
	segmentStartDist := distances[segmentStartIdx]
	segmentStartElev := gpxItems[segmentStartIdx].Elevation.Value()
	segmentProfile := []types.ElevationProfilePlotData{
		{Distance: 0, Elevation: int(segmentStartElev)},
	}
	segmentLength := 0.0
	segmentElevGain := 0.0
	segmentGradient := 0.0

	for i := climbStartingIndex + 1; i <= climbEndingIndex; i++ {
		curDist := distances[i] - segmentStartDist
		curElev := gpxItems[i].Elevation.Value()
		elevDiff := curElev - gpxItems[i-1].Elevation.Value()
		distDiff := distances[i] - distances[i-1]
		segmentLength += distDiff
		if elevDiff > 0 {
			segmentElevGain += elevDiff
		}
		segmentProfile = append(segmentProfile, types.ElevationProfilePlotData{
			Distance:  curDist,
			Elevation: int(curElev),
		})

		// Split segment if gradient changes significantly
		curGradient := calculateGradientPercent(elevDiff, distDiff)
		prevGradient := calculateGradientPercent(segmentElevGain, segmentLength)
		if i < climbEndingIndex && !checkIfClimbGradientsInRange(prevGradient, curGradient) && segmentLength >= 200 {
			segmentGradient = prevGradient
			climbSegments = append(climbSegments, types.ClimbSegment{
				ElevationProfile: segmentProfile,
				SegmentLength:    segmentLength,
				AverageGradient:  segmentGradient,
			})
			// Start new segment
			prevDistance := segmentProfile[len(segmentProfile)-1].Distance
			// segmentStartIdx = i
			// segmentStartDist = distances[i]
			segmentStartElev = curElev
			segmentProfile = []types.ElevationProfilePlotData{
				{Distance: prevDistance, Elevation: int(segmentStartElev)},
			}
			segmentLength = 0.0
			segmentElevGain = 0.0
		}
	}
	// Add last segment
	if segmentLength > 0 {
		segmentGradient = calculateGradientPercent(segmentElevGain, segmentLength)
		climbSegments = append(climbSegments, types.ClimbSegment{
			ElevationProfile: segmentProfile,
			SegmentLength:    segmentLength,
			AverageGradient:  segmentGradient,
		})
	}
	return climbSegments
}

// Find hidden climbs in the section that was not identified as a climb
// because there might be short steep sections within a longer flat section
func findHiddenClimbs(climbs *[]types.Climb, currentClimb *types.Climb, gpxItems []types.GpxItem, climbStartingIndex int, climbEndingIndex int, distances []float64, elevationGains []int) (int, int) {

	actualClimbStartingIndex, actualClimbEndingIndex := gatherHiddenClimbs(gpxItems[climbStartingIndex:climbEndingIndex+1], distances, elevationGains, climbStartingIndex)
	if actualClimbStartingIndex == -1 || actualClimbEndingIndex == -1 {
		return -1, -1
	}
	actualClimbLength := distances[climbStartingIndex+actualClimbEndingIndex] - distances[climbStartingIndex+actualClimbStartingIndex]
	actualClimbElevationGain := elevationGains[climbStartingIndex+actualClimbEndingIndex] - elevationGains[climbStartingIndex+actualClimbStartingIndex]

	currentClimb.AverageGradient = calculateGradientPercent(float64(actualClimbElevationGain), actualClimbLength)
	currentClimb.Length = actualClimbLength
	currentClimb.ElevationGain = actualClimbElevationGain
	currentClimb.Start = distances[climbStartingIndex+actualClimbStartingIndex]
	currentClimb.End = distances[climbStartingIndex+actualClimbEndingIndex]
	*climbs = append(*climbs, *currentClimb)

	return climbStartingIndex + actualClimbStartingIndex, climbStartingIndex + actualClimbEndingIndex
}
