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

func CreateGeoJSON(c *gin.Context) {
	var region models.Region
	id := c.Param("region_id")

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

	// Create a new GeoJSON record
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
}

func GetAllGeoJSON(c *gin.Context) {
	var geoJSONRecords []models.GeoJSON

	// Retrieve all GeoJSON records from the database
	if err := db.DB.Find(&geoJSONRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve GeoJSON records"})
		return
	}

	// Respond with the list of GeoJSON records
	c.JSON(http.StatusOK, geoJSONRecords)
}

func UpdateGeoJSON(c *gin.Context) {
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
		c.JSON(http.StatusNotFound, gin.H{"error": "GeoJSON not found"})
		return
	}

	// Update the existing record
	existingGeoJSON.GeoJSONData = string(geoJSONBytes)
	existingGeoJSON.MiddlePoint = middlePointString
	existingGeoJSON.Zoom = zoom
	existingGeoJSON.UpdatedAt = time.Now() // Update updatedAt timestamp

	if err := db.DB.Save(&existingGeoJSON).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update GeoJSON in database"})
		return
	}

	// Update the region's updatedAt timestamp
	region.UpdatedAt = time.Now()
	if err := db.DB.Save(&region).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update region's updatedAt in database"})
		return
	}

	// Respond with the GeoJSON record ID
	c.JSON(http.StatusOK, gin.H{"geojson_id": existingGeoJSON.ID})
}

const TimeFormat = "2006-01-02T15:04:05Z07:00"

func GetGeoJSONFromDB(c *gin.Context) {
	regionID := c.Param("id")

	var geoJSONRecord models.GeoJSON

	// Retrieve the GeoJSON record from the database
	if err := db.DB.Where("region_id = ?", regionID).First(&geoJSONRecord).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "GeoJSON not found"})
		return
	}

	// Parse the If-Modified-Since header from the request
	ifModifiedSince := c.GetHeader("If-Modified-Since")
	if ifModifiedSince != "" {
		parsedIfModifiedSince, err := time.Parse(TimeFormat, ifModifiedSince)
		if err == nil && !geoJSONRecord.UpdatedAt.After(parsedIfModifiedSince) {
			// Resource has not been modified since the date in If-Modified-Since header
			c.Status(http.StatusNotModified)
			return
		}
	}

	var geoJSONData map[string]interface{}
	if err := json.Unmarshal([]byte(geoJSONRecord.GeoJSONData), &geoJSONData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse GeoJSON"})
		return
	}

	response := gin.H{
		"geojson_data": geoJSONData,
		"created_at":   geoJSONRecord.CreatedAt,
	}

	// Set the Last-Modified header in the response
	c.Header("Last-Modified", geoJSONRecord.UpdatedAt.UTC().Format(TimeFormat))

	// Respond with the GeoJSON data and created_at
	c.JSON(http.StatusOK, response)
}
