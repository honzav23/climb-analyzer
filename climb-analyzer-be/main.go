package main

import (
	"math"
	"net/http"
	"os"

	"climb-analyzer-be/types"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tkrajina/gpxgo/gpx"
)

func parseGPX() (*gpx.GPX, error) {
	fileData, err := os.ReadFile("analyze.gpx")
	var gpxData *gpx.GPX = nil
	if err != nil {
		return gpxData, err
	}
	gpxData, err = gpx.ParseBytes(fileData)
	if err != nil {
		return gpxData, err
	}
	return gpxData, nil
}

// Extracts all points and their elevations from the GPX data
// to have a more convenient structure to work with
func extractGpxItems(gpxData *gpx.GPX) []types.GpxItem {
	gpxItems := []types.GpxItem{}
	for _, track := range gpxData.Tracks {
		for _, segment := range track.Segments {
			elevations := segment.Elevations()
			points := segment.Points
			for i := 0; i < len(points); i++ {
				gpxItems = append(gpxItems, types.GpxItem{Point: points[i], Elevation: elevations[i]})
			}
		}
	}
	return gpxItems
}

func calculateElevationGain(gpxItems []types.GpxItem) int {
	totalGain := 0

	for i := 1; i < len(gpxItems); i++ {
		elevationDiff := gpxItems[i].Elevation.Value() - gpxItems[i-1].Elevation.Value()
		if elevationDiff > 0 {
			totalGain += int(elevationDiff)
		}
	}
	return totalGain
}

func generateElevationProfile(gpxItems []types.GpxItem) []types.ElevationProfilePlotData {
	profile := []types.ElevationProfilePlotData{}
	profile = append(profile, types.ElevationProfilePlotData{Distance: 0, Elevation: int(gpxItems[0].Elevation.Value())})
	distance := float64(0)
	for i := 0; i < len(gpxItems)-1; i++ {
		distance += gpxItems[i].Point.Distance3D(&gpxItems[i+1].Point)
		profile = append(profile, types.ElevationProfilePlotData{Distance: distance, Elevation: int(gpxItems[i].Elevation.Value())})
	}
	return profile
}

func getTripRouteCoordinates(gpxItems []types.GpxItem) []types.PointCoordinates {
	coordinates := []types.PointCoordinates{}
	for _, gpxItem := range gpxItems {
		coordinates = append(coordinates, types.PointCoordinates{Latitude: gpxItem.Point.Latitude, Longitude: gpxItem.Point.Longitude})
	}

	return coordinates
}

func getSummaryInfo(response *types.AnalysisResponse, gpxData *gpx.GPX, gpxItems []types.GpxItem) {
	response.TripSummary = types.TripSummary{
		LengthKilometers: float32(math.Round(gpxData.Length3D()/100) / 10),
		ElevationGain:    calculateElevationGain(gpxItems),
		ElevationProfile: generateElevationProfile(gpxItems),
		TripCoordinates:  getTripRouteCoordinates(gpxItems),
	}
}

func analyzeClimbs(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "File not provided"})
		return
	}
	c.SaveUploadedFile(file, "./analyze.gpx")
	gpxData, err := parseGPX()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse GPX file"})
		return
	}
	gpxItems := extractGpxItems(gpxData)

	response := types.AnalysisResponse{}
	getSummaryInfo(&response, gpxData, gpxItems)
	response.FoundClimbs = IdentifyClimbs(gpxItems)
	c.IndentedJSON(http.StatusOK, response)
}

func recoveryMiddleware(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error"})
		}
	}()
	c.Next()
}

func main() {
	isProduction := false
	backendUrl := "localhost:8080"

	_, defined := os.LookupEnv("PRODUCTION")
	if defined {
		isProduction = true
	}

	if isProduction {
		gin.SetMode(gin.ReleaseMode)
		backendUrl = "0.0.0.0:8080"
	}

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:4200", "http://localhost"}
	config.AllowMethods = []string{"POST", "OPTIONS"}
	router.Use(cors.New(config))

	router.Use(recoveryMiddleware)
	router.POST("/api/analyze", analyzeClimbs)
	router.Run(backendUrl)
}
