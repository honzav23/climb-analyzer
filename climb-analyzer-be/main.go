package main

import (
	"math"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tkrajina/gpxgo/gpx"
)

type ElevationProfilePlotData struct {
	Distance  float32 `json:"distance"`
	Elevation int     `json:"elevation"`
}

type TripSummary struct {
	LengthKilometers float32                    `json:"lengthKilometers"`
	ElevationGain    int                        `json:"elevationGain"`
	ElevationProfile []ElevationProfilePlotData `json:"elevationProfile"`
}

type AnalysisResponse struct {
	TripSummary TripSummary `json:"tripSummary"`
}

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

func calculateElevationGain(gpxData *gpx.GPX) int {
	totalGain := 0
	for _, track := range gpxData.Tracks {
		for _, segment := range track.Segments {
			elevations := segment.Elevations()
			for i := 1; i < len(elevations); i++ {
				elevationDiff := elevations[i].Value() - elevations[i-1].Value()
				if elevationDiff > 0 {
					totalGain += int(elevationDiff)
				}
			}
		}
	}
	return totalGain
}

func generateElevationProfile(gpxData *gpx.GPX) []ElevationProfilePlotData {
	profile := []ElevationProfilePlotData{}
	profile = append(profile, ElevationProfilePlotData{Distance: 0, Elevation: int(gpxData.Tracks[0].Segments[0].Points[0].Elevation.Value())})
	distance := float32(0)
	for _, track := range gpxData.Tracks {
		for _, segment := range track.Segments {
			elevations := segment.Elevations()
			points := segment.Points
			for i := 1; i < len(points); i++ {
				distance += float32(points[i].Distance3D(&points[i-1]))
				profile = append(profile, ElevationProfilePlotData{Distance: distance, Elevation: int(elevations[i].Value())})
			}
		}
	}
	return profile
}

func getSummaryInfo(response *AnalysisResponse, gpxData *gpx.GPX) {
	response.TripSummary = TripSummary{
		LengthKilometers: float32(math.Round(gpxData.Length3D()/100) / 10),
		ElevationGain:    calculateElevationGain(gpxData),
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
	response := AnalysisResponse{}
	getSummaryInfo(&response, gpxData)
	response.TripSummary.ElevationProfile = generateElevationProfile(gpxData)
	c.IndentedJSON(http.StatusOK, response)
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	router.POST("/analyze", analyzeClimbs)
	router.Run("localhost:8080")
}
