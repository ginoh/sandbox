package main

import (
	"log"

	"example.com/di"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := gin.Default()

	// dependency injection
	err := di.Setup()
	if err != nil {
		log.Fatal(err)
	}
	defer di.TearDown()

	user := di.InitUser()
	r.POST("/users", user.Create)
	r.GET("/users/:id", user.FindByID)
	r.PUT("/users/:id", user.Update)
	r.DELETE("/users/:id", user.Delete)

	// Listen :8080 unless a PORT environment variable was defined.
	r.Run()
}
