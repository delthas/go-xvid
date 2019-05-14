package main

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"

	"github.com/delthas/go-xvid"
)

func main() {
	fmt.Println(xvid.GetGlobalInfo())
	if err := xvid.Init(); err != nil {
		panic(err)
	}

	f, err := os.Open("examples/data/stream.dat") // raw xvid stream
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img := xvid.Image{
		// output colorspace (xvid automatically converts its internal format to the format specified)
		// can also use ColorSpaceInternal to directly use the decoder buffers without copy or conversion
		Colorspace: xvid.ColorSpaceRGBA,
		// data will be filled in by Decode
	}

	decoder, err := xvid.NewDecoder(xvid.DecoderInit{
		Input: f,
	})
	if err != nil {
		panic(err)
	}
	defer decoder.Close()

	i := 0
	for {
		_, stats, err := decoder.Decode(xvid.DecoderFrame{
			Output: &img,
		})
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		if stats.StatsFrame == nil { // some frames can be metadata (VOL) only, skip those for this example
			continue
		}
		output := image.NewRGBA(image.Rectangle{Max: image.Point{X: decoder.Width, Y: decoder.Height}})
		output.Pix = img.Planes[0]
		output.Stride = img.Strides[0]

		// the alpha channel is set to 0 instead of 255 due to an xvid implementation bug, fix this here
		for i := 0; i < output.Stride*decoder.Height; i += output.Stride {
			for j := 0; j < decoder.Width; j++ {
				output.Pix[i+j*4+3] = 255
			}
		}

		f, err := os.Create(fmt.Sprintf("examples/data/output-%d.png", i))
		if err != nil {
			panic(err)
		}
		if err = png.Encode(f, output); err != nil {
			panic(err)
		}
		f.Close()
		i++
	}
}
