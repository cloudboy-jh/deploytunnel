package tui

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BourgeoisBear/rasterm"
	"github.com/qeesung/image2ascii/convert"
	"golang.org/x/term"
)

var (
	asciiArtCache     string
	asciiArtCacheLock sync.Mutex
	imageSupported    *bool
)

// DisplayImage tries to display the deploytunnel.png image using terminal protocols
// Falls back to ASCII art if protocols aren't supported
func DisplayImage() string {
	// Get terminal width for scaling
	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || termWidth == 0 {
		termWidth = 80 // Default fallback
	}

	// Try to load the image
	imgPath := findImagePath()
	if imgPath == "" {
		return "" // No image found, skip
	}

	// Check if terminal supports image protocols
	if supportsImageProtocol() {
		if imgStr := tryTerminalImage(imgPath, termWidth); imgStr != "" {
			return imgStr
		}
	}

	// Fall back to ASCII art
	return getASCIIArt(imgPath, termWidth)
}

// findImagePath locates the deploytunnel.png file
func findImagePath() string {
	// Try multiple locations
	var locations []string

	// Try current working directory first
	if cwd, err := os.Getwd(); err == nil {
		locations = append(locations,
			filepath.Join(cwd, "deploytunnel.png"),
			filepath.Join(cwd, "..", "deploytunnel.png"),
		)
	}

	// Try relative paths
	locations = append(locations,
		"deploytunnel.png",
		"./deploytunnel.png",
		"../deploytunnel.png",
	)

	// Try relative to executable (works for installed binaries)
	if execPath, err := os.Executable(); err == nil {
		// Resolve symlinks (important for go run and installed binaries)
		if realExec, err := filepath.EvalSymlinks(execPath); err == nil {
			execPath = realExec
		}
		execDir := filepath.Dir(execPath)
		locations = append(locations,
			filepath.Join(execDir, "deploytunnel.png"),
			filepath.Join(execDir, "..", "deploytunnel.png"),
			filepath.Join(execDir, "..", "..", "deploytunnel.png"),
		)
	}

	// Check all locations
	for _, path := range locations {
		// Convert to absolute path for better debugging
		if absPath, err := filepath.Abs(path); err == nil {
			path = absPath
		}

		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}

// supportsImageProtocol checks if the terminal supports image display
func supportsImageProtocol() bool {
	if imageSupported != nil {
		return *imageSupported
	}

	// Check environment variables for known terminals
	termProgram := os.Getenv("TERM_PROGRAM")
	kittyWindow := os.Getenv("KITTY_WINDOW_ID")
	term := os.Getenv("TERM")

	supported := false

	switch {
	case termProgram == "iTerm.app":
		supported = true
	case kittyWindow != "":
		supported = true
	case strings.Contains(term, "kitty"):
		supported = true
	case strings.Contains(term, "mlterm"):
		supported = true
	case strings.Contains(term, "yaft"):
		supported = true
	}

	imageSupported = &supported
	return supported
}

// tryTerminalImage attempts to display the image using terminal protocols
func tryTerminalImage(imgPath string, termWidth int) string {
	file, err := os.Open(imgPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return ""
	}

	// Try to encode using rasterm protocols
	var output strings.Builder

	// Check which protocol to use based on terminal
	termProgram := os.Getenv("TERM_PROGRAM")
	kittyWindow := os.Getenv("KITTY_WINDOW_ID")

	// Try Kitty protocol first (most capable)
	if kittyWindow != "" || rasterm.IsKittyCapable() {
		// Use DstCols for destination width in terminal columns
		targetCols := uint32(float64(termWidth) * 0.75)
		opts := rasterm.KittyImgOpts{
			DstCols: targetCols,
			DstRows: 0, // Auto height
		}
		if err := rasterm.KittyWriteImage(&output, img, opts); err == nil {
			return output.String() + "\n"
		}
	}

	// Try iTerm2 protocol
	if termProgram == "iTerm.app" || rasterm.IsItermCapable() {
		if err := rasterm.ItermWriteImage(&output, img); err == nil {
			return output.String() + "\n"
		}
	}

	// Try Sixel protocol as last resort
	if capable, err := rasterm.IsSixelCapable(); err == nil && capable {
		// Convert to paletted image for Sixel
		bounds := img.Bounds()
		palettedImg := image.NewPaletted(bounds, nil)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				palettedImg.Set(x, y, img.At(x, y))
			}
		}
		if err := rasterm.SixelWriteImage(&output, palettedImg); err == nil {
			return output.String() + "\n"
		}
	}

	return ""
}

// getASCIIArt generates or retrieves cached ASCII art
func getASCIIArt(imgPath string, termWidth int) string {
	asciiArtCacheLock.Lock()
	defer asciiArtCacheLock.Unlock()

	// Return cached version if available
	if asciiArtCache != "" {
		return asciiArtCache
	}

	// Generate ASCII art
	convertOptions := convert.DefaultOptions
	convertOptions.FixedWidth = int(float64(termWidth) * 0.75)
	if convertOptions.FixedWidth > 80 {
		convertOptions.FixedWidth = 80
	}
	if convertOptions.FixedWidth < 40 {
		convertOptions.FixedWidth = 40
	}
	convertOptions.FixedHeight = -1 // Auto height
	convertOptions.Colored = false  // No color for cleaner output

	converter := convert.NewImageConverter()
	asciiArt := converter.ImageFile2ASCIIString(imgPath, &convertOptions)

	// Center the ASCII art
	lines := strings.Split(asciiArt, "\n")
	var centered strings.Builder

	for _, line := range lines {
		if line != "" {
			padding := (termWidth - len(line)) / 2
			if padding > 0 {
				centered.WriteString(strings.Repeat(" ", padding))
			}
			centered.WriteString(line)
			centered.WriteString("\n")
		}
	}

	asciiArtCache = centered.String()
	return asciiArtCache
}

// ClearImageCache clears the ASCII art cache (useful for testing or terminal resize)
func ClearImageCache() {
	asciiArtCacheLock.Lock()
	defer asciiArtCacheLock.Unlock()
	asciiArtCache = ""
	imageSupported = nil
}
