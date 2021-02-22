package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
	"image/color"
	"log"
	"math"
	"math/cmplx"
)

const (
	minIterations = 200
	maxIterations = 1000
	increment     = 50
	minX          = -2.0
	maxX          = 1.25
	minY          = -1.25
)

func main() {
	pixelgl.Run(run)
}

type State string

var selectingRegionState State = "selecting"
var viewingState State = "viewing"

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Mandelbrot!",
		Bounds: pixel.R(0, 0, 1280, 984),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		log.Fatal(err)
	}

	windowWidth := cfg.Bounds.Max.X - cfg.Bounds.Min.X
	windowHeight := cfg.Bounds.Max.Y - cfg.Bounds.Min.Y
	screenRatio := windowWidth / windowHeight

	maxY := (maxX-minX)*cfg.Bounds.Max.Y/cfg.Bounds.Max.X + minY
	region := pixel.R(minX, minY, maxX, maxY)

	iterations := minIterations

	pic := renderView(cfg.Bounds, region, iterations)
	sprite := pixel.NewSprite(pic, pic.Bounds())
	sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

	state := viewingState
	var startPos pixel.Vec
	for !win.Closed() {
		win.Clear(colornames.Black)
		sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		if win.Pressed(pixelgl.MouseButtonLeft) && state == viewingState {
			startPos = win.MousePosition()
			state = selectingRegionState
		} else if win.Pressed(pixelgl.MouseButtonLeft) && state == selectingRegionState {
			imd := imdraw.New(nil)
			imd.Color = pixel.RGB(1, 1, 1)
			imd.Push(startPos)
			imd.Push(win.MousePosition())
			imd.Rectangle(2.0)
			imd.Draw(win)
		}

		if win.JustReleased(pixelgl.MouseButtonLeft) && state == selectingRegionState {
			endPos := win.MousePosition()
			region = adjustedRegionWithRatio(win.Bounds(), pixel.R(math.Min(startPos.X, endPos.X), math.Min(startPos.Y, endPos.Y), math.Max(endPos.X, startPos.X), math.Max(endPos.Y, startPos.Y)), screenRatio, region)
			iterations += increment

			pic = renderView(cfg.Bounds, region, iterations)
			sprite = pixel.NewSprite(pic, pic.Bounds())
			sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
			state = viewingState
		}

		win.Update()
	}
}

func adjustedRegionWithRatio(screen pixel.Rect, region pixel.Rect, widthToHeightRatio float64, logicalRegion pixel.Rect) (adjustedLogical pixel.Rect) {
	screenWidth := screen.Max.X - screen.Min.X
	screenHeight := screen.Max.Y - screen.Min.Y
	regionWidth := region.Max.X - region.Min.X
	regionHeight := region.Max.Y - region.Min.Y
	logicalWidth := logicalRegion.Max.X - logicalRegion.Min.X
	logicalHeight := logicalRegion.Max.Y - logicalRegion.Min.Y

	if regionWidth/regionHeight > widthToHeightRatio {
		correctedHeight := regionWidth / widthToHeightRatio
		heightIncrease := correctedHeight - regionHeight

		region.Min.Y -= heightIncrease / 2
		region.Max.Y += heightIncrease / 2
	} else {
		correctedWidth := regionHeight * widthToHeightRatio
		widthIncrease := correctedWidth - regionWidth

		region.Min.X -= widthIncrease / 2
		region.Max.X += widthIncrease / 2
	}

	adjustedLogical.Min.X = logicalRegion.Min.X + region.Min.X/screenWidth*logicalWidth
	adjustedLogical.Max.X = logicalRegion.Min.X + region.Max.X/screenWidth*logicalWidth
	adjustedLogical.Min.Y = logicalRegion.Min.Y + region.Min.Y/screenHeight*logicalHeight
	adjustedLogical.Max.Y = logicalRegion.Min.Y + region.Max.Y/screenHeight*logicalHeight

	return adjustedLogical
}

func renderView(screen pixel.Rect, region pixel.Rect, maxIterations int) (pic *pixel.PictureData) {
	pic = pixel.MakePictureData(screen)
	width := screen.Max.X
	height := screen.Max.Y

	palette := buildPalette(maxIterations)

	dx := (region.Max.X - region.Min.X) / width
	dy := (region.Max.Y - region.Min.Y) / height

	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			c := complex(region.Min.X+float64(x)*float64(dx), region.Min.Y+float64(y)*float64(dy))

			ci := getPixelColorIndex(c, minIterations)
			pic.Pix[y*int(width)+x] = palette[ci]
		}
	}

	return pic
}

func buildPalette(size int) (palette []color.RGBA) {
	palette = make([]color.RGBA, size)

	for i := 0; i < size/4; i++ {
		palette[i] = color.RGBA{uint8((float64(i) / (float64(size) / 4) * 255)), 0, 0, 255}
	}

	start := size/4 - 1
	for i := start; i < size-1; i++ {
		palette[i] = color.RGBA{255, uint8((float64(i) - float64(start)) / (float64(size) - float64(start) - 1.0) * 0.77647058823529411764705882352941 * 255), 0, 255}
	}
	// Last color is always black
	palette[size-1] = color.RGBA{0, 0, 0, 255}

	return palette
}

func getPixelColorIndex(value complex128, maxIterations int) (color int) {
	previousZ := complex(0.0, 0.0)

	n := 0
	for ; n < maxIterations-1 && cmplx.Abs(previousZ) <= complex(2, 0); n++ {
		z := cmplx.Pow(previousZ, 2) + value

		previousZ = z
	}

	return n
}
