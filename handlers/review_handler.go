package handlers

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"io"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateReview handles creating a new review with attached photos
func CreateReview(c *gin.Context) {
	var input struct {
		models.Review
		Photos []models.ReviewPhoto `json:"photos"`
	}

	// Check if content type is JSON
	if c.Request.Header.Get("Content-Type") == "application/json" {
		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review data"})
			return
		}
	} else {
		// Handle form-data
		input.DeviceID = c.PostForm("device_id")
		input.Name = c.PostForm("name")
		input.Comment = c.PostForm("comment")
		rating, err := strconv.Atoi(c.PostForm("rating"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating"})
			return
		}
		input.Rating = rating

		landmarkID, err := strconv.ParseUint(c.PostForm("landmark_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid landmark ID"})
			return
		}
		input.LandmarkID = uint(landmarkID)

		// Retrieve landmark details from the database
		var landmark models.Landmark
		if err := db.DB.First(&landmark, input.LandmarkID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark with the specified landmark_id does not exist"})
			return
		}

		// Use landmark name for directory and filename generation
		landmarkName := strings.ReplaceAll(landmark.Name, " ", "_")
		if landmarkName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark name is required"})
			return
		}

		// Handle photos
		var photos []models.ReviewPhoto
		formFiles := c.Request.MultipartForm.File["photos"]
		for _, fileHeader := range formFiles {
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open photo file"})
				return
			}
			defer file.Close()

			// Create directories for photos if not exists
			photoFolder := filepath.Join("review_photos", input.DeviceID, landmarkName)
			if err := os.MkdirAll(photoFolder, os.ModePerm); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create photo directory"})
				return
			}

			// Generate a unique file name
			fileName := fmt.Sprintf("%s_%d%s", landmarkName, time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
			filePath := filepath.Join(photoFolder, fileName)

			// Check if file with the same name exists and generate a new name if it does
			for fileExistsReview(filePath) {
				fileName = fmt.Sprintf("%s_%d%s", landmarkName, time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
				filePath = filepath.Join(photoFolder, fileName)
			}

			// Save file to disk
			if err := c.SaveUploadedFile(fileHeader, filePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
				return
			}

			// Create ReviewPhoto record
			photo := models.ReviewPhoto{
				ReviewID:  input.Review.ID,
				Name:      fileHeader.Filename,
				Path:      filePath,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			photos = append(photos, photo)
		}
		input.Photos = photos
	}

	// Create the review
	if err := db.DB.Create(&input.Review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	// Create review photos
	for i, photo := range input.Photos {
		photo.ReviewID = input.Review.ID
		if err := db.DB.Create(&photo).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review photo"})
			return
		}
		// Update the input.Photos slice with the created photo data
		input.Photos[i] = photo
	}

	// Update the input struct with the created photos and respond with the review data
	input.Review.Photos = input.Photos
	c.JSON(http.StatusCreated, input.Review)
}

// fileExists checks if a file exists at the given path
func fileExistsReview(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
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

	var input models.Review

	// Check if content type is JSON
	if c.Request.Header.Get("Content-Type") == "application/json" {
		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review data"})
			return
		}
	} else {
		// Handle form-data
		input.DeviceID = c.PostForm("device_id")
		input.Name = c.PostForm("name")
		input.Comment = c.PostForm("comment")
		rating, err := strconv.Atoi(c.PostForm("rating"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating"})
			return
		}
		input.Rating = rating

		landmarkID, err := strconv.ParseUint(c.PostForm("landmark_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid landmark ID"})
			return
		}
		input.LandmarkID = uint(landmarkID)
	}

	// Validate review data
	if err := validateReview(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if review exists
	var existingReview models.Review
	if err := db.DB.First(&existingReview, id).Error; err != nil {
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

	c.JSON(http.StatusOK, existingReview)
}

// extractLandmarkNameFromPath extracts the landmark name from the given photo path
func extractLandmarkNameFromPath(photoPath string) string {
	parts := strings.Split(photoPath, string(filepath.Separator))
	if len(parts) < 3 {
		return ""
	}
	return parts[len(parts)-2]
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
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
