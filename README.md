# Artemis Galaxy Generator

Inspired by the Artemis II mission. A generative art tool that renders spiral galaxies, galactic collisions, and animated GIFs — using only the Go standard library.

![collision example](collision.gif)

---

## Programs

| Command | Output |
|---|---|
| `go run ./cmd/generate` | `galaxy.png` — single galaxy |
| `go run ./cmd/collision` | `collision.png` — two galaxies colliding |
| `go run ./cmd/collision-gif` | `collision.gif` — animated collision |

---

## Project Structure

```
artemis-galaxy/
  galaxy/               # shared package — all generation logic
    galaxy.go
  cmd/
    generate/           # single galaxy CLI
    collision/          # static collision image
    collision-gif/      # animated collision GIF
  go.mod
```

---

## The Math

### 1. Spiral Arms — Logarithmic Spirals

Real galaxy arms follow a **logarithmic spiral**, where the radius grows exponentially with angle:

```
r = a · e^(b·θ)
```

- `θ` — angle around the center
- `b` — controls tightness of the wind (higher = looser arms)
- `a` — scale factor

We normalize `r` to `[0, maxRadius]` so all arms fit the canvas:

```
r = maxRadius · (e^(b·θ) - 1) / (e^(b·4π) - 1)
```

`θ` sweeps `[0, 4π]` (two full rotations). Each arm is the same spiral rotated by `2π · armIndex / numArms`.

Stars are scattered using **Gaussian noise** on both `r` and `θ`, with scatter proportional to `r` — tighter near the core, looser at the edge:

```
r     += r · 0.18 · N(0,1)
θ     += 0.15 · N(0,1)
```

### 2. Other Galaxy Shapes

| Shape | Distribution |
|---|---|
| **Elliptical** | Bivariate Gaussian: `x ~ N(cx, σx)`, `y ~ N(cy, σy)` with `σx ≠ σy` to stretch on one axis |
| **Ring** | Uniform angle, `r ~ N(ringRadius, ringWidth)` — Gaussian cross-section through an annulus |
| **Irregular** | 6–10 random Gaussian clumps placed asymmetrically, each with its own `σ` |

### 3. Nebula Clouds — Fractal Brownian Motion

Nebula gas clouds use **fBm** (fractional Brownian motion): layered octaves of smooth value noise where each octave doubles the frequency and halves the amplitude.

```
fBm(x, y) = Σ amplitude_i · valueNoise(x · frequency_i, y · frequency_i)
           where amplitude_i = 0.5^i,  frequency_i = 2^i
```

6 octaves are used. Large octaves define the cloud shapes; fine octaves add detail.

The underlying value noise interpolates between a shuffled permutation table using a **quintic ease curve** to avoid grid artifacts:

```
fade(t) = 6t⁵ - 15t⁴ + 10t³
```

Nebula is painted **additively** — pixels add onto the black background — and faded near the galaxy edge using:

```
edgeFade = max(0, 1 - dist / maxRadius)
intensity = max(0, fBm(x,y) - 0.45) · edgeFade · 2.2
```

### 4. Star Size — Gaussian Glow

Stars are drawn as one of four size tiers, weighted so large stars are rare:

| Size | Probability | Pixels |
|---|---|---|
| 1 | 70% | single pixel |
| 2 | 20% | 3px soft glow |
| 3 | 8% | 5px glow |
| 4 | 2% | 7px giant |

For sizes > 1, each surrounding pixel is lit with a **Gaussian kernel** using additive blending:

```
falloff(dx, dy) = e^( -(dx² + dy²) / (2σ²) )
pixel += starColor · falloff
```

This means overlapping stars and nebula regions bloom brighter together rather than overwriting each other.

### 5. Galactic Core

The bright central bulge uses the same Gaussian kernel at a larger radius, applied additively over whatever stars are already rendered:

```
falloff(d) = e^( -d² / (2 · coreRadius²) )
```

### 6. Tidal Force — Collision Physics

Each star is displaced toward the other galaxy's center using a simplified **inverse-square gravitational force**:

```
F = G / dist²
displacement = F · normalize(otherCenter - starPos)
```

A **tangential curl** component is added perpendicular to the pull direction to create the characteristic arcing tidal tails:

```
curl = F · 0.4 · ln(1 + dist/200)
displacement += curl · tangent(pullDirection)
```

Without the tangential term, stars collapse straight inward. With it, outer stars arc outward before being captured — matching real tidal tail morphology.

`G = 8,000,000` is hand-tuned so a star at `dist=300px` moves ~80px at full intensity.

### 7. GIF Animation

40 frames interpolate from `t = 0.0 → 1.0`. Galaxy positions follow a **smoothstep** ease curve:

```
tEased = t² · (3 - 2t)
```

This makes the approach start slow, accelerate, and ease into the final merge — mimicking gravitational acceleration. A slight sine arc on the vertical axis gives the orbital path curvature:

```
separationX = 280 - tEased · 200
separationY = sin(tEased · π/2) · 60
```

Stars are generated once at origin and recentered each frame — no per-frame regeneration. Each RGBA frame is quantized to 256 colors using **Floyd-Steinberg dithering** for GIF encoding.

---

## Usage Examples

```bash
# Milky Way-style
go run ./cmd/generate
# → 5000 stars, spiral, 4 arms, blue

# Head-on collision (static)
go run ./cmd/collision
# → G1: 4000 stars, spiral, 3 arms, blue
# → G2: 4000 stars, spiral, 3 arms, red
# → intensity: 0.7

# Animated collision
go run ./cmd/collision-gif
# → G1: 3000 stars, spiral, 3 arms, blue
# → G2: 3000 stars, spiral, 3 arms, red
```

---

## Dependencies

None. Only the Go standard library.
