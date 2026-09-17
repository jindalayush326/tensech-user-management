package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"tensechassignment/internal/auth"
	"tensechassignment/internal/cache"
	"tensechassignment/internal/config"
	"tensechassignment/internal/controller"
	"tensechassignment/internal/db"
	"tensechassignment/internal/middleware"
	"tensechassignment/internal/repository"
	"tensechassignment/internal/routes"
	"tensechassignment/internal/service"
)

func main() {
	cfg := config.Load()

	mongoClient, err := db.ConnectMongo(cfg.MongoURI)
	if err != nil {
		log.Fatal("mongodb connection failed:", err)
	}
	defer mongoClient.Disconnect(context.Background())

	mongoDB := mongoClient.Database(cfg.MongoDBName)
	userRepo := repository.NewMongoUserRepository(mongoDB)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	err = userRepo.EnsureIndexes(ctx)
	cancel()

	if err != nil {
		log.Fatal("failed to create indexes:", err)
	}

	redisClient := db.ConnectRedis(
		cfg.RedisAddr,
		cfg.RedisPass,
	)
	defer redisClient.Close()

	validator := auth.NewValidator(
		cfg.JWTSecret,
		cfg.JWTIssuer,
		cfg.JWTAudience,
	)

	userService := service.NewUserService(
		userRepo,
		cache.NewRedisCache(redisClient),
		cfg.CacheTTL,
	)

	csvService := service.NewCSVService(userRepo)

	userController := controller.NewUserController(userService)
	csvController := controller.NewCSVController(csvService)

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())

	routes.Setup(
		router,
		validator,
		userController,
		csvController,
	)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Println("server started on port", cfg.Port)

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("server error:", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Println("shutting down server...")

	ctx, cancel = context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("shutdown error:", err)
	}

	log.Println("server stopped")
}