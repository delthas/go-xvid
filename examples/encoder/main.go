package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"

	"github.com/delthas/go-xvid"
)

// example of a custom plugin
type MyPlugin struct {
}

func (p MyPlugin) Info() xvid.PluginFlag            { return xvid.PluginRequireDiffQuantizer }
func (p MyPlugin) Init(create xvid.PluginInit) bool { return true }
func (p MyPlugin) Close(close xvid.PluginClose)     {}
func (p MyPlugin) Before(data *xvid.PluginData)     {}
func (p MyPlugin) Frame(data *xvid.PluginData)      {}
func (p MyPlugin) After(data *xvid.PluginData) {
	// example read access to the frame quantizer
	// note this is also provided in the encoding stats so a plugin isnt needed just for this simple usage
	fmt.Println("quantizer used for frame: ", data.Quantizer)
}

func main() {
	info, err := xvid.GetGlobalInfo()
	if err != nil {
		panic(err)
	}
	if err := xvid.InitWithFlags(info.CPUFlags, xvid.DebugError); err != nil {
		panic(err)
	}

	init := xvid.NewEncoderInit(800, 800, xvid.Fraction{25, 1}, []xvid.Plugin{
		// rate-control is provided by standard xvid plugins
		xvid.PluginRC1Pass(xvid.PluginRC1PassInit{
			Bitrate: 250 * 1000, // 250 kbps
		}),
		MyPlugin{},
	})
	// disable B-frames (for this example's sake)
	init.MaxBFrames = 0

	encoder, err := xvid.NewEncoder(init)
	if err != nil {
		panic(err)
	}
	defer encoder.Close()

	of, err := os.Create("examples/data/stream.dat")
	if err != nil {
		panic(err)
	}
	defer of.Close()

	// output buffer, re-alloc style, will be allocated & reused by the encoder
	// reuse in all Encode calls for best performance
	var output []byte

	for i := 0; i < 10; i++ {
		f, err := os.Open(fmt.Sprintf("examples/data/input-%d.jpg", i))
		if err != nil {
			panic(err)
		}
		inputJpeg, err := jpeg.Decode(f)
		if err != nil {
			panic(err)
		}

		// for this example's purpose all images are known to be YUV 420
		// the xvid colorspace is chosen appropriately
		// the width & height is also known and hardcoded in the encoder Init
		// but in a real use case could be obtained from the image bounds
		input := inputJpeg.(*image.YCbCr)
		img := xvid.Image{
			Colorspace: xvid.ColorSpacePlanar,
			Planes:     [][]byte{input.Y, input.Cb, input.Cr},
			Strides:    []int{input.YStride, input.CStride, input.CStride},
		}

		n, _, err := encoder.Encode(xvid.EncoderFrame{
			Input:  &img,
			Output: &output,
		})
		if err != nil {
			panic(err)
		}
		_, err = of.Write(output[:n])
		if err != nil {
			panic(err)
		}
	}
}
