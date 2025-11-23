package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go-drive/internal/database"
)

func main() {
	// Parse command line flags
	action := flag.String("action", "migrate", "Action to perform: migrate, drop, roles")
	host := flag.String("host", getEnv("DB_HOST", "localhost"), "Database host")
	port := flag.String("port", getEnv("DB_PORT", "5432"), "Database port")
	user := flag.String("user", getEnv("DB_USER", "postgres"), "Database user")
	password := flag.String("password", getEnv("DB_PASSWORD", "postgres"), "Database password")
	dbname := flag.String("dbname", getEnv("DB_NAME", "postgres"), "Database name")
	sslmode := flag.String("sslmode", getEnv("DB_SSLMODE", "disable"), "SSL mode")

	flag.Parse()

	// Create database config
	cfg := database.Config{
		Host:     *host,
		Port:     *port,
		User:     *user,
		Password: *password,
		DBName:   *dbname,
		SSLMode:  *sslmode,
	}

	// Connect to database
	conn, err := database.NewGormConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	log.Printf("Connected to database: %s@%s:%s/%s", *user, *host, *port, *dbname)

	// Perform action
	switch *action {
	case "migrate":
		if err := database.AutoMigrate(conn.DB); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("✓ Migration completed successfully")

	case "drop":
		log.Println("⚠️  WARNING: This will drop all tables!")
		log.Println("Press Ctrl+C to cancel or Enter to continue...")
		fmt.Scanln()

		if err := database.DropAllTables(conn.DB); err != nil {
			log.Fatalf("Drop tables failed: %v", err)
		}
		log.Println("✓ All tables dropped successfully")

	case "roles":
		if err := database.CreateServiceRoles(conn.DB); err != nil {
			log.Fatalf("Create roles failed: %v", err)
		}
		log.Println("✓ Service roles created successfully")

	case "fresh":
		log.Println("⚠️  WARNING: This will drop all tables and recreate them!")
		log.Println("Press Ctrl+C to cancel or Enter to continue...")
		fmt.Scanln()

		if err := database.DropAllTables(conn.DB); err != nil {
			log.Fatalf("Drop tables failed: %v", err)
		}
		log.Println("✓ Tables dropped")

		if err := database.AutoMigrate(conn.DB); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("✓ Migration completed")

		if err := database.CreateServiceRoles(conn.DB); err != nil {
			log.Fatalf("Create roles failed: %v", err)
		}
		log.Println("✓ Service roles created")

		log.Println("✓ Fresh migration completed successfully")

	default:
		log.Fatalf("Unknown action: %s. Valid actions: migrate, drop, roles, fresh", *action)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
