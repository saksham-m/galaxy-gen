package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"galaxy/galaxy"
)

// collisionConfig holds the setup for one galaxy in the collision scene.
type collisionConfig struct {
	cx, cy    float64
	maxRadius float64
	numStars  int
	numArms   int
	shape     string
	theme     galaxy.ColorTheme
}

func main() {
	fmt.Println("Artemis Galaxy Collision Generator")
	fmt.Println("-----------------------------------")
	fmt.Println("Two galaxies, two color themes. Stars are pulled by tidal gravity.")
	fmt.Println()

	// Galaxy 1
	fmt.Println("=== Galaxy 1 (left) ===")
	g1 := promptGalaxy()

	// Galaxy 2
	fmt.Println("=== Galaxy 2 (right) ===")
	g2 := promptGalaxy()

	// Collision intensity: how strongly each galaxy's stars are pulled toward the other
	var intensity float64
	fmt.Print("\nCollision intensity? (0.0 = barely grazing, 1.0 = deep merge): ")
	fmt.Scan(&intensity)
	intensity = math.Max(0, math.Min(1, intensity))

	fmt.Println("\nGenerating collision...")

	// Position galaxies on a wider canvas
	const canvasW, canvasH = 1800, 1200
	separation := 320 + (1-intensity)*160 // closer = deeper collision

	g1.cx = float64(canvasW)/2 - separation/2
	g1.cy = float64(canvasH) / 2
	g1.maxRadius = float64(canvasH) / 2.8

	g2.cx = float64(canvasW)/2 + separation/2
	g2.cy = float64(canvasH) / 2
	g2.maxRadius = float64(canvasH) / 2.8

	img := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))
	for y := 0; y < canvasH; y++ {
		for x := 0; x < canvasW; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}

	// Draw nebula for both galaxies
	galaxy.DrawNebulaCustom(img, canvasW, canvasH, g1.cx, g1.cy, g1.maxRadius, g1.theme)
	galaxy.DrawNebulaCustom(img, canvasW, canvasH, g2.cx, g2.cy, g2.maxRadius, g2.theme)

	// Generate stars, apply tidal distortion, render
	stars1 := galaxy.GenerateStars(g1.numStars, g1.numArms, g1.shape, g1.theme, g1.cx, g1.cy, g1.maxRadius)
	stars2 := galaxy.GenerateStars(g2.numStars, g2.numArms, g2.shape, g2.theme, g2.cx, g2.cy, g2.maxRadius)

	applyTidalForce(stars1, g2.cx, g2.cy, intensity)
	applyTidalForce(stars2, g1.cx, g1.cy, intensity)

	for _, s := range stars1 {
		galaxy.DrawStarCustom(img, canvasW, canvasH, s.X, s.Y, s.C, s.Size)
	}
	for _, s := range stars2 {
		galaxy.DrawStarCustom(img, canvasW, canvasH, s.X, s.Y, s.C, s.Size)
	}

	galaxy.DrawCoreCustom(img, canvasW, canvasH, g1.cx, g1.cy, g1.theme.Core, g1.maxRadius*0.08)
	galaxy.DrawCoreCustom(img, canvasW, canvasH, g2.cx, g2.cy, g2.theme.Core, g2.maxRadius*0.08)

	out, err := os.Create("collision.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	png.Encode(out, img)

	fmt.Println("Saved to collision.png")
}

// applyTidalForce displaces each star toward the other galaxy's center.
// Stars near the other center get pulled harder (inverse square).
// A tangential component curls the tidal tails outward.
func applyTidalForce(stars []galaxy.Star, otherCX, otherCY, intensity float64) {
	// Tuned so at intensity=1.0 a star at distance=300px moves ~80px
	const G = 8_000_000.0

	for i := range stars {
		dx := otherCX - stars[i].X
		dy := otherCY - stars[i].Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < 1 {
			continue
		}

		// Radial pull toward other center
		force := G * intensity / (dist * dist)
		nx, ny := dx/dist, dy/dist
		stars[i].X += nx * force
		stars[i].Y += ny * force

		// Tangential curl: creates the characteristic arcing tidal tails
		// Perpendicular to the pull direction, scaled by distance from other center
		tx, ty := -ny, nx
		curl := force * 0.4 * math.Log1p(dist/200)
		stars[i].X += tx * curl
		stars[i].Y += ty * curl
	}
}

func promptGalaxy() collisionConfig {
	var cfg collisionConfig

	fmt.Print("  Stars? (try 4000): ")
	fmt.Scan(&cfg.numStars)

	fmt.Print("  Shape? (spiral / elliptical / ring / irregular): ")
	fmt.Scan(&cfg.shape)

	cfg.numArms = 4
	if cfg.shape == "spiral" {
		fmt.Print("  Arms? (try 3): ")
		fmt.Scan(&cfg.numArms)
	}

	var colorName string
	fmt.Print("  Color? (blue / red / gold / white): ")
	fmt.Scan(&colorName)

	var ok bool
	cfg.theme, ok = galaxy.Themes[colorName]
	if !ok {
		cfg.theme = galaxy.Themes["blue"]
	}
	fmt.Println()
	return cfg
}
