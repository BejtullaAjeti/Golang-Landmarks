package routes

import (
	"github.com/gin-gonic/gin"
	"landmarksmodule/handlers"
	"log"
)

// SetupRoutes initializes routes for the API
func SetupRoutes() {
	router := gin.Default()

	//Country endpoints
	router.GET("/countries", handlers.GetCountries)
	router.GET("/countries/:id", handlers.GetCountryByID)
	router.GET("countries/lookup", handlers.GetCountryByLatLong)
	router.POST("/countries", handlers.CreateCountry)
	router.PUT("/countries/:id", handlers.UpdateCountry)
	router.DELETE("/countries/:id", handlers.DeleteCountry)

	// Regions endpoints
	router.GET("/regions", handlers.GetRegions)
	router.POST("/regions", handlers.CreateRegion)
	router.GET("/regions/:id", handlers.GetRegionByID)
	router.PUT("/regions/:id", handlers.UpdateRegion)
	router.DELETE("/regions/:id", handlers.DeleteRegion)
	router.GET("/regions/search", handlers.SearchRegions)
	router.GET("/regions/filter", handlers.FilterRegions)
	router.POST("/regions/country", handlers.AddRegionsToCountry)

	// Cities endpoints
	router.GET("/cities", handlers.GetCities)
	router.POST("/cities", handlers.CreateCity)
	router.GET("/cities/:id", handlers.GetCityByID)
	router.PUT("/cities/:id", handlers.UpdateCity)
	router.DELETE("/cities/:id", handlers.DeleteCity)
	router.GET("/cities/search", handlers.SearchCities)
	router.GET("/cities/filter", handlers.FilterCities)
	router.GET("/cities/region", handlers.GetRegionOfCity)

	// Landmarks endpoints
	router.GET("/landmarks", handlers.GetLandmarks)
	router.POST("/landmarks", handlers.CreateLandmark)
	router.GET("/landmarks/:id", handlers.GetLandmarkByID)
	router.PUT("/landmarks/:id", handlers.UpdateLandmark)
	router.DELETE("/landmarks/:id", handlers.DeleteLandmark)
	router.GET("/landmarks/:id/average-rating", handlers.GetAverageRatingByLandmarkID)
	router.GET("/landmarks/:id/reviews", handlers.GetReviewsByLandmarkID)
	router.GET("/landmarks/:id/review-count", handlers.GetReviewCountByLandmarkID)
	router.GET("/landmarks/:id/details", handlers.GetLandmarkDetails)
	router.GET("/landmarks/search", handlers.SearchLandmarks)
	router.GET("/landmarks/filter", handlers.FilterLandmarks)
	router.GET("/landmarks/city/:city_id", handlers.GetAllLandmarksOfCity)
	router.GET("/landmarks/region/:region_id", handlers.GetAllLandmarksOfRegion)
	router.GET("/landmarks/:id/photos", handlers.GetLandmarkPhotosByLandmarkID)

	//Review endpoints
	router.GET("/reviews", handlers.GetReviews)
	router.POST("/reviewJSON", handlers.CreateReview)
	router.POST("/reviews", handlers.CreateReviewWithPhotos)
	router.GET("/reviews/:id", handlers.GetReviewByID)
	router.PUT("/reviews/:id", handlers.UpdateReview)
	router.DELETE("/reviews/:id", handlers.DeleteReview)
	router.GET("/reviews/user/:device_id", handlers.GetReviewsByDeviceID)
	router.GET("/reviews/search", handlers.SearchReviews)
	router.GET("/reviews/filter", handlers.FilterReviews)
	router.GET("/reviews/:id/photos", handlers.GetReviewPhotosByReviewID)

	//GeoJson endpoints
	router.GET("/geojson", handlers.GetAllGeoJSON)
	router.POST("/geojson/:region_id", handlers.CreateGeoJSON)
	router.PUT("/geojson/:id", handlers.UpdateGeoJSON)
	router.GET("/geojson/:id", handlers.GetGeoJSONFromDB)

	//Photo endpoints
	router.GET("/landmarkphotos", handlers.GetAllLandmarkPhotos)
	router.GET("/landmarkphotos/:id", handlers.GetLandmarkPhotoByID)
	router.POST("/landmarkphotos", handlers.CreateLandmarkPhoto)
	router.PUT("/landmarkphotos/:id", handlers.UpdateLandmarkPhoto)
	router.DELETE("/landmarkphotos/:id", handlers.DeleteLandmarkPhoto)

	router.GET("/reviewphotos", handlers.GetAllReviewPhotos)
	router.GET("/reviewphotos/:id", handlers.GetReviewPhotoByID)
	router.POST("/reviewphotos", handlers.CreateReviewPhoto)
	router.PUT("/reviewphotos/:id", handlers.UpdateReviewPhoto)
	router.DELETE("/reviewphotos/:id", handlers.DeleteReviewPhoto)

	// Start server
	err := router.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
