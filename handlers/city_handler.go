package handlers

import (
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateCity(c *gin.Context) {
	var city models.City
	if err := c.BindJSON(&city); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid city data"})
		return
	}
	var region models.Region

	// Retrieve the region associated with the city's region_id
	if err := db.DB.First(&region, city.RegionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}
	db.DB.Create(&city)

	c.JSON(http.StatusCreated, city)
}

func GetCities(c *gin.Context) {
	var cities []models.City
	db.DB.Find(&cities)

	c.JSON(http.StatusOK, cities)
}

func GetCityByID(c *gin.Context) {
	var city models.City
	id := c.Param("id")
	db.DB.First(&city, id)

	if city.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "City not found"})
		return
	}

	c.JSON(http.StatusOK, city)
}

func UpdateCity(c *gin.Context) {
	var city models.City
	id := c.Param("id")

	if err := db.DB.First(&city, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "City not found"})
		return
	}

	if err := c.BindJSON(&city); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid city data"})
		return
	}

	var region models.Region

	// Retrieve the region associated with the city's region_id
	if err := db.DB.First(&region, city.RegionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}

	db.DB.Save(&city)

	c.JSON(http.StatusOK, city)
}

func DeleteCity(c *gin.Context) {
	var city models.City
	id := c.Param("id")

	if err := db.DB.First(&city, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "City not found"})
		return
	}

	db.DB.Delete(&city)

	c.Status(http.StatusNoContent)
}

func SearchCities(c *gin.Context) {
	var cities []models.City
	query := c.Query("name")

	if query != "" {
		db.DB.Where("name LIKE ?", "%"+query+"%").Find(&cities)
	} else {
		// If no query parameter is provided, return all cities
		db.DB.Find(&cities)
	}

	c.JSON(http.StatusOK, cities)
}

func FilterCities(c *gin.Context) {
	var cities []models.City

	// Define the query parameters and their corresponding SQL clauses
	params := map[string]string{
		"min_population": "population >= ?",
		"max_population": "population <= ?",
		"min_area":       "area >= ?",
		"max_area":       "area <= ?",
		"min_latitude":   "CAST(latitude AS DECIMAL) >= ?",
		"max_latitude":   "CAST(latitude AS DECIMAL) <= ?",
		"min_longitude":  "CAST(longitude AS DECIMAL) >= ?",
		"max_longitude":  "CAST(longitude AS DECIMAL) <= ?",
	}

	query := db.DB
	for param, clause := range params {
		if value := c.Query(param); value != "" {
			query = query.Where(clause, value)
		}
	}

	if err := query.Find(&cities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cities)
}

// GetRegionOfCity retrieves the region associated with a city by its region_id
func GetRegionOfCity(c *gin.Context) {
	var city models.City
	id := c.Param("id")

	// Retrieve the city from the database
	if err := db.DB.First(&city, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "City not found"})
		return
	}

	var region models.Region

	// Retrieve the region associated with the city's region_id
	if err := db.DB.First(&region, city.RegionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}

	c.JSON(http.StatusOK, region)
}
