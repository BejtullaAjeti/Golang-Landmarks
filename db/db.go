package db

import (
	"errors"
	"fmt"

	"landmarksmodule/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open("mysql", "root@tcp(localhost:3306)/landmarks?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}

	// Enable debug mode
	DB.LogMode(true)

	// Run auto migration
	migrate()
}

func GetGeoJSONByRegionID(regionID uint) (*models.GeoJSON, error) {
	var geoJSON models.GeoJSON
	result := DB.Where("region_id = ?", regionID).First(&geoJSON)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if no GeoJSON found
		}
		return nil, result.Error
	}
	return &geoJSON, nil
}

func migrate() {
	DB.AutoMigrate(&models.Region{}, &models.City{}, &models.Landmark{}, &models.Review{}, &models.GeoJSON{})
	DB.Model(&models.City{}).AddForeignKey("region_id", "regions(id)", "RESTRICT", "RESTRICT")
	DB.Model(&models.Landmark{}).AddForeignKey("city_id", "cities(id)", "RESTRICT", "RESTRICT")
	DB.Model(&models.Review{}).AddForeignKey("landmark_id", "landmarks(id)", "RESTRICT", "RESTRICT")
	fmt.Println("Database migrated successfully")
}
