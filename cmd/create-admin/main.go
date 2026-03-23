package main

import (
	"ai-symptom-checker/config"
	"ai-symptom-checker/models"
	"ai-symptom-checker/pkg/utils"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Define flags
	name := flag.String("name", "", "Full name of the admin")
	email := flag.String("email", "", "Email address of the admin")
	username := flag.String("username", "", "Username for the admin login")
	password := flag.String("password", "", "Password for the admin")
	flag.Parse()

	// 2. Validate input
	if *name == "" || *email == "" || *username == "" || *password == "" {
		fmt.Println("Usage: go run cmd/create-admin/main.go -name \"Admin Name\" -email \"admin@example.com\" -username \"admin\" -password \"secret123\"")
		os.Exit(1)
	}

	// 3. Load Environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	config.Load()
	config.ConnectDB()
	db := config.DB

	// 4. Check if admin already exists
	var existing models.Admin
	if err := db.Where("email = ?", *email).First(&existing).Error; err == nil {
		log.Fatalf("Error: Admin with email %s already exists", *email)
	}

	// 5. Hash password
	hashedPassword, err := utils.HashPassword(*password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// 6. Create Admin
	admin := models.Admin{
		Name:     *name,
		Email:    *email,
		Username: *username,
		Password: hashedPassword,
		IsActive: true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	fmt.Printf("\nSUCCESS: Admin account created successfully!\n")
	fmt.Printf("Name:  %s\n", admin.Name)
	fmt.Printf("Email: %s\n", admin.Email)
	fmt.Println("----------------------------------------------")
}
