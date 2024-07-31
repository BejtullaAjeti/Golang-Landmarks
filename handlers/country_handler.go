package handlers

import (
	"github.com/gin-gonic/gin"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"
	"strconv"
)

// GetCountries retrieves all countries and their regions
func GetCountries(c *gin.Context) {
	var countries []models.Country

	// Retrieve all countries
	if err := db.DB.Find(&countries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve countries", "details": err.Error()})
		return
	}

	// For each country, retrieve the associated regions and their GeoJSON data
	for i := range countries {
		var regions []models.Region
		if err := db.DB.Where("country_id = ?", countries[i].ID).Find(&regions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve regions for country", "details": err.Error()})
			return
		}

		// For each region, retrieve the GeoJSON data
		for j := range regions {
			geoJSON, err := db.GetGeoJSONByRegionID(regions[j].ID)
			if err == nil && geoJSON != nil {
				regions[j].GeoJSON = geoJSON.GeoJSONData
			}
		}

		// Assign the retrieved regions to the country
		countries[i].Regions = regions
	}

	c.JSON(http.StatusOK, countries)
}

// CreateCountry creates a new country
func CreateCountry(c *gin.Context) {
	var country models.Country
	if err := c.BindJSON(&country); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country data"})
		return
	}
	db.DB.Create(&country)

	c.JSON(http.StatusCreated, country)
}

// GetCountryByID retrieves a country by its ID
func GetCountryByID(c *gin.Context) {
	var country models.Country
	id := c.Param("id")
	db.DB.First(&country, id)

	var regions []models.Region
	if err := db.DB.Where("country_id = ?", country.ID).Find(&regions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve regions for country"})
		return
	}
	for i := range regions {
		geoJSON, err := db.GetGeoJSONByRegionID(regions[i].ID)
		if err == nil && geoJSON != nil {
			regions[i].GeoJSON = geoJSON.GeoJSONData
		}
	}
	country.Regions = regions
	if country.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Country not found"})
		return
	}

	c.JSON(http.StatusOK, country)
}

func GetCountryByLatLong(c *gin.Context) {
	latitude, err1 := strconv.ParseFloat(c.Query("latitude"), 64)
	longitude, err2 := strconv.ParseFloat(c.Query("longitude"), 64)

	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude or longitude"})
		return
	}

	var country models.Country
	query := `
		SELECT *,
		       (6371 * acos(cos(radians(?)) *
		                    cos(radians(latitude)) *
		                    cos(radians(longitude) - radians(?)) +
		                    sin(radians(?)) *
		                    sin(radians(latitude)))) AS distance
		FROM countries
		ORDER BY distance
		LIMIT 1`

	if err := db.DB.Raw(query, latitude, longitude, latitude).Scan(&country).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Country not found", "details": err.Error()})
		return
	}

	var regions []models.Region
	if err := db.DB.Where("country_id = ?", country.ID).Find(&regions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve regions for country", "details": err.Error()})
		return
	}

	for i := range regions {
		geoJSON, err := db.GetGeoJSONByRegionID(regions[i].ID)
		if err == nil && geoJSON != nil {
			regions[i].GeoJSON = geoJSON.GeoJSONData
		}
	}
	country.Regions = regions

	c.JSON(http.StatusOK, country)
}

// UpdateCountry updates a country's information
func UpdateCountry(c *gin.Context) {
	var country models.Country
	id := c.Param("id")

	if err := db.DB.First(&country, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Country not found", "details": err.Error()})
		return
	}

	if err := c.BindJSON(&country); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country data", "details": err.Error()})
		return
	}

	db.DB.Save(&country)

	c.JSON(http.StatusOK, country)
}

// DeleteCountry deletes a country by its ID
func DeleteCountry(c *gin.Context) {
	var country models.Country
	id := c.Param("id")

	if err := db.DB.First(&country, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Country not found", "details": err.Error()})
		return
	}

	db.DB.Delete(&country)

	c.JSON(http.StatusOK, gin.H{"message": "Country deleted successfully"})
}
