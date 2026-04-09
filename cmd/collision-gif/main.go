package main

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"math"
	"os"

	"galaxy/galaxy"
)

const (
	gifW    = 900
	gifH    = 600
	nFrames = 40
	delay   = 4 // centiseconds per frame — 4 = 25fps
)

func main() {
	fmt.Println("Artemis Collision GIF Generator")
	fmt.Println("--------------------------------")

	fmt.Println("Galaxy 1 (left):")
	g1stars, g1theme := promptGalaxy()

	fmt.Println("Galaxy 2 (right):")
	g2stars, g2theme := promptGalaxy()

	fmt.Printf("\nRendering %d frames...\n", nFrames)

	pal := buildPalette(g1theme, g2theme)

	frames := make([]*image.Paletted, nFrames)
	delays := make([]int, nFrames)

	for f := 0; f < nFrames; f++ {
		t := float64(f) / float64(nFrames-1) // 0.0 → 1.0

		// Ease in with a smooth curve so the approach accelerates
		tEased := t * t * (3 - 2*t)

		// Galaxies arc toward each other (curved orbital approach)
		// G1 starts upper-left, G2 starts lower-right; both converge to center
		arc := math.Sin(tEased * math.Pi / 2)
		separationX := 280 - tEased*200 // 280px apart → 80px apart
		separationY := arc * 60         // slight vertical arc

		cx := float64(gifW) / 2
		cy := float64(gifH) / 2

		g1cx := cx - separationX/2
		g1cy := cy - separationY/2
		g2cx := cx + separationX/2
		g2cy := cy + separationY/2

		maxRadius := float64(gifH) / 2.8
		intensity := tEased * 0.95

		img := renderFrame(
			g1stars, g1theme, g1cx, g1cy,
			g2stars, g2theme, g2cx, g2cy,
			maxRadius, intensity,
		)

		paletted := image.NewPaletted(img.Bounds(), pal)
		draw.FloydSteinberg.Draw(paletted, img.Bounds(), img, image.Point{})

		frames[f] = paletted
		delays[f] = delay
		fmt.Printf("  frame %d/%d\n", f+1, nFrames)
	}

	// Hold the final frame a bit longer
	delays[nFrames-1] = 80

	out, err := os.Create("collision.gif")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	gif.EncodeAll(out, &gif.GIF{
		Image:     frames,
		Delay:     delays,
		LoopCount: 0, // loop forever
	})

	fmt.Println("Saved to collision.gif")
}

func renderFrame(
	g1stars []galaxy.Star, g1theme galaxy.ColorTheme, g1cx, g1cy float64,
	g2stars []galaxy.Star, g2theme galaxy.ColorTheme, g2cx, g2cy float64,
	maxRadius, intensity float64,
) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, gifW, gifH))
	for y := 0; y < gifH; y++ {
		for x := 0; x < gifW; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}

	galaxy.DrawNebulaCustom(img, gifW, gifH, g1cx, g1cy, maxRadius, g1theme)
	galaxy.DrawNebulaCustom(img, gifW, gifH, g2cx, g2cy, maxRadius, g2theme)

	// Clone star slices so we don't mutate originals between frames
	s1 := cloneStars(g1stars)
	s2 := cloneStars(g2stars)

	// Re-center stars around their galaxy's current frame position
	recenter(s1, g1cx, g1cy)
	recenter(s2, g2cx, g2cy)

	applyTidalForce(s1, g2cx, g2cy, intensity)
	applyTidalForce(s2, g1cx, g1cy, intensity)

	for _, s := range s1 {
		galaxy.DrawStarCustom(img, gifW, gifH, s.X, s.Y, s.C, s.Size)
	}
	for _, s := range s2 {
		galaxy.DrawStarCustom(img, gifW, gifH, s.X, s.Y, s.C, s.Size)
	}

	galaxy.DrawCoreCustom(img, gifW, gifH, g1cx, g1cy, g1theme.Core, maxRadius*0.08)
	galaxy.DrawCoreCustom(img, gifW, gifH, g2cx, g2cy, g2theme.Core, maxRadius*0.08)

	return img
}

