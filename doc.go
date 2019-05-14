/*
go-xvid are Go bindings to xvidcore from Xvid 1.3.X (which uses the MPEG-4 Part 2, MPEG-4 Visual, ISO/IEC 14496-2 video codec standard).
This library can encode a sequence of images to an encoded Xvid stream, decode images
from an encoded Xvid stream, and convert images between different color spaces.

Xvid and containers

go-xvid only handles raw Xvid streams. Nearly all video files commonly found are stored in a media container, that encapsulate, but are not,
raw Xvid video streams. go-xvid cannot decode or encode container data, and the raw video streams must be encapsulated or decapsulated.

Documentation

go-xvid tries to not abbreviate names and identifiers so that all the names used can easily be searched on the Internet when they are not known. This means that this documentation will not redefine or explain common codec concepts like macroblocks, quantizers, rate-control, and such. Most of the complex configuration structures can be initialized to sane default values in case the user is not familiar with advanced encoding concepts.

Initialization

Before any other function in the package can be called, Init or InitWithFlags must be called
once to initialize all internal Xvid state. There is no Close method corresponding to the Init
call.

As an exception, GetGlobalInfo, which returns general information about the runtime Xvid build,
can be called at any time before and after Init.

Errors

go-xvid defines a specific error type, Error, which is used to represent internal xvidcore errors.

Images

Images in go-xvid is stored in the Image structure, which stores both an image color space and
its data as an array of planes, which are themselves arrays of data. Each plane has a specific
stride. The classic RGBA color space has only one plane and data array but some color spaces
can have up to three. See Image for more information.

Images can be converted from one color space to another with the Convert function.

Decoding

go-xvid can decode a sequence of images from a raw encoded Xvid stream. Decoder is the struct used to decode from a stream. A Decoder is created with NewDecoder, which
takes a DecoderInit configuration struct to initialize it.

Once created, Decoder.Decode can be called in a loop to decode a single frame at a time until
the entire stream has been processed. Each decoded frame contains extra statistics returned by
Decoder.Decode.

When the Decoder is no longer needed, it must be closed with Decoder.Close to free any internal data.

Encoding

go-xvid can encode a sequence of images to a raw encoded Xvid stream. Encoder is the struct used to encode from a stream. An Encoder is created with NewEncoder, which takes an EncoderInit configuration struct to initialize it, which itself should be initialized with NewEncoderInit to sane default values.

Once created, Encoder.Encode can be called in a loop to encode a single image at a time until
all the images have been processed. Each encoded frame contains extra statistics returned by
Encoder.Encode.

When the Encoder is no longer needed, it must be closed with Encoder.Close to free any internal data.

Plugins

Plugins are used to read and write internal frame data when encoding. Some standard plugins are defined in the library but custom ones can be created by implementing the Plugin interface.

In Xvid, rate-control is achieved by using plugins (for both 1-pass rate-control and 2-pass rate-control). You will probably need to use one of these rate-control plugins when encoding (otherwise the smallest quantizer is always used).

*/
package xvid
