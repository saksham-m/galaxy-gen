package galaxy

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

const (
	Width  = 1200
	Height = 1200
)

type ColorTheme struct {
	Core color.RGBA
	Mid  color.RGBA
	Edge color.RGBA
}

var Themes = map[string]ColorTheme{
	"blue": {
		Core: color.RGBA{200, 220, 255, 255},
		Mid:  color.RGBA{100, 140, 255, 255},
		Edge: color.RGBA{40, 60, 180, 255},
	},
	"red": {
		Core: color.RGBA{255, 220, 200, 255},
		Mid:  color.RGBA{255, 120, 80, 255},
		Edge: color.RGBA{180, 40, 20, 255},
	},
	"gold": {
		Core: color.RGBA{255, 255, 200, 255},
		Mid:  color.RGBA{255, 200, 80, 255},
		Edge: color.RGBA{180, 120, 20, 255},
	},
	"white": {
		Core: color.RGBA{255, 255, 255, 255},
		Mid:  color.RGBA{200, 200, 220, 255},
		Edge: color.RGBA{120, 120, 160, 255},
	},
}

type Star struct {
	X, Y float64
	C    color.RGBA
	Size int
}

// GenerateStars returns star positions for a galaxy centered at (cx, cy).
// No drawing happens here — callers can transform positions before rendering.
func GenerateStars(numStars, numArms int, shape string, theme ColorTheme, cx, cy, maxRadius float64) []Star {
	switch shape {
	case "elliptical":
		return placeElliptical(numStars, cx, cy, maxRadius, theme)
	case "ring":
		return placeRing(numStars, cx, cy, maxRadius, theme)
	case "irregular":
		return placeIrregular(numStars, cx, cy, maxRadius, theme)
	default:
		return placeSpiral(numStars, numArms, cx, cy, maxRadius, theme)
	}
}

func placeSpiral(numStars, numArms int, cx, cy, maxRadius float64, theme ColorTheme) []Star {
	b := 0.3
	starsPerArm := numStars / numArms
	stars := make([]Star, 0, numStars)
	for arm := 0; arm < numArms; arm++ {
		armOffset := float64(arm) * (2 * math.Pi / float64(numArms))
		for i := 0; i < starsPerArm; i++ {
			t := rand.Float64()
			theta := t * 4 * math.Pi
			r := maxRadius * (math.Exp(b*theta) - 1) / (math.Exp(b*4*math.Pi) - 1)
			r += r * 0.18 * rand.NormFloat64()
			theta += 0.15 * rand.NormFloat64()
			x := cx + r*math.Cos(theta+armOffset)
			y := cy + r*math.Sin(theta+armOffset)
			ratio := math.Min(math.Abs(r)/maxRadius, 1)
			c := LerpColor(theme.Core, theme.Mid, theme.Edge, ratio)
			c.A = uint8(255 - Clamp(int(ratio*160)))
			stars = append(stars, Star{x, y, c, StarSize()})
		}
	}
	return stars
}

func placeElliptical(numStars int, cx, cy, maxRadius float64, theme ColorTheme) []Star {
	sigmaX := maxRadius * 0.35
	sigmaY := maxRadius * 0.22
	stars := make([]Star, 0, numStars)
	for i := 0; i < numStars; i++ {
		x := cx + rand.NormFloat64()*sigmaX
		y := cy + rand.NormFloat64()*sigmaY
		dx, dy := x-cx, y-cy
		ratio := math.Min(math.Sqrt((dx/sigmaX)*(dx/sigmaX)+(dy/sigmaY)*(dy/sigmaY))/3.0, 1)
		c := LerpColor(theme.Core, theme.Mid, theme.Edge, ratio)
		c.A = uint8(255 - uint8(ratio*180))
		stars = append(stars, Star{x, y, c, StarSize()})
	}
	return stars
}

