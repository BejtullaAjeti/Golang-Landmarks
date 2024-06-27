package handlers

import (
	"fmt"
	"landmarksmodule/db"
	"landmarksmodule/models"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateLandmark handles creating a new landmark with attached photos
func CreateLandmark(c *gin.Context) {
	var input struct {
		models.Landmark
		Photos []models.LandmarkPhoto `json:"photos"`
	}

	// Check if content type is JSON
	if c.Request.Header.Get("Content-Type") == "application/json" {
		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid landmark data"})
			return
		}
	} else {
		// Handle form-data
		input.Name = c.PostForm("name")
		input.Type = c.PostForm("type")
		input.Information = c.PostForm("information")
		input.Description = c.PostForm("description")
		input.Latitude = c.PostForm("latitude")
		input.Longitude = c.PostForm("longitude")

		cityID, err := strconv.ParseUint(c.PostForm("city_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid city ID"})
			return
		}
		input.CityID = uint(cityID)

		// Handle photos
		var photos []models.LandmarkPhoto
		formFiles := c.Request.MultipartForm.File["photos"]
		for _, fileHeader := range formFiles {
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open photo file"})
				return
			}
			defer file.Close()

			// Count the number of photos already associated with the landmark
			var existingPhotoCount int64
			if err := db.DB.Model(&models.LandmarkPhoto{}).Where("landmark_id = ?", input.Landmark.ID).Count(&existingPhotoCount).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count photos for landmark"})
				return
			}

			// Increment to get the next number
			photoCount := existingPhotoCount + 1

			// Create directory for photos if not exists
			landmarkName := strings.ReplaceAll(input.Landmark.Name, " ", "_")
			dirPath := filepath.Join("uploads", landmarkName)
			if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory for landmark"})
				return
			}

			// Generate a unique file name
			fileName := fmt.Sprintf("%s%d%s", landmarkName, photoCount, filepath.Ext(fileHeader.Filename))
			filePath := filepath.Join(dirPath, fileName)

			// Check if file with the same name exists and generate a new name if it does
			for fileExists(filePath) {
				photoCount++
				fileName = fmt.Sprintf("%s%d%s", landmarkName, photoCount, filepath.Ext(fileHeader.Filename))
				filePath = filepath.Join(dirPath, fileName)
			}

			// Save file to disk
			if err := c.SaveUploadedFile(fileHeader, filePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
				return
			}

			// Create LandmarkPhoto record
			photo := models.LandmarkPhoto{
				Name:      fileHeader.Filename,
				Path:      filePath,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			photos = append(photos, photo)
		}
		input.Photos = photos
	}

	// Validate city existence
	var city models.City
	if err := db.DB.First(&city, input.CityID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City with the specified city_id does not exist"})
		return
	}

	// Create the landmark
	if err := db.DB.Create(&input.Landmark).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create landmark"})
		return
	}

	// Create landmark photos
	for i, photo := range input.Photos {
		photo.LandmarkID = input.Landmark.ID
		if err := db.DB.Create(&photo).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create landmark photo"})
			return
		}
		// Update the input.Photos slice with the created photo data
		input.Photos[i] = photo
	}

	// Update the input struct with the created photos and respond with the landmark data
	input.Landmark.Photos = input.Photos
	c.JSON(http.StatusCreated, input.Landmark)
}

// fileExists checks if a file exists at the given path
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetLandmarks retrieves all landmarks including their photos
func GetLandmarks(c *gin.Context) {
	var landmarks []models.Landmark

	// Preload photos for each landmark
	if err := db.DB.Preload("Photos").Find(&landmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve landmarks"})
		return
	}

	c.JSON(http.StatusOK, landmarks)
}

func GetLandmarkByID(c *gin.Context) {
	var landmark models.Landmark
	id := c.Param("id")
	log.Printf("Fetching landmark with ID: %s", id)
	if err := db.DB.Preload("Photos").First(&landmark, id).Error; err != nil {
		log.Println("Landmark not found or an error occurred:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Landmark not found"})
		return
	}

	c.JSON(http.StatusOK, landmark)
}

