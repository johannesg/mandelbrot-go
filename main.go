package main

// import "math"
import (
	"fmt"
	"math/cmplx"
	"runtime"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"unsafe"
)

type Point struct {
	x     int
	y     int
	color byte
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() + 1)
	const width, height = 800, 600
	const topLeft, bottomRight = complex(-2, 1), complex(1, -1)

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)

	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)

	if err != nil {
		panic(err)
	}

	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STREAMING, width, height)
	if err != nil {
		panic(err)
	}

	defer texture.Destroy()

	inputChan := make(chan Point)
	resultChan := make(chan Point)

	numRoutines := runtime.NumCPU()
	for i := 0; i < numRoutines; i++ {
		go renderMandelbrot(width, height, topLeft, bottomRight, inputChan, resultChan)
	}

	go generateInput(width, height, inputChan)

	running := true

	tick := time.Tick(time.Millisecond * 50)
	tick2 := time.Tick(time.Second)

	bytesPerPixel := 4
	buffer := make([]byte, width*height*bytesPerPixel)
	pitch := width * bytesPerPixel

	numberOfPixelsProcessed := 0

	for running {

		select {
		case point := <-resultChan:

			offset := pitch*point.y + point.x*bytesPerPixel

			// fmt.Printf("Result: %v. Offset: %v. Pitch: %v\n", point, offset, pitch)
			buffer[offset] = point.color
			buffer[offset+1] = point.color
			buffer[offset+2] = point.color
			buffer[offset+3] = 255
			numberOfPixelsProcessed++

		case <-tick:
			for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
				switch e.(type) {
				case *sdl.QuitEvent:
					running = false
				case *sdl.KeyUpEvent:
					running = false
				}
			}
			texture.Update(nil, unsafe.Pointer(&buffer[0]), pitch)
			renderer.Clear()
			renderer.Copy(texture, nil, nil)
			renderer.Present()
		case <-tick2:
			fmt.Printf("Pixels: %v\n", numberOfPixelsProcessed)
		}
	}

}

func generateInput(width int, height int, input chan<- Point) {
	fmt.Printf("Generating input. W: %v, H: %v\n", width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := Point{x: x, y: y}
			input <- p
		}
	}
	fmt.Println("Generation complete")
}

func renderMandelbrot(width float64, height float64, topLeft complex128, bottomRight complex128, input <-chan Point, output chan<- Point) {
	// scale := complex(
	// 	(real(bottomRight)-real(topLeft))/width,
	// 	(imag(bottomRight)-imag(topLeft))/height)
	delta := (bottomRight - topLeft)
	scale := complex(real(delta)/width, imag(delta)/height)

	fmt.Printf("Start rendering, topLeft: %v, bottomRight: %v, scale: %v\n", topLeft, bottomRight, scale)
	for {
		point := <-input

		dot := complex(real(scale)*float64(point.x)+real(topLeft), imag(scale)*float64(point.y)+imag(topLeft))
		point.color = calcDot(dot)

		// fmt.Printf("X: %v, Y: %v, dot: %v, n: %v\n", point.x, point.y, dot, point.color)
		output <- point
	}
	fmt.Println("Stop rendering")
}

func calcDot(c complex128) byte {
	var z complex128
	const iterations int = 10000

	var n int = 0
	absz := 0.0
	for ; n < iterations && absz < 4.0; n++ {
		z = z*z + c
		absz = cmplx.Abs(z)
	}
	return byte((n * 255) / iterations)
}
