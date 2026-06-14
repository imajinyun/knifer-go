package imgx

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	randutil "github.com/imajinyun/go-knifer/internal/rand"
)

func defaultRandomInt(max int) int { return randutil.RandomInt(max) }

// randomColor returns a random RGBA color.
func randomColor(randomInt func(max int) int) color.RGBA {
	return color.RGBA{
		R: randomByte(randomInt),
		G: randomByte(randomInt),
		B: randomByte(randomInt),
		A: 255,
	}
}

func randomByte(randomInt func(max int) int) uint8 {
	n := randomInt(256)
	if n < 0 {
		return 0
	}
	if n > 255 {
		return 255
	}
	return uint8(n)
}

func colorComponent8(v uint32) uint8 {
	v >>= 8
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// fillBackground fills the image with the specified background color.
func fillBackground(img *image.RGBA, bg color.Color) {
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)
}

// drawLine draws a line from (x0,y0) to (x1,y1) using Bresenham's algorithm.
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 >= x1 {
		sx = -1
	}
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy
	for {
		if x0 >= 0 && y0 >= 0 && x0 < img.Bounds().Dx() && y0 < img.Bounds().Dy() {
			img.Set(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// drawOval draws an ellipse outline with center (cx,cy) and radii (rx,ry).
func drawOval(img *image.RGBA, cx, cy, rx, ry int, c color.Color) {
	if rx <= 0 || ry <= 0 {
		return
	}
	for angle := 0.0; angle < 2*math.Pi; angle += 0.01 {
		px := cx + int(float64(rx)*math.Cos(angle))
		py := cy + int(float64(ry)*math.Sin(angle))
		if px >= 0 && py >= 0 && px < img.Bounds().Dx() && py < img.Bounds().Dy() {
			img.Set(px, py, c)
		}
	}
}

// drawChar draws one ASCII character using the built-in 5x7 bitmap font.
func drawChar(img *image.RGBA, ch byte, x, y int, scale int, c color.Color) {
	glyph := getGlyph(ch)
	for row := 0; row < fontHeight; row++ {
		for col := 0; col < fontWidth; col++ {
			if glyph[row]&(1<<(fontWidth-1-col)) != 0 {
				// Scale each glyph pixel.
				for sy := 0; sy < scale; sy++ {
					for sx := 0; sx < scale; sx++ {
						px := x + col*scale + sx
						py := y + row*scale + sy
						if px >= 0 && py >= 0 && px < img.Bounds().Dx() && py < img.Bounds().Dy() {
							img.Set(px, py, c)
						}
					}
				}
			}
		}
	}
}

// drawString draws evenly spaced characters with random colors, centered in the image.
func drawString(img *image.RGBA, code string, w, h int, scale int, colorFunc func() color.Color) {
	charW := fontWidth*scale + scale
	totalW := charW * len(code)
	startX := (w - totalW) / 2
	charH := fontHeight * scale
	startY := (h - charH) / 2
	for i := 0; i < len(code); i++ {
		c := colorFunc()
		drawChar(img, code[i], startX+i*charW, startY, scale, c)
	}
}

// shearX applies a sinusoidal distortion along the X direction.
func shearX(img *image.RGBA, bg color.Color, randomIntRange func(min, max int) int, randomInt func(max int) int) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	period := randomIntRange(w/4, w)
	if period == 0 {
		period = w
	}
	phase := float64(randomInt(2))
	for y := 0; y < h; y++ {
		d := int(float64(period>>1) * math.Sin(float64(y)/float64(period)+2*math.Pi*phase))
		for x := w - 1; x >= 0; x-- {
			srcX := x - d
			if srcX >= 0 && srcX < w {
				img.SetRGBA(x, y, img.RGBAAt(srcX, y))
			} else {
				r, g, b, a := bg.RGBA()
				img.SetRGBA(x, y, color.RGBA{
					R: colorComponent8(r),
					G: colorComponent8(g),
					B: colorComponent8(b),
					A: colorComponent8(a),
				})
			}
		}
	}
}

// shearY applies a sinusoidal distortion along the Y direction.
func shearY(img *image.RGBA, bg color.Color, randomIntRange func(min, max int) int) {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	period := randomIntRange(h/4, h)
	if period == 0 {
		period = h
	}
	for x := 0; x < w; x++ {
		d := int(float64(period>>1) * math.Sin(float64(x)/float64(period)+2*math.Pi*7.0/20.0))
		for y := h - 1; y >= 0; y-- {
			srcY := y - d
			if srcY >= 0 && srcY < h {
				img.SetRGBA(x, y, img.RGBAAt(x, srcY))
			} else {
				r, g, b, a := bg.RGBA()
				img.SetRGBA(x, y, color.RGBA{
					R: colorComponent8(r),
					G: colorComponent8(g),
					B: colorComponent8(b),
					A: colorComponent8(a),
				})
			}
		}
	}
}

// drawThickLine draws a thick line by filling a quadrilateral.
func drawThickLine(img *image.RGBA, x1, y1, x2, y2, thickness int, c color.Color) {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	length := math.Sqrt(dx*dx + dy*dy)
	if length == 0 {
		return
	}
	scale := float64(thickness) / (2 * length)
	ddx := -scale * dy
	ddy := scale * dx
	// Four corners of the thick line polygon.
	xp := [4]int{x1 + int(ddx), x1 - int(ddx), x2 - int(ddx), x2 + int(ddx)}
	yp := [4]int{y1 + int(ddy), y1 - int(ddy), y2 - int(ddy), y2 + int(ddy)}
	fillPolygon(img, xp[:], yp[:], c)
}

// fillPolygon fills a polygon using a simple scanline algorithm.
func fillPolygon(img *image.RGBA, xp, yp []int, c color.Color) {
	if len(xp) < 3 {
		return
	}
	minY, maxY := yp[0], yp[0]
	for _, y := range yp {
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	bounds := img.Bounds()
	n := len(xp)
	for y := minY; y <= maxY; y++ {
		var nodeX []int
		j := n - 1
		for i := 0; i < n; i++ {
			if (yp[i] < y && yp[j] >= y) || (yp[j] < y && yp[i] >= y) {
				nodeX = append(nodeX, xp[i]+int(float64(y-yp[i])/float64(yp[j]-yp[i])*float64(xp[j]-xp[i])))
			}
			j = i
		}
		// Sort nodeX with bubble sort.
		for i := 0; i < len(nodeX)-1; i++ {
			for j := 0; j < len(nodeX)-i-1; j++ {
				if nodeX[j] > nodeX[j+1] {
					nodeX[j], nodeX[j+1] = nodeX[j+1], nodeX[j]
				}
			}
		}
		for i := 0; i+1 < len(nodeX); i += 2 {
			for x := nodeX[i]; x <= nodeX[i+1]; x++ {
				if x >= 0 && y >= 0 && x < bounds.Dx() && y < bounds.Dy() {
					img.Set(x, y, c)
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
