package main

import (
	"avazon-api/controllers"
	"avazon-api/middleware"
	"avazon-api/models"
	"avazon-api/services"
	"avazon-api/tools"
	"fmt"
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
		&models.AvatarCreationChat{},
		&models.AvatarCharacterCreation{},
		&models.AvatarVoiceCreation{},
		&models.AvatarImageCreation{},
		&models.Avatar{},
		&models.AvatarMusic{},
		&models.AvatarVideo{},
		&models.AvatarMusicContentCreation{},
		&models.AvatarVideoContentCreation{},
		&models.AvatarImageRemix{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	return DB
}

func InitCORS(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8081", "http://localhost:5173", "https://gid.cast-ing.kr", "https://staging.d9xje8vs9f8su.amplifyapp.com"},
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
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	InitCORS(r)

	// Init JWT keys
	err := middleware.InitKeys()
	if err != nil {
		panic(err)
	}

	// Init validator
	models.InitValidator()

	// ======= Tools =======
	// 1. keys
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		panic("OPENAI_API_KEY is not set")
	}
	openArtKey := os.Getenv("OPENART_API_KEY")
	if openArtKey == "" {
		panic("OPENART_API_KEY is not set")
	}
	elevenLabsKey := os.Getenv("ELEVENLABS_API_KEY")
	if elevenLabsKey == "" {
		panic("ELEVENLABS_API_KEY is not set")
	}
	runwayKey := os.Getenv("RUNWAY_API_KEY")
	if runwayKey == "" {
		panic("RUNWAY_API_KEY is not set")
	}
	jenAIKey := os.Getenv("JENAI_API_KEY")
	if jenAIKey == "" {
		panic("JENAI_API_KEY is not set")
	}
	// 2. components
	s3Service, err := services.NewS3Service("avazon", "us-west-1")
	if err != nil {
		fmt.Println("Error initializing S3 service:", err)
		return
	}
	openAIPainter := tools.NewOpenAIPainter(openAIKey)
	openArtPainter := tools.NewOpenArtPainter(openArtKey)
	elevenLabsVoiceActor := tools.NewElevenLabsVoiceActor(elevenLabsKey)
	runwayVideoProducer := tools.NewRunwayVideoProducer(runwayKey)
	jenAIProducer := tools.NewJENAIProducer(jenAIKey)

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

	// ======= Web Data Session Domain =======
	webDataSessionService := services.NewWebDataSessionService()
	webDataSessionController := controllers.NewWebDataSessionController(webDataSessionService)
	webDataSessionRG := r.Group("/web-data-session")
	webDataSessionRG.POST("/token/fetch", webDataSessionController.GetToken)
	webDataSessionRG.Use(middleware.JWTAuthMiddleware())
	{
		// 1. put token
		webDataSessionRG.PUT("/token", webDataSessionController.PutToken)
		// 2. put data
		webDataSessionRG.PUT("/data", webDataSessionController.PutData)
		// 3. get data
		webDataSessionRG.POST("/data/fetch", webDataSessionController.GetData)
		// 4. clear data
		webDataSessionRG.DELETE("/data", webDataSessionController.ClearData)
		// 5. clear token
		webDataSessionRG.DELETE("/token", webDataSessionController.ClearToken)
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
	// ** Avatar Creation API **
	avatarCreationService := services.NewAvatarCreateService(
		DB,
		func() tools.Assistant {
			return tools.NewOpenAIAssistant(openAIKey, "gpt-4o")
		},
		systemPromptService,
		openArtPainter,
		elevenLabsVoiceActor,
		runwayVideoProducer,
		s3Service,
	)
	avatarCreationController := controllers.NewAvatarCreationController(avatarCreationService)
	avatarCreateRG := r.Group("/avatar/create")
	avatarCreateRG.GET("/:creation_id/enter/", avatarCreationController.EnterSession) // Websocket exchange
	avatarCreateRG.GET("/:creation_id/enter", avatarCreationController.EnterSession)  // Websocket exchange
	avatarCreateRG.Use(middleware.JWTAuthMiddleware())
	{
		avatarCreateRG.POST("/new", avatarCreationController.StartCreation)
		avatarCreateRG.GET("/:creation_id", avatarCreationController.GetOneSession)
		avatarCreateRG.POST("/:creation_id", avatarCreationController.CreateAvatar)
		// also has RESTful interface
		avatarCreateRG.POST("/:creation_id/image", avatarCreationController.CreateAvatarImage)
		avatarCreateRG.POST("/:creation_id/character", avatarCreationController.CreateAvatarCharacter)
		avatarCreateRG.POST("/:creation_id/voice", avatarCreationController.CreateAvatarVoice)
	}

	// ** Avatar Public API **
	avatarService := services.NewAvatarService(DB)
	avatarController := controllers.NewAvatarController(avatarService)
	avatarPublicRG := r.Group("/avatar")
	{
		avatarPublicRG.GET("", avatarController.GetAvatars)
		avatarPublicRG.GET("/:avatar_id", avatarController.GetOneAvatar)
		// content_type: music, video
		// query-params: page, limit, avatar_id, sort_by, sort_order
		avatarPublicRG.GET("/contents/:content_type", avatarController.GetAvatarContents)
		avatarPublicRG.GET("/contents/:content_type/:content_id", avatarController.GetOneAvatarContent)
	}
	myAvatarRG := r.Group("/avatar/my")
	myAvatarRG.Use(middleware.JWTAuthMiddleware())
	{
		myAvatarRG.GET("", avatarController.GetMyAvatars)
		myAvatarRG.GET("/contents/:content_type", avatarController.GetMyAvatarContents)
	}

	// ** Avatar Content Creation API **
	avatarContentCreationService := services.NewAvatarContentCreationService(
		DB,
		s3Service,
		systemPromptService,
		openAIPainter,
		openArtPainter,
		jenAIProducer,
		runwayVideoProducer,
	)
	avatarContentCreationController := controllers.NewAvatarContentCreationController(avatarContentCreationService)
	avatarCreationRG := r.Group("/avatar/:avatar_id/contents/create")
	avatarCreationRG.Use(middleware.JWTAuthMiddleware())
	{
		// music : prompt -> create by one step
		avatarCreationRG.POST("/music", avatarContentCreationController.StartMusicCreation)
		avatarCreationRG.GET("/music", avatarContentCreationController.GetMusicCreations)
		avatarCreationRG.GET("/music/:creation_id", avatarContentCreationController.GetOneMusicCreation)
		avatarCreationRG.POST("/music/:creation_id/confirm", avatarContentCreationController.ConfirmAvatarMusic) // confirm with NFT

		// video : prompt -> create by two step (1. image, 2. video)
		avatarCreationRG.POST("/video/image", avatarContentCreationController.StartVideoImageCreation)
		avatarCreationRG.GET("/video", avatarContentCreationController.GetVideoCreation)
		avatarCreationRG.GET("/video/:creation_id", avatarContentCreationController.GetOneVideoCreation)
		avatarCreationRG.POST("/video/image/:creation_id/create", avatarContentCreationController.StartVideoCreationFromImage)
		avatarCreationRG.POST("/video/image/:creation_id/confirm", avatarContentCreationController.ConfirmAvatarVideo) // confirm with NFT
	}

	// ** Avatar Remix API **
	avatarRemixService := services.NewAvatarRemixService(DB, s3Service, openArtPainter)
	avatarRemixController := controllers.NewAvatarRemixController(avatarRemixService)
	avatarRemixRG := r.Group("/avatar/:avatar_id/remix")
	avatarRemixRG.Use(middleware.JWTAuthMiddleware())
	{
		avatarRemixRG.POST("/image", avatarRemixController.StartImageRemix)
		avatarRemixRG.GET("/image/:remix_id", avatarRemixController.GetOneImageRemix)
		avatarRemixRG.POST("/image/:remix_id/confirm", avatarRemixController.ConfirmImageRemix)
	}

	r.Run(":8080")
}
