package main

import (
	"avazon-api/controllers"
	"avazon-api/middleware"
	"avazon-api/models"
	"avazon-api/services"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initLocalDB() *gorm.DB {
	DB, err := gorm.Open(sqlite.Open("./test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	return DB
}

func InitDB() *gorm.DB {
	DB := initLocalDB()
	err := DB.AutoMigrate(
		&models.User{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	return DB
}

func InitCORS(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8081", "https://gid.cast-ing.kr"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-OAuth2-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func main() {
	DB := InitDB()
	r := gin.Default()
	InitCORS(r)

	// Init JWT keys
	err := middleware.InitKeys()
	if err != nil {
		panic(err)
	}

	userService := services.NewUserService(DB)
	userController := controllers.NewUserController(userService)

	userRG := r.Group("/users")
	userRG.POST("/oauth2/:provider", userController.OAuth2Login)
	userRG.Use(middleware.JWTAuthMiddleware("user", "admin"))
	{
		userRG.GET("/me", userController.GetMyInfo)
	}
	userRG.Use(middleware.JWTAuthMiddleware("refresh"))
	{
		userRG.POST("/token/refresh", userController.RefreshToken)
	}

	r.Run(":8080")
}
