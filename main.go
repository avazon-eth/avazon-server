package main

import (
	"avazon-api/controllers"
	"avazon-api/middleware"
	"avazon-api/models"
	"avazon-api/services"
	"avazon-api/tools"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
		&models.SystemPrompt{},
		&models.SystemPromptUsage{},
		&models.User{},
		&models.AvatarCreation{},
		&models.AvatarCharacterCreation{},
		&models.AvatarVoiceCreation{},
		&models.AvatarImageCreation{},
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
	if os.Getenv("PROFILE") == "local" || os.Getenv("PROFILE") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	DB := InitDB()
	r := gin.Default()
	InitCORS(r)

	// Init JWT keys
	err := middleware.InitKeys()
	if err != nil {
		panic(err)
	}

	// ======= Tools =======
	// 1. keys
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		panic("OPENAI_API_KEY is not set")
	}

	// ======= System Prompt Domain =======
	// system prompts
	systemPromptService := services.NewSystemPromptService(
		DB,
		func() tools.Assistant {
			return tools.NewOpenAIAssistant(openAIKey, "gpt-4o")
		},
	)
	systemPromptController := controllers.NewSystemPromptController(systemPromptService)
	systemPromptRG := r.Group("/system/prompts")
	systemPromptRG.Use(middleware.JWTAuthMiddleware("admin"))
	{
		systemPromptRG.POST("/:prompt_id", systemPromptController.CreateSystemPrompt)
		systemPromptRG.GET("/", systemPromptController.GetAllSystemPrompts)
		systemPromptRG.DELETE("/:prompt_id", systemPromptController.DeleteSystemPrompt)
		systemPromptRG.POST("/usages/:agent_id/use/:prompt_id", systemPromptController.UpdateSystemPromptUsage)
		systemPromptRG.GET("/usages", systemPromptController.GetAllSystemPromptUsages)
		systemPromptRG.DELETE("/usages/:agent_id", systemPromptController.DeleteSystemPromptUsage)
	}

	// ======= User Domain =======
	userService := services.NewUserService(DB)
	userController := controllers.NewUserController(userService)
	userRG := r.Group("/users")
	userRG.POST("/oauth2/:provider", userController.OAuth2Login)
	userRG.Use(middleware.JWTAuthMiddleware("user", "admin"))
	{
		userRG.GET("/me", userController.GetMyInfo)
	}
	r.POST("/users/token/refresh", middleware.JWTAuthMiddleware("refresh"), userController.RefreshToken)

	// ======= Avatar Domain =======
	// avatar creation
	avatarCreationService := services.NewAvatarCreationService(DB)
	avatarCreationController := controllers.NewAvatarCreationController(avatarCreationService)
	avatarCreateRG := r.Group("/avatar/create")
	avatarCreateRG.Use(middleware.JWTAuthMiddleware())
	{
		avatarCreateRG.POST("/new", avatarCreationController.StartCreation)
		avatarCreateRG.GET("/:id", avatarCreationController.GetOneSession)
		avatarCreateRG.POST("/:id", avatarCreationController.CreateAvatar)
	}
	avatarCreateRG.GET("/:id/enter/", avatarCreationController.EnterSession) // WebSocket

	r.Run(":8080")
}
