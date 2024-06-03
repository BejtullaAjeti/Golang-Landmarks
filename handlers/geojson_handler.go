package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"
	"os"
	"path/filepath"
)

func CreateGeoJSONFile(c *gin.Context) {
	var region models.Region
	id := c.Param("id")

	// Retrieve the region from the database
	if err := db.DB.First(&region, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}

	// Parse the coordinates from the JSON string
	var coordinates [][]float64
	if err := json.Unmarshal([]byte(region.Coordinates), &coordinates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse coordinates"})
		return
	}

	// Create the GeoJSON structure
	geoJSON := models.GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		Features: []models.GeoJSONFeature{
			{
				Type: "Feature",
				Geometry: models.GeoJSONGeometry{
					Type:        "Polygon",
					Coordinates: coordinates,
				},
			},
		},
	}

	// Marshal the GeoJSON structure to JSON
	geoJSONData, err := json.MarshalIndent(geoJSON, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create GeoJSON"})
		return
	}

	// Define the file path
	filePath := filepath.Join("geojson_files", "region_"+id+".geojson")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Write the GeoJSON data to a file
	if err := os.WriteFile(filePath, geoJSONData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write GeoJSON file"})
		return
	}

	// Respond with the file path
	c.JSON(http.StatusOK, gin.H{"file_path": filePath})
}
