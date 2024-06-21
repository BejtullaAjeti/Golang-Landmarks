package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"
	"strings"
)

// CreateReviewPhoto creates a new review photo
func CreateReviewPhoto(c *gin.Context) {
	var input models.ReviewPhoto
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review photo data"})
		return
	}

	// Decode base64 image data
	imageData, err := decodeBase64Image(input.Image)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Image = imageData

	if err := db.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review photo"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

// GetAllReviewPhotos retrieves all review photos
func GetAllReviewPhotos(c *gin.Context) {
	var photos []models.ReviewPhoto

	if err := db.DB.Find(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve review photos"})
		return
	}

	c.JSON(http.StatusOK, photos)
}

// GetReviewPhotoByID retrieves a review photo by ID
func GetReviewPhotoByID(c *gin.Context) {
	var photo models.ReviewPhoto
	id := c.Param("id")

	if err := db.DB.First(&photo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review photo not found"})
		return
	}

	c.JSON(http.StatusOK, photo)
}

// UpdateReviewPhoto updates a review photo by ID
func UpdateReviewPhoto(c *gin.Context) {
	var photo models.ReviewPhoto
	id := c.Param("id")

	if err := db.DB.First(&photo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review photo not found"})
		return
	}

	var input models.ReviewPhoto
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review photo data"})
		return
	}

	// Decode base64 image data if provided
	if input.Image != "" {
		imageData, err := decodeBase64Image(input.Image)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		photo.Image = imageData
	}

	if err := db.DB.Save(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review photo"})
		return
	}

	c.JSON(http.StatusOK, photo)
}

// DeleteReviewPhoto deletes a review photo by ID
func DeleteReviewPhoto(c *gin.Context) {
	var photo models.ReviewPhoto
	id := c.Param("id")

	if err := db.DB.First(&photo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review photo not found"})
		return
	}

	if err := db.DB.Delete(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review photo"})
		return
	}

	c.Status(http.StatusNoContent)
}

// Helper function to decode base64 image data
func decodeBase64Image(encodedImage string) (string, error) {
	parts := strings.Split(encodedImage, ",")
	if len(parts) != 2 {
		return "", errors.New("invalid base64 image data")
	}

	// Read base64 encoded image data
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("base64 decoding error: %v", err)
	}

	return string(decoded), nil
}
