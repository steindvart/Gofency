package captcha

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	service := NewService("../../assets")

	captcha, err := service.Generate()
	if err != nil {
		t.Fatalf("Failed to generate captcha: %v", err)
	}

	if captcha == nil {
		t.Fatal("Generated captcha is nil")
	}

	if len(captcha.Answer) != digitCount {
		t.Errorf("Expected answer length %d, got %d", digitCount, len(captcha.Answer))
	}

	if len(captcha.Image) == 0 {
		t.Error("Generated image is empty")
	}

	// Verify answer contains only digits
	for _, r := range captcha.Answer {
		if r < '0' || r > '9' {
			t.Errorf("Answer contains non-digit character: %c", r)
		}
	}
}

func TestLoadFromAssets(t *testing.T) {
	service := NewService("../../assets")

	captcha, err := service.LoadFromAssets()
	if err != nil {
		t.Fatalf("Failed to load captcha from assets: %v", err)
	}

	if captcha == nil {
		t.Fatal("Loaded captcha is nil")
	}

	if len(captcha.Answer) == 0 {
		t.Error("Captcha answer is empty")
	}

	if len(captcha.Image) == 0 {
		t.Error("Captcha image is empty")
	}
}

func TestServiceWithInvalidPath(t *testing.T) {
	service := NewService("/nonexistent/path")

	// Should still work by generating captcha dynamically
	captcha, err := service.LoadFromAssets()
	if err != nil {
		t.Fatalf("Failed to generate captcha dynamically: %v", err)
	}

	if captcha == nil {
		t.Fatal("Generated captcha is nil")
	}

	if len(captcha.Answer) != digitCount {
		t.Errorf("Expected answer length %d, got %d", digitCount, len(captcha.Answer))
	}
}
