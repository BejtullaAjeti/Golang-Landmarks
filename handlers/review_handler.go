package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateReview handles creating a new review with attached images
func CreateReview(c *gin.Context) {
	var input struct {
		models.Review
		Photos []string `json:"photos"` // base64 encoded images
	}

	// Bind the JSON request body to the input struct
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review data"})
		return
	}

	// Validate review data
	if err := validateReview(&input.Review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if landmark exists
	if err := checkLandmarkExists(input.LandmarkID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the review
	if err := db.DB.Create(&input.Review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	// Create review photos if provided
	if len(input.Photos) > 0 {
		for _, base64image := range input.Photos {
			photo := models.ReviewPhoto{
				ReviewID: input.Review.ID,
				Image:    base64image,
			}
			if err := db.DB.Create(&photo).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save review photo"})
				return
			}
		}
	}

	// Return the created review
	c.JSON(http.StatusCreated, input.Review)
}

// validateReview validates the review data
func validateReview(review *models.Review) error {
	if review.Rating < 1 || review.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	if review.DeviceID == "" {
		return fmt.Errorf("device ID cannot be empty")
	}
	return nil
}

// checkLandmarkExists checks if the landmark exists
func checkLandmarkExists(landmarkID uint) error {
	var landmark models.Landmark
	if err := db.DB.First(&landmark, landmarkID).Error; err != nil {
		return fmt.Errorf("landmark ID does not exist")
	}
	return nil
}

// GetReviews retrieves all reviews including their photos
func GetReviews(c *gin.Context) {
	var reviews []models.Review

	// Preload photos for each review
	if err := db.DB.Preload("Photos").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reviews"})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

// GetReviewByID returns a review by ID
func GetReviewByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var review models.Review
	if err := db.DB.Preload("Photos").First(&review, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve review"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// GetReviewsByLandmarkID retrieves all reviews for a specific landmark based on its ID
func GetReviewsByLandmarkID(c *gin.Context) {
	landmarkID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid landmark ID"})
		return
	}

	var landmark models.Landmark
	if err := db.DB.First(&landmark, landmarkID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark ID does not exist"})
		return
	}

	var reviews []models.Review
	if err := db.DB.Where("landmark_id = ?", landmarkID).Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reviews"})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

// GetReviewCountByLandmarkID retrieves the count of reviews for a specific landmark based on its ID
func GetReviewCountByLandmarkID(c *gin.Context) {
	landmarkID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid landmark ID"})
		return
	}

	var landmark models.Landmark
	if err := db.DB.First(&landmark, landmarkID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark ID does not exist"})
		return
	}

	var count int64
	if err := db.DB.Model(&models.Review{}).Where("landmark_id = ?", landmarkID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"review_count": count})
}

// GetAverageRatingByLandmarkID calculates the average rating for a specific landmark
func GetAverageRatingByLandmarkID(c *gin.Context) {
	landmarkID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid landmark ID"})
		return
	}

	var landmark models.Landmark
	if err := db.DB.First(&landmark, landmarkID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark ID does not exist"})
		return
	}

	var result struct {
		AverageRating float64 `json:"average_rating"`
	}
	if err := db.DB.Model(&models.Review{}).
		Select("AVG(rating) as average_rating").
		Where("landmark_id = ?", landmarkID).
		Scan(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate average rating"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetReviewsByDeviceID retrieves all reviews by a user based on their DeviceID
func GetReviewsByDeviceID(c *gin.Context) {
	deviceID := c.Param("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	var reviews []models.Review
	if err := db.DB.Where("device_id = ?", deviceID).Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reviews"})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

// UpdateReview updates a review by ID
func UpdateReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var input struct {
		models.Review
		Photos []string `json:"photos"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review data"})
		return
	}

	// Validate review data
	if err := validateReview(&input.Review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if review exists
	var existingReview models.Review
	if err := db.DB.Preload("Photos").First(&existingReview, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	// Update review fields
	existingReview.DeviceID = input.DeviceID
	existingReview.Name = input.Name
	existingReview.Comment = input.Comment
	existingReview.Rating = input.Rating
	existingReview.LandmarkID = input.LandmarkID

	if err := db.DB.Save(&existingReview).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	// Handle photos update
	if len(input.Photos) > 0 {
		// Delete existing photos
		if err := db.DB.Where("review_id = ?", existingReview.ID).Delete(&models.ReviewPhoto{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review photos"})
			return
		}

		// Create new photos
		for _, base64image := range input.Photos {
			imageData, err := base64.StdEncoding.DecodeString(base64image)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode base64 image"})
				return
			}
			photo := models.ReviewPhoto{
				ReviewID: existingReview.ID,
				Image:    string(imageData),
			}
			if err := db.DB.Create(&photo).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review photos"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, existingReview)
}

// DeleteReview deletes a review by ID
func DeleteReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var review models.Review
	if err := db.DB.First(&review, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review"})
		return
	}

	if err := db.DB.Delete(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review"})
		return
	}

	c.Status(http.StatusNoContent)
}

// SearchReviews retrieves reviews matching the keyword in name, comment, or rating
func SearchReviews(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keyword is required"})
		return
	}

	var reviews []models.Review
	if err := db.DB.Where("name LIKE ? OR comment LIKE ? OR CAST(rating AS CHAR) LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search reviews"})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

func FilterReviews(c *gin.Context) {
	var reviews []models.Review

	// Define the query parameters and their corresponding SQL clauses
	params := map[string]string{
		"min_rating": "rating >= ?",
		"max_rating": "rating <= ?",
	}

	query := db.DB
	for param, clause := range params {
		if value := c.Query(param); value != "" {
			query = query.Where(clause, value)
		}
	}

	if err := query.Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reviews"})
		return
	}

	if len(reviews) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No review found"})
		return
	}

	c.JSON(http.StatusOK, reviews)
}
