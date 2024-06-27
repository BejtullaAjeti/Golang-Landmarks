package handlers

import (
	"fmt"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateLandmarkPhoto handles the upload and storage of landmark photos
func CreateLandmarkPhoto(c *gin.Context) {
	landmarkID := c.PostForm("landmark_id")
	var landmark models.Landmark
	if err := db.DB.First(&landmark, landmarkID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark with the specified landmark_id does not exist"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo data"})
		return
	}

	// Create a unique file name based on landmark ID and name
	landmarkName := strings.ReplaceAll(landmark.Name, " ", "_")
	fileName := fmt.Sprintf("%d_%s%s", landmark.ID, landmarkName, filepath.Ext(file.Filename))
	filePath := filepath.Join("uploads", fileName)

	// Save the file to the disk
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
		return
	}

	// Create a new LandmarkPhoto record
	photo := models.LandmarkPhoto{
		LandmarkID: landmark.ID,
		Name:       file.Filename,
		Path:       filePath,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := db.DB.Create(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create landmark photo"})
		return
	}

	c.JSON(http.StatusCreated, photo)
}

func GetAllLandmarkPhotos(c *gin.Context) {
	var photos []models.LandmarkPhoto
	if err := db.DB.Find(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve landmark photos"})
		return
	}

	c.JSON(http.StatusOK, photos)
}

func GetLandmarkPhotoByID(c *gin.Context) {
	id := c.Param("id")
	var photo models.LandmarkPhoto
	if err := db.DB.First(&photo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Landmark photo not found"})
		return
	}

	c.JSON(http.StatusOK, photo)
}

func UpdateLandmarkPhoto(c *gin.Context) {
	id := c.Param("id")
	var photo models.LandmarkPhoto

	if err := db.DB.First(&photo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Landmark photo not found"})
		return
	}

	var input models.LandmarkPhoto
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo data"})
		return
	}

	// Update fields
	photo.Name = input.Name
	photo.Path = input.Path
	photo.UpdatedAt = time.Now()

	if err := db.DB.Save(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update landmark photo"})
		return
	}

	c.JSON(http.StatusOK, photo)
}

func DeleteLandmarkPhoto(c *gin.Context) {
	id := c.Param("id")
	var photo models.LandmarkPhoto

	if err := db.DB.First(&photo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Landmark photo not found"})
		return
	}

	if err := db.DB.Delete(&photo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete landmark photo"})
		return
	}

	// Optionally, delete the file from the disk
	if err := os.Remove(photo.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete photo file"})
		return
	}

	c.Status(http.StatusNoContent)
}
