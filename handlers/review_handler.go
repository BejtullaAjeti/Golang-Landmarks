package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CreateReview handles creating a new review
func CreateReview(c *gin.Context) {
	var input models.Review

	// Parse JSON input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review data"})
		return
	}

	// Validate review data
	if err := validateReview(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if landmark exists
	if err := checkLandmarkExists(input.LandmarkID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the review
	if err := db.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func CreateReviewWithPhotos(c *gin.Context) {
	// Extract JSON part from the form data
	reviewJSON := c.PostForm("review")
	var input models.Review

	// Parse JSON input for review
	if err := json.Unmarshal([]byte(reviewJSON), &input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review data"})
		return
	}

	// Validate review data
	if err := validateReview(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if landmark exists
	if err := checkLandmarkExists(input.LandmarkID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the review
	if err := db.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	// Process photos if any
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	files := form.File["photos"]
	if files != nil {
		for _, file := range files {
			// Read the file into memory
			fileBytes, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read photo file"})
				return
			}

			var landmark models.Landmark
			if err := db.DB.First(&landmark, input.LandmarkID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark associated with the review not found"})
				return
			}

			// Generate a unique file name based on landmark name and current time
			deviceId := strings.ReplaceAll(input.DeviceID, " ", "_")
			landmarkName := strings.ReplaceAll(landmark.Name, " ", "_")
			fileName := fmt.Sprintf("%s_%s_%d%s", deviceId, landmarkName, time.Now().UnixNano(), filepath.Ext(file.Filename))

			// Upload file to S3 with directory structure
			uploadPath := fmt.Sprintf("%s/%s/%s", deviceId, landmarkName, fileName)
			uploadErr := uploadFileToS3(fileBytes, uploadPath)
			if uploadErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo to S3"})
				return
			}
			err = fileBytes.Close()
			if err != nil {
				return
			}
			// Construct the S3 URL for the uploaded file
			s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", "golang-backend-photos", uploadPath)

			// Create a new ReviewPhoto record
			photo := models.ReviewPhoto{
				ReviewID:  input.ID,
				Name:      file.Filename,
				Path:      s3URL,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := db.DB.Create(&photo).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review photo"})
				return
			}
		}
	}

	c.JSON(http.StatusCreated, input)
}

// validateReview validates the review data
func validateReview(review *models.Review) error {
	if review.Rating < 1 || review.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	if review.DeviceID == "" {
		return fmt.Errorf("device ID cannot be empty")
	}
	if review.Name == "" {
		return fmt.Errorf("name cannot be empty")
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

func GetReviews(c *gin.Context) {
	var reviews []models.Review

	// Retrieve reviews from the database
	if err := db.DB.Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reviews"})
		return
	}

	// Populate PhotoLinks for each review
	for i := range reviews {
		var photos []models.ReviewPhoto
		if err := db.DB.Where("review_id = ?", reviews[i].ID).Find(&photos).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve photos for review"})
			return
		}

		var photoLinks []string
		for _, photo := range photos {
			photoLinks = append(photoLinks, photo.Path)
		}

		reviews[i].Photos = nil // Clear Photos field
		reviews[i].PhotoLinks = photoLinks
	}

	c.JSON(http.StatusOK, reviews)
}

func GetReviewByID(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve review"})
		return
	}

	var photos []models.ReviewPhoto
	if err := db.DB.Where("review_id = ?", review.ID).Find(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve photos for review"})
		return
	}

	var photoLinks []string
	for _, photo := range photos {
		photoLinks = append(photoLinks, photo.Path)
	}

	review.Photos = nil // Clear Photos field
	review.PhotoLinks = photoLinks

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
	id := c.Param("id")
	var review models.Review

	if err := db.DB.First(&review, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	if err := c.ShouldBindJSON(&review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}

	// Save updated landmark data
	if err := db.DB.Save(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}

	c.JSON(http.StatusOK, review)
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

	// Delete associated review photos
	if err := db.DB.Where("review_id = ?", review.ID).Delete(&models.ReviewPhoto{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review photos"})
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