func placeRing(numStars int, cx, cy, maxRadius float64, theme ColorTheme) []Star {
	ringRadius := maxRadius * 0.6
	ringWidth := maxRadius * 0.12
	stars := make([]Star, 0, numStars)
	for i := 0; i < numStars; i++ {
		theta := rand.Float64() * 2 * math.Pi
		r := ringRadius + rand.NormFloat64()*ringWidth
		x := cx + r*math.Cos(theta)
		y := cy + r*math.Sin(theta)
		dist := math.Abs(r - ringRadius)
		ratio := math.Min(dist/(ringWidth*3), 1)
		c := LerpColor(theme.Core, theme.Mid, theme.Edge, ratio)
		c.A = uint8(200 - uint8(ratio*150))
		stars = append(stars, Star{x, y, c, StarSize()})
	}
	return stars
}

func placeIrregular(numStars int, cx, cy, maxRadius float64, theme ColorTheme) []Star {
	numClumps := 6 + rand.Intn(5)
	clumps := make([][2]float64, numClumps)
	for i := range clumps {
		angle := rand.Float64() * 2 * math.Pi
		dist := rand.Float64() * maxRadius * 0.5
		clumps[i] = [2]float64{cx + dist*math.Cos(angle), cy + dist*math.Sin(angle)}
	}
	stars := make([]Star, 0, numStars)
	starsPerClump := numStars / numClumps
	for _, clump := range clumps {
		sigma := maxRadius * (0.08 + rand.Float64()*0.12)
		for i := 0; i < starsPerClump; i++ {
			x := clump[0] + rand.NormFloat64()*sigma
			y := clump[1] + rand.NormFloat64()*sigma
			dx, dy := x-cx, y-cy
			ratio := math.Min(math.Sqrt(dx*dx+dy*dy)/maxRadius, 1)
			c := LerpColor(theme.Core, theme.Mid, theme.Edge, ratio)
			c.A = uint8(220 - uint8(ratio*160))
			stars = append(stars, Star{x, y, c, StarSize()})
		}
	}
	return stars
}

// DrawNebula paints a soft noise-based cloud layer before stars.
func DrawNebula(img *image.RGBA, cx, cy, maxRadius float64, theme ColorTheme) {
	DrawNebulaCustom(img, Width, Height, cx, cy, maxRadius, theme)
}

func DrawNebulaCustom(img *image.RGBA, w, h int, cx, cy, maxRadius float64, theme ColorTheme) {
	p := buildPerm()
	scale := 3.5 / float64(w)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := float64(x)-cx, float64(y)-cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > maxRadius*1.1 {
				continue
			}
			n := fbm(float64(x)*scale, float64(y)*scale, 6, p)
			edgeFade := math.Max(0, 1-dist/maxRadius)
			intensity := math.Max(0, n-0.45) * edgeFade * 2.2
			if intensity <= 0 {
				continue
			}
			c := theme.Mid
			existing := img.RGBAAt(x, y)
			img.Set(x, y, color.RGBA{
				R: Clamp(int(existing.R) + int(float64(c.R)*intensity*0.35)),
				G: Clamp(int(existing.G) + int(float64(c.G)*intensity*0.35)),
				B: Clamp(int(existing.B) + int(float64(c.B)*intensity*0.45)),
				A: 255,
			})
		}
	}
}

// DrawCore renders a soft glowing galactic center.
func DrawCore(img *image.RGBA, cx, cy float64, c color.RGBA, radius float64) {
	DrawCoreCustom(img, Width, Height, cx, cy, c, radius)
}

func DrawCoreCustom(img *image.RGBA, w, h int, cx, cy float64, c color.RGBA, radius float64) {
	r := int(radius * 3)
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > float64(r) {
				continue
			}
			falloff := math.Exp(-dist * dist / (2 * radius * radius))
			px := int(cx) + dx
			py := int(cy) + dy
			if px < 0 || px >= w || py < 0 || py >= h {
				continue
			}
			existing := img.RGBAAt(px, py)
			img.Set(px, py, color.RGBA{
				R: Clamp(int(existing.R) + int(float64(c.R)*falloff)),
				G: Clamp(int(existing.G) + int(float64(c.G)*falloff)),
				B: Clamp(int(existing.B) + int(float64(c.B)*falloff)),
				A: 255,
			})
		}
	}
}

