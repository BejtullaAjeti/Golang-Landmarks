package handlers

import (
	"context"
	"fmt"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

var s3Client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config, %v", err))
	}
	s3Client = s3.NewFromConfig(cfg)
}

// CreateLandmarkPhoto handles the upload and storage of landmark photos to S3
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

	// Read the file into memory
	fileBytes, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read photo file"})
		return
	}
	defer func(fileBytes multipart.File) {
		err := fileBytes.Close()
		if err != nil {
		}
	}(fileBytes)

	// Generate a unique file name based on landmark name and current time
	landmarkName := strings.ReplaceAll(landmark.Name, " ", "_")
	fileName := fmt.Sprintf("%s_%d%s", landmarkName, time.Now().UnixNano(), filepath.Ext(file.Filename))

	// Upload file to S3 with directory structure
	uploadPath := fmt.Sprintf("%s/%s", landmarkName, fileName)
	uploadErr := uploadFileToS3(fileBytes, uploadPath)
	if uploadErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo to S3"})
		return
	}

	// Construct the S3 URL for the uploaded file
	s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", "golang-backend-photos", uploadPath)

	// Create a new LandmarkPhoto record
	photo := models.LandmarkPhoto{
		LandmarkID: landmark.ID,
		Name:       file.Filename,
		Path:       s3URL,
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
func GetLandmarkPhotosByLandmarkID(c *gin.Context) {
	landmarkID := c.Param("id")

	var photos []models.LandmarkPhoto
	if err := db.DB.Where("landmark_id = ?", landmarkID).Find(&photos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve landmark photos"})
		return
	}

	if len(photos) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No photos found for the specified landmark"})
		return
	}

	photoLinks := make([]string, len(photos))
	for i, photo := range photos {
		photoLinks[i] = photo.Path
	}

	c.JSON(http.StatusOK, gin.H{"photo_links": photoLinks})
}
