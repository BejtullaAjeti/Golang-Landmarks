package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"
	"time"
)

func CreateGeoJSONInDB(c *gin.Context) {
	var region models.Region
	id := c.Param("id")

	// Retrieve the region from the database
	if err := db.DB.First(&region, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}

	// Parse the JSON request body
	var geoJSONData map[string]interface{}
	if err := c.BindJSON(&geoJSONData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid GeoJSON data"})
		return
	}

	// Marshal the GeoJSON data
	geoJSONBytes, err := json.Marshal(geoJSONData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal GeoJSON"})
		return
	}

	// Extract middle point coordinates from request body
	var middlePoint [2]float64
	if middlePointJSON, ok := geoJSONData["middle_point"].([]interface{}); ok && len(middlePointJSON) == 2 {
		if x, ok := middlePointJSON[0].(float64); ok {
			middlePoint[0] = x
		}
		if y, ok := middlePointJSON[1].(float64); ok {
			middlePoint[1] = y
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid middle_point format"})
		return
	}

	// Extract zoom level from request body
	var zoom float64
	if zoomJSON, ok := geoJSONData["zoom"].(float64); ok {
		zoom = zoomJSON
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid zoom format"})
		return
	}

	// Convert middle point to string format
	middlePointString := fmt.Sprintf("[%f, %f]", middlePoint[0], middlePoint[1])

	// Check if a GeoJSON record already exists for the region
	var existingGeoJSON models.GeoJSON
	if err := db.DB.Where("region_id = ?", region.ID).First(&existingGeoJSON).Error; err != nil {
		// If not found, create a new record
		geoJSONRecord := models.GeoJSON{
			RegionID:    region.ID,
			GeoJSONData: string(geoJSONBytes),
			MiddlePoint: middlePointString,
			Zoom:        zoom,
			UpdatedAt:   time.Now(), // Set updated timestamp
		}

		if err := db.DB.Create(&geoJSONRecord).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save GeoJSON to database"})
			return
		}

		// Respond with the GeoJSON record ID
		c.JSON(http.StatusOK, gin.H{"geojson_id": geoJSONRecord.ID})
		return
	}

	// If found, update the existing record
	existingGeoJSON.GeoJSONData = string(geoJSONBytes)
	existingGeoJSON.MiddlePoint = middlePointString
	existingGeoJSON.Zoom = zoom
	existingGeoJSON.UpdatedAt = time.Now() // Update updatedAt timestamp

	if err := db.DB.Save(&existingGeoJSON).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update GeoJSON in database"})
		return
	}

	// Respond with the GeoJSON record ID
	c.JSON(http.StatusOK, gin.H{"geojson_id": existingGeoJSON.ID})
}

func GetGeoJSONFromDB(c *gin.Context) {
	regionID := c.Param("id")

	var geoJSONRecord models.GeoJSON
	if err := db.DB.Where("id = ?", regionID).First(&geoJSONRecord).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "GeoJSON not found"})
		return
	}

	var geoJSONData map[string]interface{}
	if err := json.Unmarshal([]byte(geoJSONRecord.GeoJSONData), &geoJSONData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse GeoJSON"})
		return
	}

	// Respond with the GeoJSON data
	c.JSON(http.StatusOK, geoJSONData)
}