// DrawStar renders a star with a size-dependent Gaussian glow.
func DrawStar(img *image.RGBA, x, y float64, c color.RGBA, size int) {
	DrawStarCustom(img, Width, Height, x, y, c, size)
}

func DrawStarCustom(img *image.RGBA, w, h int, x, y float64, c color.RGBA, size int) {
	ix, iy := int(x), int(y)
	if ix < 0 || ix >= w || iy < 0 || iy >= h {
		return
	}
	if size == 1 {
		img.Set(ix, iy, c)
		return
	}
	radius := float64(size)
	spread := size + 1
	for dy := -spread; dy <= spread; dy++ {
		for dx := -spread; dx <= spread; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			falloff := math.Exp(-dist * dist / (2 * radius * radius))
			nx, ny := ix+dx, iy+dy
			if nx < 0 || nx >= w || ny < 0 || ny >= h {
				continue
			}
			existing := img.RGBAAt(nx, ny)
			img.Set(nx, ny, color.RGBA{
				R: Clamp(int(existing.R) + int(float64(c.R)*falloff)),
				G: Clamp(int(existing.G) + int(float64(c.G)*falloff)),
				B: Clamp(int(existing.B) + int(float64(c.B)*falloff)),
				A: 255,
			})
		}
	}
}

func LerpColor(core, mid, edge color.RGBA, ratio float64) color.RGBA {
	if ratio < 0.5 {
		t := ratio * 2
		return color.RGBA{
			R: uint8(float64(core.R)*(1-t) + float64(mid.R)*t),
			G: uint8(float64(core.G)*(1-t) + float64(mid.G)*t),
			B: uint8(float64(core.B)*(1-t) + float64(mid.B)*t),
			A: 255,
		}
	}
	t := (ratio - 0.5) * 2
	return color.RGBA{
		R: uint8(float64(mid.R)*(1-t) + float64(edge.R)*t),
		G: uint8(float64(mid.G)*(1-t) + float64(edge.G)*t),
		B: uint8(float64(mid.B)*(1-t) + float64(edge.B)*t),
		A: 255,
	}
}

func StarSize() int {
	r := rand.Float64()
	switch {
	case r < 0.70:
		return 1
	case r < 0.90:
		return 2
	case r < 0.98:
		return 3
	default:
		return 4
	}
}

func Clamp(v int) uint8 {
	if v > 255 {
		return 255
	}
	if v < 0 {
		return 0
	}
	return uint8(v)
}

// --- Value noise / fBm ---

func buildPerm() [512]int {
	var p [512]int
	for i := 0; i < 256; i++ {
		p[i] = i
	}
	rand.Shuffle(256, func(i, j int) { p[i], p[j] = p[j], p[i] })
	for i := 0; i < 256; i++ {
		p[256+i] = p[i]
	}
	return p
}

func fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

func valueNoise2D(x, y float64, p [512]int) float64 {
	xi := int(math.Floor(x)) & 255
	yi := int(math.Floor(y)) & 255
	xf := x - math.Floor(x)
	yf := y - math.Floor(y)
	u, v := fade(xf), fade(yf)
	aa := float64(p[p[xi]+yi]) / 255.0
	ba := float64(p[p[xi+1]+yi]) / 255.0
	ab := float64(p[p[xi]+yi+1]) / 255.0
	bb := float64(p[p[xi+1]+yi+1]) / 255.0
	return aa*(1-u)*(1-v) + ba*u*(1-v) + ab*(1-u)*v + bb*u*v
}

func fbm(x, y float64, octaves int, p [512]int) float64 {
	val, amp, freq := 0.0, 0.5, 1.0
	for i := 0; i < octaves; i++ {
		val += amp * valueNoise2D(x*freq, y*freq, p)
		amp *= 0.5
		freq *= 2.0
	}
	return val
}
