package handlers

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/net/context"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config, %v", err))
	}
	s3Client = s3.NewFromConfig(cfg)
}

// CreateReviewPhoto handles the upload and storage of review photos to S3
func CreateReviewPhoto(c *gin.Context) {
	reviewID := c.PostForm("review_id")
	var review models.Review
	if err := db.DB.First(&review, reviewID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Review with the specified review_id does not exist"})
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

	var landmark models.Landmark
	if err := db.DB.First(&landmark, review.LandmarkID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Landmark associated with the review not found"})
		return
	}
	// Generate a unique file name based on landmark name and current time
	deviceId := strings.ReplaceAll(review.DeviceID, " ", "_")
	landmarkName := strings.ReplaceAll(landmark.Name, " ", "_")
	fileName := fmt.Sprintf("%s_%s_%d%s", deviceId, landmarkName, time.Now().UnixNano(), filepath.Ext(file.Filename))

	// Upload file to S3 with directory structure
	uploadPath := fmt.Sprintf("%s/%s/%s", deviceId, landmarkName, fileName)
	uploadErr := uploadFileToS3(fileBytes, uploadPath)
	if uploadErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo to S3"})
		return
	}

	// Construct the S3 URL for the uploaded file
	s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", "golang-backend-photos", uploadPath)

	// Create a new LandmarkPhoto record
	photo := models.ReviewPhoto{
		ReviewID:  review.ID,
		Name:      file.Filename,
		Path:      s3URL,
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
