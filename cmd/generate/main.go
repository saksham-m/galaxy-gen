package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"galaxy/galaxy"
)

func main() {
	fmt.Println("Welcome to the Artemis Galaxy Generator!")
	fmt.Println("-----------------------------------------")

	var numStars int
	fmt.Print("How many stars? (try 5000): ")
	fmt.Scan(&numStars)

	var shape string
	fmt.Print("Galaxy shape? (spiral / elliptical / ring / irregular): ")
	fmt.Scan(&shape)

	numArms := 4
	if shape == "spiral" {
		fmt.Print("How many spiral arms? (try 4): ")
		fmt.Scan(&numArms)
	}

	var colorName string
	fmt.Print("Color theme? (blue / red / gold / white): ")
	fmt.Scan(&colorName)

	theme, ok := galaxy.Themes[colorName]
	if !ok {
		fmt.Println("Unknown color, defaulting to blue.")
		theme = galaxy.Themes["blue"]
	}

	fmt.Println("\nGenerating galaxy...")

	img := image.NewRGBA(image.Rect(0, 0, galaxy.Width, galaxy.Height))
	for y := 0; y < galaxy.Height; y++ {
		for x := 0; x < galaxy.Width; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}

	cx, cy := float64(galaxy.Width)/2, float64(galaxy.Height)/2
	maxRadius := float64(galaxy.Width) / 2.5

	galaxy.DrawNebula(img, cx, cy, maxRadius, theme)

	stars := galaxy.GenerateStars(numStars, numArms, shape, theme, cx, cy, maxRadius)
	for _, s := range stars {
		galaxy.DrawStar(img, s.X, s.Y, s.C, s.Size)
	}

	galaxy.DrawCore(img, cx, cy, theme.Core, maxRadius*0.08)

	out, err := os.Create("galaxy.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	png.Encode(out, img)

	fmt.Println("Saved to galaxy.png")
}
