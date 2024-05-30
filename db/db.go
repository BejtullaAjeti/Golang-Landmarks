package db

import (
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

func migrate() {
	DB.AutoMigrate(&models.Region{}, &models.City{}, &models.Landmark{}, &models.Review{})
	DB.Model(&models.City{}).AddForeignKey("region_id", "regions(id)", "RESTRICT", "RESTRICT")
	DB.Model(&models.Landmark{}).AddForeignKey("city_id", "cities(id)", "RESTRICT", "RESTRICT")
	DB.Model(&models.Review{}).AddForeignKey("landmark_id", "landmarks(id)", "RESTRICT", "RESTRICT")
	fmt.Println("Database migrated successfully")
}