// recenter shifts a star slice so that its origin (0,0 reference) maps to (cx, cy).
// Stars are generated with an arbitrary center; we normalize to origin first.
func recenter(stars []galaxy.Star, cx, cy float64) {
	if len(stars) == 0 {
		return
	}
	// Stars were generated at origin (0,0) — shift to target center
	for i := range stars {
		stars[i].X += cx
		stars[i].Y += cy
	}
}

func cloneStars(src []galaxy.Star) []galaxy.Star {
	dst := make([]galaxy.Star, len(src))
	copy(dst, src)
	return dst
}

func applyTidalForce(stars []galaxy.Star, otherCX, otherCY, intensity float64) {
	const G = 8_000_000.0
	for i := range stars {
		dx := otherCX - stars[i].X
		dy := otherCY - stars[i].Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < 1 {
			continue
		}
		force := G * intensity / (dist * dist)
		nx, ny := dx/dist, dy/dist
		stars[i].X += nx * force
		stars[i].Y += ny * force
		tx, ty := -ny, nx
		curl := force * 0.4 * math.Log1p(dist/200)
		stars[i].X += tx * curl
		stars[i].Y += ty * curl
	}
}

// buildPalette creates a 256-color palette covering both galaxy themes.
func buildPalette(t1, t2 galaxy.ColorTheme) color.Palette {
	p := make(color.Palette, 0, 256)
	p = append(p, color.RGBA{0, 0, 0, 255}) // black

	addGradient := func(a, b color.RGBA, steps int) {
		for i := 0; i < steps; i++ {
			frac := float64(i) / float64(steps-1)
			p = append(p, color.RGBA{
				R: uint8(float64(a.R)*(1-frac) + float64(b.R)*frac),
				G: uint8(float64(a.G)*(1-frac) + float64(b.G)*frac),
				B: uint8(float64(a.B)*(1-frac) + float64(b.B)*frac),
				A: 255,
			})
		}
	}

	addGradient(color.RGBA{0, 0, 0, 255}, t1.Edge, 20)
	addGradient(t1.Edge, t1.Mid, 25)
	addGradient(t1.Mid, t1.Core, 25)

	addGradient(color.RGBA{0, 0, 0, 255}, t2.Edge, 20)
	addGradient(t2.Edge, t2.Mid, 25)
	addGradient(t2.Mid, t2.Core, 25)

	// Bright whites for star cores and overlap blooms
	addGradient(t1.Core, color.RGBA{255, 255, 255, 255}, 15)
	addGradient(t2.Core, color.RGBA{255, 255, 255, 255}, 15)

	// Pad with Plan 9 palette entries to hit 256
	for len(p) < 256 {
		p = append(p, palette.Plan9[len(p)%len(palette.Plan9)])
	}

	return p[:256]
}

type galaxyInput struct {
	numStars int
	numArms  int
	shape    string
	theme    galaxy.ColorTheme
}

func promptGalaxy() ([]galaxy.Star, galaxy.ColorTheme) {
	var in galaxyInput

	fmt.Print("  Stars? (try 3000): ")
	fmt.Scan(&in.numStars)

	fmt.Print("  Shape? (spiral / elliptical / ring / irregular): ")
	fmt.Scan(&in.shape)

	in.numArms = 4
	if in.shape == "spiral" {
		fmt.Print("  Arms? (try 3): ")
		fmt.Scan(&in.numArms)
	}

	var colorName string
	fmt.Print("  Color? (blue / red / gold / white): ")
	fmt.Scan(&colorName)

	var ok bool
	in.theme, ok = galaxy.Themes[colorName]
	if !ok {
		in.theme = galaxy.Themes["blue"]
	}
	fmt.Println()

	// Generate stars at origin — recenter() will place them each frame
	stars := galaxy.GenerateStars(in.numStars, in.numArms, in.shape, in.theme, 0, 0, float64(gifH)/2.8)
	return stars, in.theme
}
