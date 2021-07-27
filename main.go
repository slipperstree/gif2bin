package main

import (
	"flag"
	"fmt"
	"image/gif"
	"io"
	"math"
	"os"
	"sync"
)

var isCircular bool
var numLeds int
var ledOffset int
var isBit2 bool
var width int
var height int
var isC51Code bool

func convertGIF(inputFilename string) {
	var g *gif.GIF
	var input *os.File
	var output *os.File
	var err error
	var outputFilename string

	if input, err = os.Open(inputFilename); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to open input file: %s\n", inputFilename)
		return
	}
	defer input.Close()

	if g, err = gif.DecodeAll(input); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to decode GIF: %s\n", inputFilename)
		return
	}

	if isC51Code {
		outputFilename = inputFilename + ".c"
	} else {
		outputFilename = inputFilename + ".bin"
	}

	if output, err = os.OpenFile(outputFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to open output file: %s\n", outputFilename)
		return
	}
	defer output.Close()

	if isBit2 {
		convertGIFRectangularBit2(output, g)
	} else {
		if isCircular {
			convertGIFCircular(output, g)
		} else {
			convertGIFRectangular(output, g)
		}
	}

}

func convertGIFRectangular(output io.Writer, g *gif.GIF) {
	for _, image := range g.Image {
		for y := 0; y < image.Rect.Max.Y; y++ {
			for x := 0; x < image.Rect.Max.X; x++ {
				r, g, b, a := image.At(x, y).RGBA()
				r = r * a / 255
				g = g * a / 255
				b = b * a / 255
				output.Write([]byte{
					byte(r),
					byte(g),
					byte(b),
				})
			}
		}
	}

}

func getHexChar(decChar uint8) uint8 {
	if decChar < 10 {
		return 0x30 + decChar
	}

	if decChar == 10 {
		return 'A'
	}

	if decChar == 11 {
		return 'B'
	}

	if decChar == 12 {
		return 'C'
	}

	if decChar == 13 {
		return 'D'
	}

	if decChar == 14 {
		return 'E'
	}

	if decChar == 15 {
		return 'F'
	}

	return 0
}

func convertGIFRectangularBit2(output io.Writer, g *gif.GIF) {
	var byteData uint8
	var byteCnt uint
	var frames uint

	frames = 0

	println("w:", width, "h:", height)

	for _, image := range g.Image {
		frames++
		if frames > 3 {
			//return
		}
		println("F:", frames, "Rows:", image.Rect.Max.Y, "Cols", image.Rect.Max.X)
		for y := 0; y < height; y++ {
			byteData = 0
			for x := 0; x < width; x++ {
				if y >= image.Rect.Max.Y || x >= image.Rect.Max.X {
					// put 0
				} else {
					r, g, b, a := image.At(x, y).RGBA()
					// r,g,b from 0 - 65535
					if r < 20000 || g < 20000 || b < 20000 {
						// put 1 black
						byteData |= (1 << (7 - (x % 8)))
					} else {
						// put 0
					}
					r = r * a / 255
					g = g * a / 255
					b = b * a / 255
				}

				if x%8 == 7 {
					// byte data ready, output
					if isC51Code {
						output.Write([]byte{
							' ', '0', 'x',
							getHexChar((byteData / 16)),
							getHexChar((byteData % 16)),
							',',
						})
					} else {
						output.Write([]byte{
							byte(byteData),
						})
					}
					byteCnt++
					byteData = 0
				}
			}

			// check if have left bits
			if (width % 8) > 0 {
				if isC51Code {
					output.Write([]byte{
						' ', '0', 'x',
						getHexChar((byteData / 16)),
						getHexChar((byteData % 16)),
						',',
					})
				} else {
					output.Write([]byte{
						byte(byteData),
					})
				}
				byteCnt++
				byteData = 0
			}

			// New Line
			if isC51Code {
				output.Write([]byte{
					'\r', '\n',
				})
			}
		}
	}

	println(byteCnt)
}

func convertGIFCircular(output io.Writer, g *gif.GIF) {
	for _, image := range g.Image {
		centerX := float64((image.Rect.Max.X - image.Rect.Min.X) / 2)
		centerY := float64((image.Rect.Max.Y - image.Rect.Min.Y) / 2)

		radius := centerX
		if centerX > centerY {
			radius = centerY
		}

		for i := 0; i < 360; i++ {
			dx := radius * math.Cos(float64(i)/180*math.Pi) / float64(numLeds)
			dy := radius * math.Sin(float64(i)/180*math.Pi) / float64(numLeds)
			offsetX := dx * float64(ledOffset)
			offsetY := dy * float64(ledOffset)
			offsetRatio := float64(numLeds-ledOffset) / float64(numLeds)
			dx *= offsetRatio
			dy *= offsetRatio
			x := centerX + offsetX
			y := centerY + offsetY

			for j := 0; j < numLeds; j++ {
				r, g, b, a := image.At(int(x), int(y)).RGBA()
				r = r * a / 255
				g = g * a / 255
				b = b * a / 255
				output.Write([]byte{
					byte(r),
					byte(g),
					byte(b),
				})

				x += dx
				y += dy
			}
		}
	}

}

func init() {
	flag.BoolVar(&isCircular, "circular", false, "pack pixels in higher-res and circular way1")
	flag.IntVar(&numLeds, "num-leds", 0, "set number of leds (only required when using -circular)")
	flag.IntVar(&ledOffset, "led-offset", 0, "set led offset (only used when using -circular)")
	flag.BoolVar(&isBit2, "bit2", true, "output 2-bit color format(black and white only)")
	flag.IntVar(&width, "w", 50, "set width")
	flag.IntVar(&height, "h", 50, "set height")
	flag.BoolVar(&isC51Code, "c51", true, "output c51 array code")
}

func main() {
	var wg sync.WaitGroup

	flag.Parse()

	if isCircular && numLeds == 0 {
		fmt.Println("must specify -numleds higher than 0 when using -circular")
		return
	}

	for _, filename := range flag.Args() {
		wg.Add(1)

		go func() {
			convertGIF(filename)
			wg.Done()
		}()
	}

	wg.Wait()
}
