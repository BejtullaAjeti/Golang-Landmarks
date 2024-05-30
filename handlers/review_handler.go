package handlers

import (
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateReview handles creating a new review
func CreateReview(c *gin.Context) {
	var review models.Review

	// Bind the JSON request body to the review struct
	if err := c.BindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review data"})
		return
	}

	// Check if rating is between 1 and 5
	if review.Rating < 1 || review.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 1 and 5"})
		return
	}

	// Check if device ID is not empty
	if review.DeviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID cannot be empty"})
		return
	}

	// Check if landmark ID exists
	var landmark models.Landmark
	if err := db.DB.First(&landmark, review.LandmarkID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark ID does not exist"})
		return
	}

	// Create the review in the database
	if err := db.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	// Return the created review
	c.JSON(http.StatusCreated, review)

	db.DB.Create(&review)
}

// GetReviews returns all reviews
func GetReviews(c *gin.Context) {
	var reviews []models.Review
	db.DB.Find(&reviews)
	c.JSON(http.StatusOK, reviews)
}

// GetReviewByID returns a review by ID
func GetReviewByID(c *gin.Context) {
	var review models.Review
	id := c.Param("id")
	db.DB.First(&review, id)

	if review.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// UpdateReview updates a review by ID
func UpdateReview(c *gin.Context) {
	var review models.Review
	id := c.Param("id")

	if err := db.DB.First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	if err := c.BindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review data"})
		return
	}
	if review.Rating < 1 || review.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 1 and 5"})
		return
	}
	var landmark models.Landmark
	if err := db.DB.First(&landmark, review.LandmarkID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark ID does not exist"})
		return
	}

	db.DB.Save(&review)
	c.JSON(http.StatusOK, review)
}

// DeleteReview deletes a review by ID
func DeleteReview(c *gin.Context) {
	var review models.Review
	id := c.Param("id")

	if err := db.DB.First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	db.DB.Delete(&review)
	c.Status(http.StatusNoContent)
}

// GetAverageRatingByLandmarkID calculates the average rating for a specific landmark
func GetAverageRatingByLandmarkID(c *gin.Context) {
	var result struct {
		AverageRating float64 `json:"average_rating"`
	}
	landmarkID := c.Param("id")

	// Check if the landmark ID exists
	var landmark models.Landmark
	if err := db.DB.First(&landmark, landmarkID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark ID does not exist"})
		return
	}

	err := db.DB.Model(&models.Review{}).
		Select("AVG(rating) as average_rating").
		Where("landmark_id = ?", landmarkID).
		Scan(&result).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate average rating"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetReviewsByDeviceID retrieves all reviews by a user based on their UUID
func GetReviewsByDeviceID(c *gin.Context) {
	var reviews []models.Review

	// Get the device ID from the request parameter
	uuid := c.Param("device_id")

	// Check if the device ID exists
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	// Query the database for reviews with the specified device ID
	db.DB.Where("uuid = ?", uuid).Find(&reviews)

	// Return the reviews in the response
	c.JSON(http.StatusOK, reviews)
}
