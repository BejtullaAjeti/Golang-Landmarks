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
	if err := c.BindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review data"})
		return
	}

	db.DB.Create(&review)
	c.JSON(http.StatusCreated, review)
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
	uuid := c.Param("device_id")

	db.DB.Where("device_id = ?", uuid).Find(&reviews)
	c.JSON(http.StatusOK, reviews)
}
