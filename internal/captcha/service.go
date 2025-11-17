package captcha

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/big"
	"os"
	"path/filepath"
)

const (
	captchaWidth    = 200
	captchaHeight   = 80
	digitCount      = 4
	noiseLinesCount = 10
	noiseDotsCount  = 100
)

// CaptchaImage represents a captcha image with its answer
type CaptchaImage struct {
	Image  []byte
	Answer string
}

// Service handles captcha generation
type Service struct {
	assetsDir string
}

// NewService creates a new captcha service
func NewService(assetsDir string) *Service {
	return &Service{assetsDir: assetsDir}
}

// Generate creates a new captcha image
func (s *Service) Generate() (*CaptchaImage, error) {
	// Generate random digits
	answer := ""
	for i := 0; i < digitCount; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return nil, fmt.Errorf("failed to generate random digit: %w", err)
		}
		answer += n.String()
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, captchaWidth, captchaHeight))

	// Fill background with light gray
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{240, 240, 240, 255}}, image.Point{}, draw.Src)

	// Add some noise lines
	for i := 0; i < noiseLinesCount; i++ {
		s.drawNoiseLine(img)
	}

	// Draw digits
	digitWidth := captchaWidth / digitCount
	for i, digit := range answer {
		x := i*digitWidth + digitWidth/4
		y := captchaHeight / 2
		s.drawDigit(img, digit, x, y)
	}

	// Add more noise dots
	for i := 0; i < noiseDotsCount; i++ {
		s.drawNoiseDot(img)
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return &CaptchaImage{
		Image:  buf.Bytes(),
		Answer: answer,
	}, nil
}

// LoadFromAssets loads a random captcha from the assets directory
func (s *Service) LoadFromAssets() (*CaptchaImage, error) {
	// Check if assets directory exists
	captchaDir := filepath.Join(s.assetsDir, "captcha")
	if _, err := os.Stat(captchaDir); os.IsNotExist(err) {
		// If no assets, generate a new captcha
		return s.Generate()
	}

	// List all PNG files
	files, err := filepath.Glob(filepath.Join(captchaDir, "*.png"))
	if err != nil || len(files) == 0 {
		// If no files found, generate a new captcha
		return s.Generate()
	}

	// Pick a random file
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(files))))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random index: %w", err)
	}

	selectedFile := files[n.Int64()]

	// Extract answer from filename (e.g., "5647.png" -> "5647")
	filename := filepath.Base(selectedFile)
	answer := filename[:len(filename)-4] // Remove .png extension

	// Read image file
	data, err := os.ReadFile(selectedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read captcha file: %w", err)
	}

	return &CaptchaImage{
		Image:  data,
		Answer: answer,
	}, nil
}

// drawDigit draws a single digit on the image
func (s *Service) drawDigit(img *image.RGBA, digit rune, x, y int) {
	// Simple 7-segment style digit drawing
	segments := s.getDigitSegments(digit)
	segmentColor := color.RGBA{50, 50, 50, 255}

	for _, seg := range segments {
		s.drawSegment(img, seg, x, y, segmentColor)
	}
}

type segment struct {
	x1, y1, x2, y2 int
}

func (s *Service) getDigitSegments(digit rune) []segment {
	// Simple representation of digits using segments
	// This is a simplified version - in production you'd use a font
	baseSegments := map[rune][]segment{
		'0': {{-10, -20, 10, -20}, {10, -20, 10, 20}, {10, 20, -10, 20}, {-10, 20, -10, -20}},
		'1': {{0, -20, 0, 20}},
		'2': {{-10, -20, 10, -20}, {10, -20, 10, 0}, {10, 0, -10, 0}, {-10, 0, -10, 20}, {-10, 20, 10, 20}},
		'3': {{-10, -20, 10, -20}, {10, -20, 10, 20}, {10, 20, -10, 20}, {-5, 0, 10, 0}},
		'4': {{-10, -20, -10, 0}, {-10, 0, 10, 0}, {10, -20, 10, 20}},
		'5': {{10, -20, -10, -20}, {-10, -20, -10, 0}, {-10, 0, 10, 0}, {10, 0, 10, 20}, {10, 20, -10, 20}},
		'6': {{10, -20, -10, -20}, {-10, -20, -10, 20}, {-10, 20, 10, 20}, {10, 20, 10, 0}, {10, 0, -10, 0}},
		'7': {{-10, -20, 10, -20}, {10, -20, 10, 20}},
		'8': {{-10, -20, 10, -20}, {10, -20, 10, 20}, {10, 20, -10, 20}, {-10, 20, -10, -20}, {-10, 0, 10, 0}},
		'9': {{10, 20, 10, -20}, {10, -20, -10, -20}, {-10, -20, -10, 0}, {-10, 0, 10, 0}},
	}
	return baseSegments[digit]
}

func (s *Service) drawSegment(img *image.RGBA, seg segment, offsetX, offsetY int, col color.RGBA) {
	x1 := offsetX + seg.x1
	y1 := offsetY + seg.y1
	x2 := offsetX + seg.x2
	y2 := offsetY + seg.y2

	s.drawLine(img, x1, y1, x2, y2, col)
}

func (s *Service) drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.RGBA) {
	dx := math.Abs(float64(x2 - x1))
	dy := math.Abs(float64(y2 - y1))
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx - dy

	for {
		if x1 >= 0 && x1 < captchaWidth && y1 >= 0 && y1 < captchaHeight {
			img.Set(x1, y1, col)
			// Make line thicker
			if x1+1 < captchaWidth {
				img.Set(x1+1, y1, col)
			}
			if y1+1 < captchaHeight {
				img.Set(x1, y1+1, col)
			}
		}

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (s *Service) drawNoiseLine(img *image.RGBA) {
	x1, _ := rand.Int(rand.Reader, big.NewInt(captchaWidth))
	y1, _ := rand.Int(rand.Reader, big.NewInt(captchaHeight))
	x2, _ := rand.Int(rand.Reader, big.NewInt(captchaWidth))
	y2, _ := rand.Int(rand.Reader, big.NewInt(captchaHeight))

	col := color.RGBA{200, 200, 200, 255}
	s.drawLine(img, int(x1.Int64()), int(y1.Int64()), int(x2.Int64()), int(y2.Int64()), col)
}

func (s *Service) drawNoiseDot(img *image.RGBA) {
	x, _ := rand.Int(rand.Reader, big.NewInt(captchaWidth))
	y, _ := rand.Int(rand.Reader, big.NewInt(captchaHeight))

	col := color.RGBA{180, 180, 180, 255}
	img.Set(int(x.Int64()), int(y.Int64()), col)
}
