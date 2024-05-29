package routes

import (
	"landmarksmodule/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes initializes routes for the API
func SetupRoutes() {
	router := gin.Default()

	// Regions endpoints
	router.GET("/regions", handlers.GetRegions)
	router.POST("/regions", handlers.CreateRegion)
	router.GET("/regions/:id", handlers.GetRegionByID)
	router.PUT("/regions/:id", handlers.UpdateRegion)
	router.DELETE("/regions/:id", handlers.DeleteRegion)
	router.GET("/regions/search", handlers.SearchRegions)
	router.GET("/regions/filter", handlers.FilterRegions)

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
	router.GET("/landmarks/search", handlers.SearchLandmarks)
	router.GET("/landmarks/filter", handlers.FilterLandmarks)
	router.GET("/landmarks/city/:city_id", handlers.GetAllLandmarksOfCity)
	router.GET("landmarks/region/:region_id", handlers.GetAllLandmarksOfRegion)

	// Start server
	router.Run(":8080")
}
