package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gofency/internal/captcha"
)

func main() {
	// Create captcha service
	service := captcha.NewService("assets")

	// Generate sample captchas with specific answers
	samples := []string{"1234", "5678", "9012", "3456", "7890", "2468", "1357", "9753"}

	captchaDir := "assets/captcha"
	if err := os.MkdirAll(captchaDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	fmt.Println("Generating sample captcha images...")

	for i, answer := range samples {
		img, err := service.Generate()
		if err != nil {
			log.Printf("Failed to generate captcha %d: %v", i, err)
			continue
		}

		// Save with the answer as filename
		filename := filepath.Join(captchaDir, answer+".png")
		if err := os.WriteFile(filename, img.Image, 0644); err != nil {
			log.Printf("Failed to write captcha file %s: %v", filename, err)
			continue
		}

		fmt.Printf("Generated %s\n", filename)
	}

	fmt.Println("Done! Generated", len(samples), "captcha images.")
}
