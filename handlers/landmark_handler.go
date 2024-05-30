package handlers

import (
	"landmarksmodule/db"
	"landmarksmodule/models"
	"log"

	"github.com/gin-gonic/gin"
)

func CreateLandmark(c *gin.Context) {
	var landmark models.Landmark
	if err := c.BindJSON(&landmark); err != nil {
		c.JSON(400, gin.H{"error": "Invalid landmark data"})
		return
	}
	var city models.City
	if err := db.DB.First(&city, landmark.CityID).Error; err != nil {
		c.JSON(400, gin.H{"error": "City with the specified city_id does not exist"})
		return
	}

	db.DB.Create(&landmark)

	c.JSON(201, landmark)
}

func GetLandmarks(c *gin.Context) {
	var landmarks []models.Landmark
	db.DB.Find(&landmarks)

	c.JSON(200, landmarks)
}

func GetLandmarkByID(c *gin.Context) {
	var landmark models.Landmark
	id := c.Param("id")
	log.Printf("Fetching landmark with ID: %s", id) // Add logging
	db.DB.First(&landmark, id)

	if landmark.ID == 0 {
		log.Println("Landmark not found") // Add logging
		c.JSON(404, gin.H{"error": "Landmark not found"})
		return
	}
	c.JSON(200, landmark)
}

func UpdateLandmark(c *gin.Context) {
	var landmark models.Landmark
	id := c.Param("id")

	if err := db.DB.First(&landmark, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Landmark not found"})
		return
	}

	if err := c.BindJSON(&landmark); err != nil {
		c.JSON(400, gin.H{"error": "Invalid landmark data"})
		return
	}
	var city models.City
	if err := db.DB.First(&city, landmark.CityID).Error; err != nil {
		c.JSON(400, gin.H{"error": "City with the specified city_id does not exist"})
		return
	}

	db.DB.Save(&landmark)

	c.JSON(200, landmark)
}

func DeleteLandmark(c *gin.Context) {
	var landmark models.Landmark
	id := c.Param("id")

	if err := db.DB.First(&landmark, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "landmark not found"})
		return
	}

	db.DB.Delete(&landmark)

	c.Status(204)
}

// SearchLandmarks searches for landmarks based on a keyword
func SearchLandmarks(c *gin.Context) {
	var landmarks []models.Landmark
	keyword := c.Query("keyword")

	db.DB.Where("name LIKE ?", "%"+keyword+"%").
		Or("description LIKE ?", "%"+keyword+"%").
		Find(&landmarks)

	c.JSON(200, landmarks)
}

// FilterLandmarks handles filtering landmarks based on city ID or type
func FilterLandmarks(c *gin.Context) {
	var landmarks []models.Landmark
	var err error

	cityID := c.Query("city_id")
	landmarkType := c.Query("type")
	minLatitude := c.Query("min_latitude")
	maxLatitude := c.Query("max_latitude")
	minLongitude := c.Query("min_longitude")
	maxLongitude := c.Query("max_longitude")

	query := db.DB

	if cityID != "" {
		query = query.Where("city_id = ?", cityID)
	}
	if landmarkType != "" {
		query = query.Where("type = ?", landmarkType)
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

	err = query.Find(&landmarks).Error

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve landmarks"})
		return
	}

	if len(landmarks) == 0 {
		c.JSON(404, gin.H{"message": "No landmarks found"})
		return
	}

	c.JSON(200, landmarks)
}

// GetAllLandmarksOfCity retrieves all landmarks belonging to a specific city
func GetAllLandmarksOfCity(c *gin.Context) {
	var landmarks []models.Landmark
	cityID := c.Param("city_id")

	// Check if the city with the specified city_id exists
	var city models.City
	if err := db.DB.First(&city, cityID).Error; err != nil {
		c.JSON(400, gin.H{"error": "City with the specified city_id does not exist"})
		return
	}

	// City with the specified city_id exists, proceed to fetch all landmarks of the city
	db.DB.Where("city_id = ?", cityID).Find(&landmarks)

	c.JSON(200, landmarks)
}

// GetAllLandmarksOfRegion retrieves all landmarks belonging to a specific region
func GetAllLandmarksOfRegion(c *gin.Context) {
	var landmarks []models.Landmark
	regionID := c.Param("region_id")

	// Check if the region with the specified region_id exists
	var region models.Region
	if err := db.DB.First(&region, regionID).Error; err != nil {
		c.JSON(400, gin.H{"error": "Region with the specified region_id does not exist"})
		return
	}

	// Region with the specified region_id exists, proceed to fetch all landmarks of the region
	db.DB.Joins("JOIN cities ON landmarks.city_id = cities.id").
		Where("cities.region_id = ?", regionID).
		Find(&landmarks)

	c.JSON(200, landmarks)
}
