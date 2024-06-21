package handlers

import (
	"landmarksmodule/db"
	"landmarksmodule/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateLandmarkPhoto(c *gin.Context) {
	var input models.LandmarkPhoto

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo data"})
		return
	}

	// Set created and updated timestamps
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	if err := db.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create landmark photo"})
		return
	}

	c.JSON(http.StatusCreated, input)
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
	photo.Image = input.Image
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

	c.Status(http.StatusNoContent)
}
