package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"go-drive/internal/database"
	pb "go-drive/proto/user"
	"go-drive/services/user-service/repository"
	"go-drive/services/user-service/service"
)

func main() {
	// Get configuration from environment
	port := getEnv("GRPC_PORT", "50051")
	dbHost := getEnv("DB_HOST", "postgres")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "postgres")
	dbUser := getEnv("DB_USER", "user_service")
	dbPassword := getEnv("DB_PASSWORD", "user_service_password")
	sslMode := getEnv("DB_SSLMODE", "disable")

	// Initialize database connection using shared package
	dbConfig := database.Config{
		Host:            dbHost,
		Port:            dbPort,
		User:            dbUser,
		Password:        dbPassword,
		DBName:          dbName,
		SSLMode:         sslMode,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	log.Printf("Connecting to database: %s@%s:%s/%s", dbUser, dbHost, dbPort, dbName)

	repo, err := repository.NewPostgresUserRepository(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer repo.Close()

	log.Println("Database connection established successfully")

	// Create gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register user service
	userService := service.NewUserService(repo)
	pb.RegisterUserServiceServer(grpcServer, userService)

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("user.UserService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service for debugging
	reflection.Register(grpcServer)

	log.Printf("User service starting on port %s", port)

	// Start server in goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down user service...")
	grpcServer.GracefulStop()
	log.Println("User service stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
