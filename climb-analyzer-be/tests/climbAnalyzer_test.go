package summary_test

import (
	"climb-analyzer-be/climbAnalyzer"
	"climb-analyzer-be/types"
	"os"
	"testing"

	"github.com/tkrajina/gpxgo/gpx"
)

func setupTest() (*gpx.GPX, []types.GpxItem, error) {
	fileData, err := os.ReadFile("export.gpx")
	var gpxData *gpx.GPX = nil
	gpxItems := []types.GpxItem{}
	if err != nil {
		return gpxData, gpxItems, err
	}
	gpxData, err = gpx.ParseBytes(fileData)
	if err != nil {
		return gpxData, gpxItems, err
	}
	gpxItems = climbAnalyzer.ExtractGpxItems(gpxData)
	return gpxData, gpxItems, nil
}

func TestFileParsing(t *testing.T) {

	// Test if the .gpx file can be read and parsed without errors
	fileData, err := os.ReadFile("export.gpx")
	if err != nil {
		t.Error("Unable to read .gpx file")
	}
	_, err = gpx.ParseBytes(fileData)
	if err != nil {
		t.Error("Unable to parse .gpx file")
	}

	// Test if invalid file returns an error
	_, err = gpx.ParseBytes([]byte("invalid gpx data"))
	if err == nil {
		t.Error("Expected error when parsing invalid gpx data, but got none")
	}
}

func TestTripSummary(t *testing.T) {
	gpxData, gpxItems, err := setupTest()
	if err != nil {
		t.Fatal("Unable to parse .gpx file")
	}

	response := types.AnalysisResponse{}
	climbAnalyzer.GetSummaryInfo(&response, gpxData, gpxItems)

	// Validate the trip summary
	if response.TripSummary.LengthKilometers != 73.4 {
		t.Errorf("Expected trip length 73.4 km, got %v km", response.TripSummary.LengthKilometers)
	}
	if response.TripSummary.ElevationGain != 1107 {
		t.Errorf("Expected elevation gain 1107 m, got %v m", response.TripSummary.ElevationGain)
	}
}
