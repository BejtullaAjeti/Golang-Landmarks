package main

import (
	"landmarksmodule/db"
	"landmarksmodule/routes"
)

func main() {
	db.Init()
	routes.SetupRoutes()

}
