package handlers

import (
	"github.com/gin-gonic/gin"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"log"
	"net/http"
)

// CreateLandmark handles creating a new landmark without photos
func CreateLandmark(c *gin.Context) {
	var input models.Landmark

	// Check if content type is JSON
	if c.Request.Header.Get("Content-Type") == "application/json" {
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid landmark data"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content-Type must be application/json"})
		return
	}

	// Validate city existence
	var city models.City
	if err := db.DB.First(&city, input.CityID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City with the specified city_id does not exist"})
		return
	}

	// Create the landmark
	if err := db.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create landmark"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func GetLandmarks(c *gin.Context) {
	var landmarks []models.Landmark

	// Retrieve landmarks from the database
	if err := db.DB.Find(&landmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve landmarks"})
		return
	}

	// Populate PhotoLinks and omit Photos
	for i := range landmarks {
		// Retrieve photos
		var photos []models.LandmarkPhoto
		if err := db.DB.Where("landmark_id = ?", landmarks[i].ID).Find(&photos).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve photos for landmark"})
			return
		}

		var photoLinks []string
		for _, photo := range photos {
			photoLinks = append(photoLinks, photo.Path)
		}
		landmarks[i].Photos = nil // Clear Photos field
		landmarks[i].PhotoLinks = photoLinks

		// Retrieve top 10 reviews
		var reviews []models.Review
		if err := db.DB.Where("landmark_id = ?", landmarks[i].ID).Order("created_at desc").Limit(10).Find(&reviews).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reviews for landmark"})
			return
		}

		// For each review, populate PhotoLinks and omit Photos
		for j := range reviews {
			var reviewPhotos []models.ReviewPhoto
			if err := db.DB.Where("review_id = ?", reviews[j].ID).Find(&reviewPhotos).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve photos for review"})
				return
			}

			var reviewPhotoLinks []string
			for _, photo := range reviewPhotos {
				reviewPhotoLinks = append(reviewPhotoLinks, photo.Path)
			}
			reviews[j].Photos = nil // Clear Photos field
			reviews[j].PhotoLinks = reviewPhotoLinks
		}

		// Add reviews to the landmark
		landmarks[i].Reviews = reviews
	}

	c.JSON(http.StatusOK, landmarks)
}

func GetLandmarkByID(c *gin.Context) {
	var landmark models.Landmark
	id := c.Param("id")
	log.Printf("Fetching landmark with ID: %s", id)
	if err := db.DB.First(&landmark, id).Error; err != nil {
		log.Println("Landmark not found or an error occurred:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Landmark not found"})
		return
	}

	var photos []models.LandmarkPhoto
	if err := db.DB.Where("landmark_id = ?", landmark.ID).Find(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve photos for landmark"})
		return
	}

	var photoLinks []string
	for _, photo := range photos {
		photoLinks = append(photoLinks, photo.Path)
	}

	landmark.Photos = nil // Clear Photos field
	landmark.PhotoLinks = photoLinks

	c.JSON(http.StatusOK, landmark)
}

func GetLandmarkDetails(c *gin.Context) {
	var (
		landmark      models.Landmark
		reviews       []models.Review
		reviewCount   int64
		averageRating float64
		userReview    models.Review // User review for the landmark
	)

	landmarkID := c.Param("id")
	deviceID := c.Query("device_id") // Get device ID from query parameter

	log.Printf("Fetching details for landmark with ID: %s", landmarkID)

	// Fetch Landmark with Photos
	if err := db.DB.Preload("Photos").First(&landmark, landmarkID).Error; err != nil {
		log.Println("Landmark not found or an error occurred:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Landmark not found"})
		return
	}

	// Fetch Reviews
	db.DB.Where("landmark_id = ?", landmarkID).Find(&reviews)

	// Fetch Review Count
	db.DB.Model(&models.Review{}).Where("landmark_id = ?", landmarkID).Count(&reviewCount)

	// Fetch Average Rating
	var result struct {
		AverageRating float64
	}
	db.DB.Model(&models.Review{}).
		Select("AVG(rating) as average_rating").
		Where("landmark_id = ?", landmarkID).
		Scan(&result)
	averageRating = result.AverageRating

	// If device ID is provided, fetch reviews by device ID for the landmark
	if deviceID != "" {
		var userReviews []models.Review
		db.DB.Where("landmark_id = ? AND device_id = ?", landmarkID, deviceID).Find(&userReviews)
		if len(userReviews) > 0 {
			userReview = userReviews[0] // Assuming there's only one review per user for a landmark
		}
	}

	// Construct JSON response
	response := gin.H{
		"landmark":       landmark,
		"reviews":        reviews,
		"review_count":   reviewCount,
		"average_rating": averageRating,
	}

	// Include the user's review only if device ID is provided
	if deviceID != "" {
		response["user_review"] = userReview
	}

	// Send JSON response
	c.JSON(http.StatusOK, response)
}

