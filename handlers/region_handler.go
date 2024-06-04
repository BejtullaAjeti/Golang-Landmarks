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

// GetRegions returns all regions
func GetRegions(c *gin.Context) {
	var regions []models.Region
	db.DB.Find(&regions)

	c.JSON(http.StatusOK, regions)
}

// GetRegionByID returns a region by ID
func GetRegionByID(c *gin.Context) {
	var region models.Region
	id := c.Param("id")
	db.DB.First(&region, id)

	if region.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
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
	minPopulation := c.Query("min_population")
	maxPopulation := c.Query("max_population")
	minLatitude := c.Query("min_latitude")
	maxLatitude := c.Query("max_latitude")
	minLongitude := c.Query("min_longitude")
	maxLongitude := c.Query("max_longitude")

	var regions []models.Region
	query := db.DB

	if minPopulation != "" {
		query = query.Where("population >= ?", minPopulation)
	}
	if maxPopulation != "" {
		query = query.Where("population <= ?", maxPopulation)
	}
	if minLatitude != "" {
		query = query.Where("CAST(latitude AS DECIMAL) >= ?", minLatitude)
	}
	if maxLatitude != "" {
		query = query.Where("CAST(latitude AS DECIMAL) <= ?", maxLatitude)
	}
	if minLongitude != "" {
		query = query.Where("CAST(longitude AS DECIMAL) >= ?", minLongitude)
	}
	if maxLongitude != "" {
		query = query.Where("CAST(longitude AS DECIMAL) <= ?", maxLongitude)
	}

	query.Find(&regions)
	c.JSON(http.StatusOK, regions)
}
