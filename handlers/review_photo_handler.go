package handlers

import (
	"fmt"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateReviewPhoto handles the upload and storage of review photos
func CreateReviewPhoto(c *gin.Context) {
	deviceID := c.PostForm("device_id")
	var review models.Review
	if err := db.DB.First(&review, c.PostForm("review_id")).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Review with the specified review_id does not exist"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo data"})
		return
	}

	// Count the number of photos already associated with the review
	var photoCount int64
	db.DB.Model(&models.ReviewPhoto{}).Where("review_id = ?", review.ID).Count(&photoCount)
	photoCount++ // Increment to get the next number

	// Create a unique file name based on review ID, name, and photo count
	dirPath := filepath.Join("review_uploads", deviceID)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory for review"})
		return
	}

	fileName := fmt.Sprintf("%d_%d%s", review.ID, photoCount, filepath.Ext(file.Filename))
	filePath := filepath.Join(dirPath, fileName)

	// Save the file to the disk
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
		return
	}

	// Create a new ReviewPhoto record
	photo := models.ReviewPhoto{
		ReviewID:  review.ID,
		Name:      file.Filename,
		Path:      filePath,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := db.DB.Create(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review photo"})
		return
	}

	c.JSON(http.StatusCreated, photo)
}

func GetAllReviewPhotos(c *gin.Context) {
	var photos []models.ReviewPhoto
	if err := db.DB.Find(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve review photos"})
		return
	}

	c.JSON(http.StatusOK, photos)
}

func GetReviewPhotoByID(c *gin.Context) {
	id := c.Param("id")
	var photo models.ReviewPhoto
	if err := db.DB.First(&photo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review photo not found"})
		return
	}

	c.JSON(http.StatusOK, photo)
}

func UpdateReviewPhoto(c *gin.Context) {
	id := c.Param("id")
	var photo models.ReviewPhoto

	if err := db.DB.First(&photo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review photo not found"})
		return
	}

	var input models.ReviewPhoto
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo data"})
		return
	}

	// Update fields
	photo.Name = input.Name
	photo.Path = input.Path
	photo.UpdatedAt = time.Now()

	if err := db.DB.Save(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review photo"})
		return
	}

	c.JSON(http.StatusOK, photo)
}

func DeleteReviewPhoto(c *gin.Context) {
	id := c.Param("id")
	var photo models.ReviewPhoto

	if err := db.DB.First(&photo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review photo not found"})
		return
	}

	if err := db.DB.Delete(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review photo"})
		return
	}

	// Optionally, delete the file from the disk
	if err := os.Remove(photo.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete photo file"})
		return
	}

	c.Status(http.StatusNoContent)
}
