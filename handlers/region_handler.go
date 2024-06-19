package handlers

import (
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateRegion creates a new region
func CreateRegion(c *gin.Context) {
	var region models.Region
	if err := c.BindJSON(&region); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region data"})
		return
	}

	db.DB.Create(&region)

	c.JSON(http.StatusCreated, region)
}

// GetRegions returns all regions with associated GeoJSON if available
func GetRegions(c *gin.Context) {
	var regions []models.Region
	db.DB.Find(&regions)

	// Fetch GeoJSON for each region if available
	for i := range regions {
		geoJSON, err := db.GetGeoJSONByRegionID(regions[i].ID)
		if err == nil && geoJSON != nil {
			regions[i].GeoJSON = geoJSON.GeoJSONData
		}
	}

	c.JSON(http.StatusOK, regions)
}

// GetRegionByID returns a region by ID along with its GeoJSON data if available
func GetRegionByID(c *gin.Context) {
	var region models.Region
	id := c.Param("id")
	db.DB.First(&region, id)

	if region.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}

	// Fetch GeoJSON associated with the region
	geoJSON, err := db.GetGeoJSONByRegionID(region.ID)
	if err == nil && geoJSON != nil {
		region.GeoJSON = geoJSON.GeoJSONData
	}

	c.JSON(http.StatusOK, region)
}

// UpdateRegion updates a region by ID
func UpdateRegion(c *gin.Context) {
	var region models.Region
	id := c.Param("id")

	if err := db.DB.First(&region, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}

	if err := c.BindJSON(&region); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region data"})
		return
	}

	db.DB.Save(&region)

	c.JSON(http.StatusOK, region)
}

// DeleteRegion deletes a region by ID
func DeleteRegion(c *gin.Context) {
	var region models.Region
	id := c.Param("id")

	if err := db.DB.First(&region, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}

	db.DB.Delete(&region)

	c.Status(http.StatusNoContent)
}

func SearchRegions(c *gin.Context) {
	var regions []models.Region
	name := c.Query("name")
	query := db.DB.Where("name LIKE ?", "%"+name+"%")
	query.Find(&regions)
	c.JSON(http.StatusOK, regions)
}

func FilterRegions(c *gin.Context) {
	var regions []models.Region

	// Define the query parameters and their corresponding SQL clauses
	params := map[string]string{
		"min_population": "population >= ?",
		"max_population": "population <= ?",
	}

	query := db.DB
	for param, clause := range params {
		if value := c.Query(param); value != "" {
			query = query.Where(clause, value)
		}
	}

	if err := query.Find(&regions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve regions"})
		return
	}

	if len(regions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No region found"})
		return
	}

	c.JSON(http.StatusOK, regions)
}
