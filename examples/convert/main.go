package main

import (
	"image"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/delthas/go-xvid"
)

func main() {
	// for this example let's pretend we want to disable SSE 4.1
	info, err := xvid.GetGlobalInfo()
	if err != nil {
		panic(err)
	}
	if err := xvid.InitWithFlags(info.CPUFlags & ^xvid.CPU_SSE41, 0); err != nil {
		panic(err)
	}

	f, err := os.Open("examples/data/convert-input.jpg")
	if err != nil {
		panic(err)
	}
	inputJpeg, err := jpeg.Decode(f)
	f.Close()
	if err != nil {
		panic(err)
	}

	// for this example's purpose the image is known to be YUV 420
	// the xvid colorspace is chosen appropriately
	inputYUV := inputJpeg.(*image.YCbCr)
	width := inputYUV.Bounds().Dx()
	height := inputYUV.Bounds().Dy()

	input := xvid.Image{
		Colorspace: xvid.ColorSpacePlanar,
		Planes:     [][]byte{inputYUV.Y, inputYUV.Cb, inputYUV.Cr},
		Strides:    []int{inputYUV.YStride, inputYUV.CStride},
	}

	output := xvid.Image{
		// output colorspace
		Colorspace: xvid.ColorSpaceRGBA,
		// let's flip this image vertically, for this example's sake
		VerticalFlip: true,
		// data will be filled in by Decode
	}

	if err := xvid.Convert(input, &output, width, height, false); err != nil {
		panic(err)
	}

	outputRGBA := image.NewRGBA(image.Rectangle{Max: image.Point{X: width, Y: height}})
	outputRGBA.Pix = output.Planes[0]
	outputRGBA.Stride = output.Strides[0]

	f, err = os.Create("examples/data/convert-output.png")
	if err != nil {
		panic(err)
	}
	if err = png.Encode(f, outputRGBA); err != nil {
		panic(err)
	}
	f.Close()
}