// UpdateLandmark updates a landmark by ID
func UpdateLandmark(c *gin.Context) {
	id := c.Param("id")
	var landmark models.Landmark

	// Check if the landmark exists
	if err := db.DB.First(&landmark, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Landmark not found"})
		return
	}

	// Bind updated landmark data from JSON request body
	if err := c.ShouldBindJSON(&landmark); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}

	// Validate city existence
	var city models.City
	if err := db.DB.First(&city, landmark.CityID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City with the specified city_id does not exist"})
		return
	}

	// Save updated landmark data
	if err := db.DB.Save(&landmark).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update landmark"})
		return
	}

	c.JSON(http.StatusOK, landmark)
}

func DeleteLandmark(c *gin.Context) {
	var landmark models.Landmark
	id := c.Param("id")

	if err := db.DB.First(&landmark, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "landmark not found"})
		return
	}

	db.DB.Delete(&landmark)

	c.Status(http.StatusNoContent)
}

func SearchLandmarks(c *gin.Context) {
	var landmarks []models.Landmark
	keyword := c.Query("keyword")

	db.DB.Where("name LIKE ?", "%"+keyword+"%").
		Or("description LIKE ?", "%"+keyword+"%").
		Find(&landmarks)

	c.JSON(http.StatusOK, landmarks)
}

func FilterLandmarks(c *gin.Context) {
	var landmarks []models.Landmark

	// Define the query parameters and their corresponding SQL clauses
	params := map[string]string{
		"city_id":       "city_id = ?",
		"type":          "type = ?",
		"min_latitude":  "CAST(latitude AS DECIMAL) >= ?",
		"max_latitude":  "CAST(latitude AS DECIMAL) <= ?",
		"min_longitude": "CAST(longitude AS DECIMAL) >= ?",
		"max_longitude": "CAST(longitude AS DECIMAL) <= ?",
	}

	query := db.DB
	for param, clause := range params {
		if value := c.Query(param); value != "" {
			query = query.Where(clause, value)
		}
	}

	if err := query.Find(&landmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve landmarks"})
		return
	}

	if len(landmarks) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No landmarks found"})
		return
	}

	c.JSON(http.StatusOK, landmarks)
}

func GetAllLandmarksOfCity(c *gin.Context) {
	var landmarks []models.Landmark
	cityID := c.Param("city_id")

	var city models.City
	if err := db.DB.First(&city, cityID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City with the specified city_id does not exist"})
		return
	}

	db.DB.Where("city_id = ?", cityID).Find(&landmarks)

	c.JSON(http.StatusOK, landmarks)
}

func GetAllLandmarksOfRegion(c *gin.Context) {
	var landmarks []models.Landmark
	regionID := c.Param("region_id")

	var region models.Region
	if err := db.DB.First(&region, regionID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Region with the specified region_id does not exist"})
		return
	}

	db.DB.Joins("JOIN cities ON landmarks.city_id = cities.id").
		Where("cities.region_id = ?", regionID).
		Find(&landmarks)

	c.JSON(http.StatusOK, landmarks)
}