func GetLandmarkDetails(c *gin.Context) {
	var (
		landmark      models.Landmark
		reviews       []models.Review
		reviewCount   int64
		averageRating float64
		userReview    models.Review // User review for the landmark
	)

	landmarkID := c.Param("id")
	deviceID := c.Query("device_id") // Get device ID from query parameter

	log.Printf("Fetching details for landmark with ID: %s", landmarkID)

	// Fetch Landmark with Photos
	if err := db.DB.Preload("Photos").First(&landmark, landmarkID).Error; err != nil {
		log.Println("Landmark not found or an error occurred:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Landmark not found"})
		return
	}

	// Fetch Reviews
	db.DB.Where("landmark_id = ?", landmarkID).Find(&reviews)

	// Fetch Review Count
	db.DB.Model(&models.Review{}).Where("landmark_id = ?", landmarkID).Count(&reviewCount)

	// Fetch Average Rating
	var result struct {
		AverageRating float64
	}
	db.DB.Model(&models.Review{}).
		Select("AVG(rating) as average_rating").
		Where("landmark_id = ?", landmarkID).
		Scan(&result)
	averageRating = result.AverageRating

	// If device ID is provided, fetch reviews by device ID for the landmark
	if deviceID != "" {
		var userReviews []models.Review
		db.DB.Where("landmark_id = ? AND device_id = ?", landmarkID, deviceID).Find(&userReviews)
		if len(userReviews) > 0 {
			userReview = userReviews[0] // Assuming there's only one review per user for a landmark
		}
	}

	// Construct JSON response
	response := gin.H{
		"landmark":       landmark,
		"reviews":        reviews,
		"review_count":   reviewCount,
		"average_rating": averageRating,
	}

	// Include the user's review only if device ID is provided
	if deviceID != "" {
		response["user_review"] = userReview
	}

	// Send JSON response
	c.JSON(http.StatusOK, response)
}

// UpdateLandmark updates a landmark by ID
func UpdateLandmark(c *gin.Context) {
	id := c.Param("id")
	var landmark models.Landmark

	// Check if the landmark exists
	if err := db.DB.First(&landmark, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Landmark not found"})
		return
	}

	// Bind updated landmark data from JSON request body or form-data
	if c.Request.Header.Get("Content-Type") == "application/json" {
		if err := c.BindJSON(&landmark); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
			return
		}
	} else {
		// Handle form-data
		landmark.Name = c.PostForm("name")
		landmark.Type = c.PostForm("type")
		landmark.Information = c.PostForm("information")
		landmark.Description = c.PostForm("description")
		landmark.Latitude = c.PostForm("latitude")
		landmark.Longitude = c.PostForm("longitude")

		cityID, err := strconv.ParseUint(c.PostForm("city_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid city ID"})
			return
		}
		landmark.CityID = uint(cityID)
	}

	// Check if the city exists
	var city models.City
	if err := db.DB.First(&city, landmark.CityID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City with the specified city_id does not exist"})
		return
	}

	// Save updated landmark data
	if err := db.DB.Save(&landmark).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update landmark"})
		return
	}

	c.JSON(http.StatusOK, landmark)
}

func DeleteLandmark(c *gin.Context) {
	var landmark models.Landmark
	id := c.Param("id")

	if err := db.DB.First(&landmark, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "landmark not found"})
		return
	}

	db.DB.Delete(&landmark)

	c.Status(http.StatusNoContent)
}

func SearchLandmarks(c *gin.Context) {
	var landmarks []models.Landmark
	keyword := c.Query("keyword")

	db.DB.Where("name LIKE ?", "%"+keyword+"%").
		Or("description LIKE ?", "%"+keyword+"%").
		Find(&landmarks)

	c.JSON(http.StatusOK, landmarks)
}

func FilterLandmarks(c *gin.Context) {
	var landmarks []models.Landmark

	// Define the query parameters and their corresponding SQL clauses
	params := map[string]string{
		"city_id":       "city_id = ?",
		"type":          "type = ?",
		"min_latitude":  "CAST(latitude AS DECIMAL) >= ?",
		"max_latitude":  "CAST(latitude AS DECIMAL) <= ?",
		"min_longitude": "CAST(longitude AS DECIMAL) >= ?",
		"max_longitude": "CAST(longitude AS DECIMAL) <= ?",
	}

	query := db.DB
	for param, clause := range params {
		if value := c.Query(param); value != "" {
			query = query.Where(clause, value)
		}
	}

	if err := query.Find(&landmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve landmarks"})
		return
	}

	if len(landmarks) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No landmarks found"})
		return
	}

	c.JSON(http.StatusOK, landmarks)
}

func GetAllLandmarksOfCity(c *gin.Context) {
	var landmarks []models.Landmark
	cityID := c.Param("city_id")

	var city models.City
	if err := db.DB.First(&city, cityID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City with the specified city_id does not exist"})
		return
	}

	db.DB.Where("city_id = ?", cityID).Find(&landmarks)

	c.JSON(http.StatusOK, landmarks)
}

func GetAllLandmarksOfRegion(c *gin.Context) {
	var landmarks []models.Landmark
	regionID := c.Param("region_id")

	var region models.Region
	if err := db.DB.First(&region, regionID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Region with the specified region_id does not exist"})
		return
	}

	db.DB.Joins("JOIN cities ON landmarks.city_id = cities.id").
		Where("cities.region_id = ?", regionID).
		Find(&landmarks)

	c.JSON(http.StatusOK, landmarks)
}
