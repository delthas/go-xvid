package xvid

// #cgo LDFLAGS: -lxvidcore
// #include "goxvid.h"
import "C"
import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

func cbool(b bool) C.int {
	if b {
		return 1
	}
	return 0
}

// Error is an internal Xvid error, which can be returned by most
// go-xvid functions.
type Error struct {
	code int
}

func (e *Error) Error() string {
	switch e.code {
	case C.XVID_ERR_FAIL:
		return "xvid: general fault"
	case C.XVID_ERR_MEMORY:
		return "xvid: memory allocation error"
	case C.XVID_ERR_FORMAT:
		return "xvid: file format error"
	case C.XVID_ERR_VERSION:
		return "xvid: version not supported"
	case C.XVID_ERR_END:
		return "xvid: end of stream reached"
	default:
		return "xvid: unknown error: code " + strconv.Itoa(e.code)
	}
}

func xvidErr(err C.int) *Error {
	return &Error{int(err)}
}

// QuantizerRange specifies the allowed range of a quantization parameter.
type QuantizerRange struct {
	// minimum quantizer value, inclusive, 0 defaults to 2, must be between 1 and 31
	Min int
	// maximum quantizer value, inclusive, 0 defaults to 31, must be between 1 and 31
	Max int
}

// CPUFlag is a flag (or a bitwise-or union of flags) for CPU-specific features.
type CPUFlag uint

const (
	CPU_ASM      CPUFlag = C.XVID_CPU_ASM
	CPU_MMX      CPUFlag = C.XVID_CPU_MMX
	CPU_MMXEXT   CPUFlag = C.XVID_CPU_MMXEXT
	CPU_SSE      CPUFlag = C.XVID_CPU_SSE
	CPU_SSE2     CPUFlag = C.XVID_CPU_SSE2
	CPU_SSE3     CPUFlag = C.XVID_CPU_SSE3
	CPU_SSE41    CPUFlag = C.XVID_CPU_SSE41
	CPU_3DNOW    CPUFlag = C.XVID_CPU_3DNOW
	CPU_3DNOWEXT CPUFlag = C.XVID_CPU_3DNOWEXT
	CPU_TSC      CPUFlag = C.XVID_CPU_TSC
)

// DebugFlag is a flag (or a bitwise-or union of flags) for printing of specific types of debug messages to standard error.
type DebugFlag uint

const (
	DebugError         DebugFlag = C.XVID_DEBUG_ERROR
	DebugStartCode     DebugFlag = C.XVID_DEBUG_STARTCODE
	DebugHeader        DebugFlag = C.XVID_DEBUG_HEADER
	DebugTimecode      DebugFlag = C.XVID_DEBUG_TIMECODE
	DebugMacroBlocks   DebugFlag = C.XVID_DEBUG_MB
	DebugCoefficients  DebugFlag = C.XVID_DEBUG_COEFF
	DebugMotionVectors DebugFlag = C.XVID_DEBUG_MV
	DebugRateControl   DebugFlag = C.XVID_DEBUG_RC
)

var DebugDebug DebugFlag = DebugFlag(C.DEBUG_DEBUG)

// Version is a semver (semantic versioning) version.
type Version struct {
	version int
}

// Major returns the major number part of the version: 1.2.3 returns 1.
func (v Version) Major() int {
	return (v.version >> 16) & 0xff
}

// Minor returns the minor number part of the version: 1.2.3 returns 2.
func (v Version) Minor() int {
	return (v.version >> 8) & 0xff
}

// Minor returns the patch number part of the version: 1.2.3 returns 3.
func (v Version) Patch() int {
	return v.version & 0xff
}

// String returns a readable string respresentation of the version, e.g. 1.2.3.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
}

// ColorSpace is the color space of an Image.
// See https://fourcc.org/yuv.php for details about each color space.
type ColorSpace struct {
	value int
	// number of image planes for colorspace
	Planes int
	// bits per pixel average (over all planes)
	BitsPerPixel int
	// bits per pixel for each plane
	BitsPerPixelPlanes []int
}

var (
	// YUV 4:2:0 planar, like ColorSpaceI420 but with 3 buffers, planes[0] is Y, planes[1] is U, planes[2] is V
	ColorSpacePlanar ColorSpace = ColorSpace{C.XVID_CSP_PLANAR, 3, 12, []int{8, 2, 2}}
	// YUV 4:2:0 planar, packed as YUV
	ColorSpaceI420 ColorSpace = ColorSpace{C.XVID_CSP_I420, 1, 12, []int{12}}
	// YUV 4:2:0 planar, packed as YVU
	ColorSpaceYV12 ColorSpace = ColorSpace{C.XVID_CSP_YV12, 1, 12, []int{12}}
	// YUV 4:2:2 packed
	ColorSpaceYUY2 ColorSpace = ColorSpace{C.XVID_CSP_YUY2, 1, 16, []int{16}}
	// YUV 4:2:2 packed
	ColorSpaceUYVY ColorSpace = ColorSpace{C.XVID_CSP_UYVY, 1, 16, []int{16}}
	// YUV 4:2:2 packed
	ColorSpaceYVYU ColorSpace = ColorSpace{C.XVID_CSP_YVYU, 1, 16, []int{16}}
	// 24-bit RGB packed
	ColorSpaceRGB ColorSpace = ColorSpace{C.XVID_CSP_RGB, 1, 24, []int{24}}
	// 32-bit BGRA packed; because of an Xvid bug, the alpha channel will be cleared out to 0 (fully transparent) instead of 255 (fully opaque)
	ColorSpaceBGRA ColorSpace = ColorSpace{C.XVID_CSP_BGRA, 1, 32, []int{32}}
	// 32-bit ABGR packed; because of an Xvid bug, the alpha channel will be cleared out to 0 (fully transparent) instead of 255 (fully opaque)
	ColorSpaceABGR ColorSpace = ColorSpace{C.XVID_CSP_ABGR, 1, 32, []int{32}}
	// 32-bit RGBA packed; because of an Xvid bug, the alpha channel will be cleared out to 0 (fully transparent) instead of 255 (fully opaque)
	ColorSpaceRGBA ColorSpace = ColorSpace{C.XVID_CSP_RGBA, 1, 32, []int{32}}
	// 32-bit ARGB packed; because of an Xvid bug, the alpha channel will be cleared out to 0 (fully transparent) instead of 255 (fully opaque)
	ColorSpaceARGB ColorSpace = ColorSpace{C.XVID_CSP_ARGB, 1, 32, []int{32}}
	// 24-bit BGR packed
	ColorSpaceBGR ColorSpace = ColorSpace{C.XVID_CSP_BGR, 1, 24, []int{24}}
	// 16-bit RGB555 packed
	ColorSpaceRGB555 ColorSpace = ColorSpace{C.XVID_CSP_RGB555, 1, 16, []int{16}}
	// 16-bit RGB565 packed
	ColorSpaceRGB565 ColorSpace = ColorSpace{C.XVID_CSP_RGB565, 1, 16, []int{16}}
	// only for decoding: YUV 4:2:0 planar, but uses internal decoder buffers and strides rather than copying to a buffer; invalid after any call to a Decoder method
	ColorSpaceInternal ColorSpace = ColorSpace{C.XVID_CSP_INTERNAL, 3, 12, []int{8, 2, 2}}
	// only for decoding: don't output anything
	ColorSpaceNoOutput ColorSpace = ColorSpace{C.XVID_CSP_NULL, 0, 0, []int{}}
	// TODO frame slice rendering support
	// decoder only: 4:2:0 planar, per slice rendering
	// ColorSpaceSLICE    = ColorSpace{C.XVID_CSP_SLICE, 3}
)

// DecoderFlag is a flag (or a bitwise-or union of flags) for decoding a frame, set in each frame.
type DecoderFlag uint

const (
	// lowdelay mode
	DecoderLowDelay DecoderFlag = C.XVID_LOWDELAY
	// indicate break/discontinuity in streaming
	DecoderDiscontinuity DecoderFlag = C.XVID_DISCONTINUITY
	// perform luma deblocking
	DecoderDeblockLuma DecoderFlag = C.XVID_DEBLOCKY
	// perform chroma deblocking
	DecoderDeblockChroma DecoderFlag = C.XVID_DEBLOCKUV
	// perform luma deringing, requires deblocking to work
	DecoderDeringLuma DecoderFlag = C.XVID_DERINGY
	// perform chroma deringing, requires deblocking to work
	DecoderDeringChroma DecoderFlag = C.XVID_DERINGUV
	// adds film grain
	DecoderFilmGrain DecoderFlag = C.XVID_FILMEFFECT
)

// EncoderFlag is a flag (or a bitwise-or union of flags) for encoding frames, set in Encoder.Init.
type EncoderFlag uint

const (
	// packed B-frames; strongly discouraged
	EncoderPacked EncoderFlag = C.XVID_GLOBAL_PACKED
	// closed GOP
	EncoderClosedGOP EncoderFlag = C.XVID_GLOBAL_CLOSED_GOP
	// require plugins to use the original image for PSNR calculation
	EncoderEnableExtraStats EncoderFlag = C.XVID_GLOBAL_EXTRASTATS_ENABLE
	// write DivX5 userdata string, implied by EncoderPacked
	EncoderWriteDivX5UserData EncoderFlag = C.XVID_GLOBAL_DIVX5_USERDATA
)

// VOLFlag is a flag (or a bitwise-or union of flags) for encoding a group of frames, set in Encoder.Encode.
type VOLFlag uint

const (
	// enable MPEG type quantization
	VOLMPEGQuantization VOLFlag = C.XVID_VOL_MPEGQUANT
	// enable plane sse stats
	VOLExtraStats VOLFlag = C.XVID_VOL_EXTRASTATS
	// enable quarterpel: frames will encoded as quarterpel
	VOLQuarterPixel VOLFlag = C.XVID_VOL_QUARTERPEL
	// enable GMC (global motion compensation); frames will be checked for gmc suitability
	VOLGMC VOLFlag = C.XVID_VOL_GMC
	// enable interlaced encoding
	VOLInterlacing VOLFlag = C.XVID_VOL_INTERLACING
)

// VOPFlag is a flag (or a bitwise-or union of flags) for encoding a single frame, set in Encoder.Encode.
type VOPFlag uint

const (
	// print debug messages in frames
	VOPDebug VOPFlag = C.XVID_VOP_DEBUG
	// use halfpel interpolation
	VOPHalfPixel VOPFlag = C.XVID_VOP_HALFPEL
	// use 4 motion vectors per MB
	VOPInter4Vectors VOPFlag = C.XVID_VOP_INTER4V
	// use trellis based R-D "optimal" quantization
	VOPTrellisQuantization VOPFlag = C.XVID_VOP_TRELLISQUANT
	// enable chroma optimization pre-filter
	VOPChromaOptimization VOPFlag = C.XVID_VOP_CHROMAOPT
	// use 'cartoon mode'
	VOPCartoon VOPFlag = C.XVID_VOP_CARTOON
	// enable greyscale only mode (even for  color input material chroma is ignored)
	VOPGreyscale VOPFlag = C.XVID_VOP_GREYSCALE
	// high quality ac prediction
	VOPHighQualityACPrediction VOPFlag = C.XVID_VOP_HQACPRED
	// enable DCT-ME and use it for mode decision
	VOPModeDecisionRD VOPFlag = C.XVID_VOP_MODEDECISION_RD
	// use simplified R-D mode decision
	VOPFastModeDecisionRD VOPFlag = C.XVID_VOP_FAST_MODEDECISION_RD
	// enable rate-distortion mode decision in b-frames
	VOPRateDistortionBFrames VOPFlag = C.XVID_VOP_RD_BVOP
	// use PSNR-HVS-M as metric for rate-distortion optimizations
	VOPRateDistortionPSNRHVSM VOPFlag = C.XVID_VOP_RD_PSNRHVSM
	// only valid with VOLInterlacing, set upper-field-first flag
	VOPUpperFieldFirst VOPFlag = C.XVID_VOP_TOPFIELDFIRST
	// only valid with VOLInterlacing, set alternate vertical scan flag
	VOPAlternateSscan VOPFlag = C.XVID_VOP_ALTERNATESCAN
)

// MotionFlag is a flag (or a bitwise-or union of flags) of motion estimation flags for encoding a single frame, set in Encoder.Encode.
type MotionFlag uint

const (
	// Motion Estimation Search Patterns

	// use advdiamonds instead of diamonds as search pattern
	MotionAdvancedDiamond16 MotionFlag = C.XVID_ME_ADVANCEDDIAMOND16
	// use advdiamond for MotionExtendSearch8
	MotionAdvancedDiamond8 MotionFlag = C.XVID_ME_ADVANCEDDIAMOND8
	// use squares instead of diamonds as search pattern
	MotionUseSquares16 MotionFlag = C.XVID_ME_USESQUARES16
	// use square for MotionExtendSearch8
	MotionUseSquares8 MotionFlag = C.XVID_ME_USESQUARES8

	// SAD operator based flags

	MotionHalfPixelRefine16    MotionFlag = C.XVID_ME_HALFPELREFINE16
	MotionHalfPixelRefine8     MotionFlag = C.XVID_ME_HALFPELREFINE8
	MotionQuarterPixelRefine16 MotionFlag = C.XVID_ME_QUARTERPELREFINE16
	MotionQuarterPixelRefine8  MotionFlag = C.XVID_ME_QUARTERPELREFINE8
	MotionGMERefine            MotionFlag = C.XVID_ME_GME_REFINE
	// extend PMV by more searches
	MotionExtendSearch16 MotionFlag = C.XVID_ME_EXTSEARCH16
	// use diamond/square for extended 8x8 search
	MotionExtendSearch8 MotionFlag = C.XVID_ME_EXTSEARCH8
	// also use chroma for P_VOP/S_VOP ME
	MotionChromaPFrame MotionFlag = C.XVID_ME_CHROMA_PVOP
	// also use chroma for B_VOP ME
	MotionChromaBFrame MotionFlag = C.XVID_ME_CHROMA_BVOP
	// use low-complexity refinement functions
	MotionFastRefine16 MotionFlag = C.XVID_ME_FASTREFINE16
	// low-complexity 8x8 sub-block refinement
	MotionFastRefine8 MotionFlag = C.XVID_ME_FASTREFINE8

	// Rate Distortion based flags

	// only valid with VOPModeDecisionRD, perform RD-based halfpel refinement
	MotionHalfPixelRefine16RD MotionFlag = C.XVID_ME_HALFPELREFINE16_RD
	// only valid with VOPModeDecisionRD, perform RD-based halfpel refinement for 8x8 mode
	MotionHalfPixelRefine8RD MotionFlag = C.XVID_ME_HALFPELREFINE8_RD
	// only valid with VOPModeDecisionRD, perform RD-based qpel refinement
	MotionQuarterPixelRefine16RD MotionFlag = C.XVID_ME_QUARTERPELREFINE16_RD
	// only valid with VOPModeDecisionRD, perform RD-based qpel refinement for 8x8 mode
	MotionQuarterPixelRefine8RD MotionFlag = C.XVID_ME_QUARTERPELREFINE8_RD
	// only valid with VOPModeDecisionRD, perform RD-based search using square pattern enable XVID_ME_EXTSEARCH8 to do this in 8x8 search as well
	MotionExtendSearchRD MotionFlag = C.XVID_ME_EXTSEARCH_RD
	// only valid with VOPModeDecisionRD, always check vector equal to prediction
	MotionCheckPredictionRD MotionFlag = C.XVID_ME_CHECKPREDICTION_RD

	// Other

	// speed-up ME by detecting stationary scenes
	MotionDetectStaticMotion MotionFlag = C.XVID_ME_DETECT_STATIC_MOTION
	// speed-up by skipping b-frame delta search
	MotionSkipDeltaSearch MotionFlag = C.XVID_ME_SKIP_DELTASEARCH
	// speed-up by partly skipping interpolate mode
	MotionFastModeInterpolate MotionFlag = C.XVID_ME_FAST_MODEINTERPOLATE
	// speed-up by early exiting b-search
	MotionBFrameEarlyStop MotionFlag = C.XVID_ME_BFRAME_EARLYSTOP
)

// EncoderProfile is a profile (and level) used for encoding; should be set to EncoderProfileAuto to detect automatically.
type EncoderProfile uint

const (
	// automatically choose profile
	EncoderProfileAuto EncoderProfile = 0
	// simple profile, level L0
	EncoderProfileS_L0 EncoderProfile = C.XVID_PROFILE_S_L0
	// simple profile, level L1
	EncoderProfileS_L1 EncoderProfile = C.XVID_PROFILE_S_L1
	// simple profile, level L2
	EncoderProfileS_L2 EncoderProfile = C.XVID_PROFILE_S_L2
	// simple profile, level L3
	EncoderProfileS_L3 EncoderProfile = C.XVID_PROFILE_S_L3
	// simple profile, level L4a
	EncoderProfileS_L4A EncoderProfile = C.XVID_PROFILE_S_L4a
	// simple profile, level L5
	EncoderProfileS_L5 EncoderProfile = C.XVID_PROFILE_S_L5
	// simple profile, level L6
	EncoderProfileS_L6 EncoderProfile = C.XVID_PROFILE_S_L6
	// advanced real-time simple profile, level L1
	EncoderProfileARTS_L1 EncoderProfile = C.XVID_PROFILE_ARTS_L1
	// advanced real-time simple profile, level L2
	EncoderProfileARTS_L2 EncoderProfile = C.XVID_PROFILE_ARTS_L2
	// advanced real-time simple profile, level L3
	EncoderProfileARTS_L3 EncoderProfile = C.XVID_PROFILE_ARTS_L3
	// advanced real-time simple profile, level L4
	EncoderProfileARTS_L4 EncoderProfile = C.XVID_PROFILE_ARTS_L4
	// advanced simple profile, level L0
	EncoderProfileAS_L0 EncoderProfile = C.XVID_PROFILE_AS_L0
	// advanced simple profile, level L1
	EncoderProfileAS_L1 EncoderProfile = C.XVID_PROFILE_AS_L1
	// advanced simple profile, level L2
	EncoderProfileAS_L2 EncoderProfile = C.XVID_PROFILE_AS_L2
	// advanced simple profile, level L3
	EncoderProfileAS_L3 EncoderProfile = C.XVID_PROFILE_AS_L3
	// advanced simple profile, level L4
	EncoderProfileAS_L4 EncoderProfile = C.XVID_PROFILE_AS_L4
)

// FrameType is the type of a frame that was decoded [D], that was encoded (in EncodeStats) [E], or to be encoded [S].
// Each fields description has a set of letters to show when the field is used.
type FrameType int

const (
	// [D] VOL (metadata) was decoded
	FrameTypeVOL FrameType = C.XVID_TYPE_VOL
	// [S] automatically determine coding type
	FrameTypeAuto FrameType = C.XVID_TYPE_AUTO
	// [D,E,S] intra frame
	FrameTypeI FrameType = C.XVID_TYPE_IVOP
	// [D,E,S] predicted frame
	FrameTypeP FrameType = C.XVID_TYPE_PVOP
	// [D,E,S] bidirectionally encoded
	FrameTypeB FrameType = C.XVID_TYPE_BVOP
	// [D,E,S] predicted+sprite frame
	FrameTypeS FrameType = C.XVID_TYPE_SVOP

	// never used in the public API
	// [E] nothing was encoded
	frameTypeNothing FrameType = C.XVID_TYPE_NOTHING
)

// ZoneType is a kind of bitrate Zone, which is applied on a range of frames while encoding.
type ZoneType uint

const (
	// enforce a specific quantizer, value is the quantizer, recommended range is 2-31, must be >= 1
	ZoneModeQuantizer ZoneType = C.XVID_ZONE_QUANT
	// enforce a specific frame weight, value is the weight, default weight is 1, must be > 0
	ZoneModeWeight ZoneType = C.XVID_ZONE_WEIGHT
)

// PluginFlag is a flag (or a bitwise-or union of flags) of data a custom Plugin needs access to.
type PluginFlag uint

const (
	// plugin needs a copy of the original (uncompressed) image
	PluginRequireOriginal PluginFlag = C.XVID_REQORIGINAL
	// plugin needs psnr between the uncompressed and compressed image
	PluginRequirePSNR PluginFlag = C.XVID_REQPSNR
	// plugin needs the diff quantizer table
	PluginRequireDiffQuantizer PluginFlag = C.XVID_REQDQUANTS
	// plugin needs the lambda table
	PluginRequireLambda PluginFlag = C.XVID_REQLAMBDA
)

// PixelAspectRatio is a frame pixel aspect ratio (PAR), given as an integer fraction of a pixel width and height.
// There are both standard frame pixel aspect ratios, defined in go-xvid, and user-defined pixel aspect ratios, which can be unexact due to precision loss (clamped to [1, 255]).
type PixelAspectRatio struct {
	// pixel width ratio
	Width int
	// pixel height ratio
	Height int
	value  int
}

var (
	// square pixel
	PixelAspectRatio11VGA PixelAspectRatio = PixelAspectRatio{1, 1, C.XVID_PAR_11_VGA}
	// 12:11 pixel (seldom used in 4:3 pal 625-line)
	PixelAspectRatio43PAL PixelAspectRatio = PixelAspectRatio{12, 11, C.XVID_PAR_43_PAL}
	// 10:11 pixel (seldom used in 4:3 pal 525-line)
	PixelAspectRatio43NTSC PixelAspectRatio = PixelAspectRatio{10, 11, C.XVID_PAR_43_NTSC}
	// 16:11 pixel (seldom used in 16:9 pal 625-line)
	PixelAspectRatio169PAL PixelAspectRatio = PixelAspectRatio{16, 11, C.XVID_PAR_169_PAL}
	// 40:33 pixel (seldom used in 16:9 ntsc 525-line)
	PixelAspectRatio169NTSC PixelAspectRatio = PixelAspectRatio{40, 33, C.XVID_PAR_169_NTSC}
)

func newPixelAspectRatio(width int, height int) PixelAspectRatio {
	return PixelAspectRatio{
		value:  C.XVID_PAR_EXT,
		Width:  width,
		Height: height,
	}
}

// Fraction is an exact integer fraction to represent a decimal number without precision loss.
type Fraction struct {
	Numerator int
	// must not be 0
	Denominator int
}

// Float returns the actual value of a fraction, which is Numerator / Denominator
func (f Fraction) Float() float32 {
	return float32(f.Numerator) / float32(f.Denominator)
}

// Image represents an input or output image data and its color space.
//
// The data is stored in multiple buffers, one per image plane; each plane has a specific stride (data size in bytes per line).
//
// When used as input, ColorSpace must be set to the actual color space of the image data;
// Planes and Strides must contain at least as many planes as the ColorSpace.Planes count,
// and each plane must contain enough image data (corresponding to the image bounds, color space, and stride).
// Strides can be set to 0 to assume compact data.
//
// When used as output, ColorSpace must be set to the desired color space of the image data;
// xvid will automatically convert the data from its internal color space to the target one.
// Planes can be nil, in which case it will be created to contain the output data;
// if it is not nil it must contain at least as many planes as the ColorSpace.Planes count.
// Each plane can be nil, in which case it will be created to contain the output data; if it is
// not nil it must contain enough image data (corresponding to the image bounds, color space, and stride).
// Strides can be nil, in which case it will be created and filled with zeroes.
// Each stride can be 0, in which case the value will be replaced with the actual data size per line,
// to have compact data.
//
// When used as output with the special ColorSpaceInternal color space, the strides will be ignored
// and replaced with the actual internal encoder buffer strides; and the planes buffers will be ignored
// and replaced with the internal encoder buffers. The data is valid until the next call to any of the Encoder methods.
type Image struct {
	// image color space, determines the number of planes
	Colorspace ColorSpace
	// whether to flip the image vertically, during converting (only set on the output image), decoding, or encoding
	VerticalFlip bool
	// image planes, each plane contains image data
	Planes [][]byte
	// planes strides (bytes per row)
	Strides []int
}

func (i *Image) nativeInput(width int, height int) (*C.xvid_image_t, error) {
	if len(i.Planes) != i.Colorspace.Planes {
		return nil, fmt.Errorf("xvid: unexpected number of planes for image, expected %d, got %d", i.Colorspace.Planes, len(i.Planes))
	}
	if i.Strides == nil {
		i.Strides = make([]int, i.Colorspace.Planes)
	} else if len(i.Strides) != i.Colorspace.Planes {
		return nil, fmt.Errorf("xvid: unexpected number of strides for image, expected %d, got %d", i.Colorspace.Planes, len(i.Planes))
	}
	if i.Colorspace.value == ColorSpaceInternal.value {
		return nil, fmt.Errorf("xvid: unexpected colorspace ColorSpaceInternal, use only for ouput")
	}
	var cPlanes [4]unsafe.Pointer
	var cStrides [4]C.int
	for j, v := range i.Planes {
		l := width * height * i.Colorspace.BitsPerPixelPlanes[j] / 8
		if len(v) < l {
			return nil, fmt.Errorf("xvid: not enough space in plane %d, need at least %d, got %d", j, l, len(v))
		}
		cPlanes[j] = unsafe.Pointer(&i.Planes[j][0])
		s := width * i.Colorspace.BitsPerPixelPlanes[j] / 8
		if i.Strides[j] == 0 {
			cStrides[j] = C.int(s)
		} else if i.Strides[j] < s {
			return nil, fmt.Errorf("xvid: insufficient stride in plane %d (strides is the total length of row, not just the offset), need at least %d, got %d", j, s, i.Strides[j])
		} else {
			cStrides[j] = C.int(i.Strides[j])
		}
	}
	return &C.xvid_image_t{
		csp:    C.int(i.Colorspace.value),
		plane:  cPlanes,
		stride: cStrides,
	}, nil
}

func (i *Image) nativeOutput(width int, height int) (*C.xvid_image_t, error) {
	if i.Planes == nil {
		i.Planes = make([][]byte, i.Colorspace.Planes)
	} else if len(i.Planes) != i.Colorspace.Planes {
		return nil, fmt.Errorf("xvid: unexpected number of planes for image, expected %d, got %d", i.Colorspace.Planes, len(i.Planes))
	}
	if i.Strides == nil {
		i.Strides = make([]int, i.Colorspace.Planes)
	} else if len(i.Strides) != i.Colorspace.Planes {
		return nil, fmt.Errorf("xvid: unexpected number of strides for image, expected %d, got %d", i.Colorspace.Planes, len(i.Planes))
	}
	var cPlanes [4]unsafe.Pointer
	var cStrides [4]C.int
	if width > 0 && height > 0 && i.Colorspace.value != ColorSpaceInternal.value {
		for j, v := range i.Planes {
			s := width * i.Colorspace.BitsPerPixelPlanes[j] / 8
			if i.Strides[j] == 0 {
				cStrides[j] = C.int(s)
				i.Strides[j] = s // TODO this replaces the auto-0 with a non-0 value, is it ok?
			} else if i.Strides[j] < s {
				return nil, fmt.Errorf("xvid: insufficient stride in plane %d (strides is the total length of row, not just the offset), need at least %d, got %d", j, s, i.Strides[j])
			} else {
				cStrides[j] = C.int(i.Strides[j])
			}
			l := s * height
			if v == nil {
				i.Planes[j] = make([]byte, l)
			} else if len(v) < l {
				return nil, fmt.Errorf("xvid: not enough space in plane %d, need at least %d, got %d", j, l, len(v))
			}
			cPlanes[j] = unsafe.Pointer(&i.Planes[j][0])
		}
	}
	csp := i.Colorspace.value
	if i.VerticalFlip {
		csp |= int(C.CSP_VFLIP)
	}
	return &C.xvid_image_t{
		csp:    C.int(csp),
		plane:  cPlanes,
		stride: cStrides,
	}, nil
}

// GlobalInfo stores global information about Xvid, obtained from GetGlobalInfo.
type GlobalInfo struct {
	// runtime version of xvidcore
	Version Version
	// runtime build string of xvidcore
	Build string
	// supported cpu features found
	CPUFlags CPUFlag
	// count of system threads found
	NumThreads int
}

// GetGlobalInfo returns global information about Xvid, can be called before any Init method.
// If an error is returned, no further Xvid functions are expected to work.
func GetGlobalInfo() (*GlobalInfo, error) {
	var cGlobalInfo C.xvid_gbl_info_t
	cGlobalInfo.version = C.XVID_VERSION
	if code := C.xvid_global(nil, C.XVID_GBL_INFO, unsafe.Pointer(&cGlobalInfo), nil); code != 0 {
		return nil, xvidErr(code)
	}
	build := ""
	if cGlobalInfo.build != nil {
		build = C.GoString(cGlobalInfo.build)
	}
	return &GlobalInfo{
		Version:    Version{int(cGlobalInfo.actual_version)},
		Build:      build,
		CPUFlags:   CPUFlag(cGlobalInfo.cpu_flags),
		NumThreads: int(cGlobalInfo.num_threads),
	}, nil
}

// Init initializes Xvid and must be called once before calling any other method, except GetGlobalInfo.
// Alternatively InitWithFlags can be used to specify custom CPU and debug flags.
// Init uses all the available CPU features and doesn't enable any debug.
// If an error is returned, initialization failed and no further Xvid functions are expected to work.
// There is no global Close() function corresponding to Init.
func Init() error {
	var cGlobalInit C.xvid_gbl_init_t
	cGlobalInit.version = C.XVID_VERSION
	if code := C.xvid_global(nil, C.XVID_GBL_INIT, unsafe.Pointer(&cGlobalInit), nil); code != 0 {
		return xvidErr(code)
	}
	return nil
}

// InitWithFlags initializes Xvid and must be called once before calling any other method, except GetGlobalInfo.
// Alternatively Init can be used to initialize with all available CPU features and no debug.
// If an error is returned, initialization failed and no further Xvid functions are expected to work.
// There is no global Close() function corresponding to InitWithFlags.
func InitWithFlags(cpuFlags CPUFlag, debugFlags DebugFlag) error {
	var cGlobalInit C.xvid_gbl_init_t
	cGlobalInit.version = C.XVID_VERSION
	cGlobalInit.cpu_flags = C.uint(cpuFlags | CPUFlag(C.CPU_FORCE))
	cGlobalInit.debug = C.int(debugFlags)
	if code := C.xvid_global(nil, C.XVID_GBL_INIT, unsafe.Pointer(&cGlobalInit), nil); code != 0 {
		return xvidErr(code)
	}
	return nil
}

// Converts converts an Image from a color space to another. Init (or InitWithFlags) must be called once before calling this function.
// An error can be returned because of invalid input or output images, or due to an internal Xvid error.
func Convert(input Image, output Image, width int, height int, interlacing bool) error {
	cInput, err := input.nativeOutput(width, height)
	if err != nil {
		return err
	}
	cOutput, err := output.nativeOutput(width, height)
	if err != nil {
		return err
	}
	cConvertInfo := C.xvid_gbl_convert_t{
		version:     C.XVID_VERSION,
		input:       *cInput,
		output:      *cOutput,
		width:       C.int(width),
		height:      C.int(height),
		interlacing: cbool(interlacing),
	}
	if code := C.xvid_global(nil, C.XVID_GBL_CONVERT, unsafe.Pointer(&cConvertInfo), nil); code != 0 {
		return xvidErr(code)
	}
	return nil
}

// Decoder is an initialized Xvid decoder.
// To create a Decoder, use NewDecoder.
// A Decoder must be closed after use, by calling its Close method.
// To decode a frame, use the Decode method.
type Decoder struct {
	// current frame width in pixels
	Width int
	// current frame height in pixels
	Height int
	handle unsafe.Pointer
	r      io.Reader
	buf    []byte
	i      int
	n      int
	eof    bool
	err    error // permanent error
}

// DecoderInit is information used to create a Decoder in NewDecoder.
// Its Input field must be set to the Reader from which to read an encoded raw Xvid stream data from.
type DecoderInit struct {
	// Reader from which to read encoded frame data.
	// the Reader will not be closed automatically, it has to be caller-closed after the Decoder is finished.
	Input io.Reader
	// optional initial frame width in pixels (can be automatically detected by the Decoder)
	Width int
	// optional initial frame height in pixels (can be automatically detected by the Decoder)
	Height int
	// optional FourCC code of the raw Xvid stream
	FourCC int
	// optional number of threads to use for decoding, 0 meaning single-threaded
	NumThreads int
}

// DecoderFrame is information used when decoding a frame in Decoder.Decode.
type DecoderFrame struct {
	// output image to store the decoded data to
	Output *Image
	// optional decoder flags to use for decoding the frame
	DecodeFlags DecoderFlag
	// optional brightness offset, 0 meaning no offset
	Brightness int
}

// DecoderStats is information about a decoded frame, returned by Decoder.Decode.
// If the frame is a metadata pseudo-frame (VOL), StatsVOL is not nil, otherwise
// StatsFrame is not nil.
type DecoderStats struct {
	// type of the decoded frame
	FrameType FrameType
	// non-nil if the frame type is FrameTypeVOL
	StatsVOL *DecoderStatsVOL
	// non-nil if the frame type is not FrameTypeVOL
	StatsFrame *DecoderStatsFrame
}

var decoderStatsNothing = DecoderStats{FrameType: frameTypeNothing}

// DecoderStatsVOL is information specific to a metadata pseudo-frame, returned by Decoder.Decode in DecoderStats.
type DecoderStatsVOL struct {
	// whether the frame is interlaced
	Interlacing bool
	// frame width in pixels
	Width int
	// frame height in pixels
	Height int
	// frame pixel aspect ratio
	PixelAspectRatio PixelAspectRatio
}

// DecoderStatsFrame is information specific to an actual non-metadata non-empty frame, returned by Decoder.Decode in DecoderStats.
type DecoderStatsFrame struct {
	// valid only for interlaced frames (see DecoderStatsVOL.Interlacing), whether the interlacing is upper field first
	UpperFieldFirst bool
	// macroblock quantizers table (one quantizer per macroblock), can be nil
	Quantizers []int32
	// quantizers table stride (equal to the count of macroblocks in a line)
	QuantizersStride int

	// TimeBase and TimeImplement are currently unimplemented in libxvidcore
	// TimeIncrement is useless without access to vop_time_increment_resolution
	// TimeBase int
	// TimeIncrement int
}

// NewDecoder creates a new Decoder based on a DecoderInit configuration. Init (or InitWithFlags) must be called once before calling this function.
// Once created and finished using, a Decoder must be freed by calling Decoder.Close().
// The Decoder is non-nil if and only if the returned error is nil.
// An internal error can be returned by Xvid, in which case the Decoder won't be created.
func NewDecoder(init DecoderInit) (*Decoder, error) {
	cDecoreCreate := C.xvid_dec_create_t{
		version:     C.XVID_VERSION,
		width:       C.int(init.Width),
		height:      C.int(init.Height),
		fourcc:      C.int(init.FourCC),
		num_threads: C.int(init.NumThreads),
	}
	if code := C.xvid_decore(nil, C.XVID_DEC_CREATE, unsafe.Pointer(&cDecoreCreate), nil); code != 0 {
		return nil, xvidErr(code)
	}
	var buf []byte
	if init.Input != nil {
		buf = make([]byte, 4*1024*1024) // highly unlikely that any frame will be larger than 2MB
	}
	return &Decoder{
		handle: cDecoreCreate.handle,
		Width:  init.Width,
		Height: init.Height,
		r:      init.Input,
		buf:    buf,
		i:      -1,
	}, nil
}

// Decode decodes a single non-empty frame (either metadata (VOL) or an actual frame) from the encoded Xvid stream.
//
// Decode returns an int, which is the length in bytes of the frame that was read. Decode might buffer up data from
// its Reader so the returned value might be less than the actual data read.
//
// Decode returns a DecoderStats, which stores information about the decoded frame,
// which can be either a VOL (metadata), or an actual frame. If Decoder returns a non-nil error,
// DecoderStats in invalid.
//
// Decode returns an error, which if it is not nil can be either io.EOF or another error.
// If it is io.EOF, this is an expected value which means that the entire stream has been decoded.
// Otherwise, it is an unexpected value, which can be due to invalid images, reader i/o errors, or internal
// Xvid errors.
//
// In any case, the Decoder should not be used after any error and Decode will always return
// the same error after an error occurs. The Decoder must still be closed with Close.
func (d *Decoder) Decode(frame DecoderFrame) (int, DecoderStats, error) {
	if d.r == nil {
		return 0, decoderStatsNothing, errors.New("xvid: Input Reader is nil, must be passed in Init")
	}

	if d.err != nil {
		return 0, decoderStatsNothing, d.err
	}

	if d.i == -1 { // initial read burst
		d.i = 0
		r, err := io.ReadFull(d.r, d.buf[d.n:])
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				d.eof = true
			} else {
				d.err = io.EOF
				return 0, decoderStatsNothing, d.err
			}
		}
		d.n += r
	}

	total := 0
	for { // read at least one non-nothing frame
		if d.eof && d.n-d.i <= 1 { // no bytes remaining: flush decoder
			r, stats, err := d.decodeBuffer(frame, nil)
			d.i += r
			total += r
			if err != nil {
				if e, ok := err.(*Error); ok && e.code == C.XVID_ERR_END {
					d.err = io.EOF
				} else {
					d.err = err
				}
				return 0, decoderStatsNothing, d.err
			}
			if stats.FrameType == frameTypeNothing {
				continue
			}
			return total, stats, nil
		}

		if !d.eof && d.i > len(d.buf)/2 {
			copy(d.buf[:d.n-d.i], d.buf[d.i:d.n])
			d.n = d.n - d.i
			d.i = 0
			r, err := io.ReadFull(d.r, d.buf[d.n:])
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					d.eof = true
				} else {
					d.err = err
					return 0, decoderStatsNothing, d.err
				}
			}
			d.n += r
		}
		r, stats, err := d.decodeBuffer(frame, d.buf[d.i:d.n])
		if err != nil {
			d.err = err
			return 0, decoderStatsNothing, d.err
		}
		if d.i+r > d.n { // read past buffer limit... abort decoding
			d.err = errors.New("xvid: read past buffer limit, should not happen")
			return 0, decoderStatsNothing, d.err
		}
		d.i += r
		total += r
		if stats.FrameType != frameTypeNothing {
			return total, stats, nil
		}
	}
}

// TODO make this public if someone needs this (with better documentation)
// decodes one (possibly empty) frame from the input buffer
// this low-level method should not be used directly, use Decode instead to automatically handle data buffering
// if you need to use this method check the Decode method source code to see how to use it
// no error and int=0 means the decoder needs more data
// at the end of the stream call with input=nil to flush decoder
// due to implementation quirks the buffer length will be reduced to the nearest length multiple of 8 below the buffer length
// due to implementation quirks the decoder might read more data past the buffer end if the buffer is small and only contains part of a frame
func (d *Decoder) decodeBuffer(frame DecoderFrame, input []byte) (int, DecoderStats, error) {
	l := -1
	var bitstream unsafe.Pointer = nil
	if input != nil {
		l = len(input)
		l = l - l%8
		if l == 0 {
			return 0, DecoderStats{FrameType: frameTypeNothing}, nil
		}
		bitstream = unsafe.Pointer(&input[0])
	}
	cOutput, err := frame.Output.nativeOutput(d.Width, d.Height)
	if err != nil {
		return 0, DecoderStats{FrameType: frameTypeNothing}, err
	}
	cDecoreFrame := C.xvid_dec_frame_t{
		version:    C.XVID_VERSION,
		general:    C.int(frame.DecodeFlags),
		bitstream:  bitstream,
		length:     C.int(l),
		output:     *cOutput,
		brightness: C.int(frame.Brightness),
	}
	cDecodeStats := C.xvid_dec_stats_t{
		version: C.XVID_VERSION,
	}
	code := C.xvid_decore(d.handle, C.XVID_DEC_DECODE, unsafe.Pointer(&cDecoreFrame), unsafe.Pointer(&cDecodeStats))
	if code < 0 {
		return 0, DecoderStats{FrameType: frameTypeNothing}, xvidErr(code)
	}
	if frame.Output.Colorspace.value == ColorSpaceInternal.value {
		j := 0
		for j < ColorSpaceInternal.Planes {
			l := d.Width * d.Height * frame.Output.Colorspace.BitsPerPixelPlanes[j] / 8
			sh := reflect.SliceHeader{
				Data: uintptr(cDecoreFrame.output.plane[j]),
				Len:  l,
				Cap:  l,
			}
			frame.Output.Planes[j] = *(*[]byte)(unsafe.Pointer(&sh))
			frame.Output.Strides[j] = int(cDecoreFrame.output.stride[j])
		}
	}
	stats := DecoderStats{
		FrameType: FrameType(cDecodeStats._type),
	}
	if stats.FrameType > 0 {
		cVopData := C.vop_data(&cDecodeStats)
		var quantizers []int32
		if cVopData.qscale != nil {
			mbWidth := (d.Width + 15) / 16
			mbHeight := (d.Height + 15) / 16
			if mbWidth != int(cVopData.qscale_stride) {
				// macroblock size computation mismatch, should not happen
				// dont return any quantizers
				// TODO: print to stderr?
			} else {
				n := mbWidth * mbHeight
				quantizers = make([]int32, n)
				sh := reflect.SliceHeader{
					Data: uintptr(unsafe.Pointer(cVopData.qscale)),
					Len:  n,
					Cap:  n,
				}
				if C.sizeof_int == unsafe.Sizeof(int32(0)) {
					copy(quantizers, *(*[]int32)(unsafe.Pointer(&sh)))
				} else {
					for i := range quantizers {
						quantizers[i] = int32((*(*[]C.int)(unsafe.Pointer(&sh)))[i])
					}
				}
			}
		}
		stats.StatsFrame = &DecoderStatsFrame{
			UpperFieldFirst:  cVopData.general&C.XVID_VOP_TOPFIELDFIRST != 0,
			Quantizers:       quantizers,
			QuantizersStride: int(cVopData.qscale_stride),
		}
	} else if stats.FrameType == FrameTypeVOL {
		cVolData := C.vol_data(&cDecodeStats)
		var par PixelAspectRatio
		switch cVolData.par {
		case C.XVID_PAR_11_VGA:
			par = PixelAspectRatio11VGA
		case C.XVID_PAR_43_PAL:
			par = PixelAspectRatio43PAL
		case C.XVID_PAR_43_NTSC:
			par = PixelAspectRatio43NTSC
		case C.XVID_PAR_169_PAL:
			par = PixelAspectRatio169PAL
		case C.XVID_PAR_169_NTSC:
			par = PixelAspectRatio169NTSC
		case C.XVID_PAR_EXT:
			par = newPixelAspectRatio(int(cVolData.par_width), int(cVolData.par_height))
		default:
			par = PixelAspectRatio11VGA
		}
		stats.StatsVOL = &DecoderStatsVOL{
			Interlacing:      cVolData.general&C.XVID_VOL_INTERLACING != 0,
			Width:            int(cVolData.width),
			Height:           int(cVolData.height),
			PixelAspectRatio: par,
		}
		d.Width = stats.StatsVOL.Width
		d.Height = stats.StatsVOL.Height
	}
	return int(code), stats, nil
}

// Close closes any internal resources specific to the Decoder.
// It must be called exactly once per Decoder and no other methods of the Decoder
// must be called after Close.
func (d *Decoder) Close() {
	C.xvid_decore(d.handle, C.XVID_DEC_DESTROY, nil, nil)
}

// Plugin is an Xvid plugin that is used during the encoding process as a callback
// for both read and write operations to some internal frame encoding data.
// Plugins are used in the Encoder methods.
//
// Some default plugins are provided, which include 1-pass and 2-pass rate control.
//
// Custom plugins can also be created by implementing the Plugin interface.
type Plugin interface {
	// called during Encoder.Init
	// return the information the plugins needs
	Info() PluginFlag
	// called during Encoder.Init to init the plugin (any startup setup should be done here)
	// return false to disable the plugin for an Encoder
	Init(create PluginInit) bool
	// called during Encoder.Close to close any open plugin resources
	Close(close PluginClose)

	// frame-related callbacks
	// check FrameData for what fields are readable/writable depending on the call

	// called during Encoder.Encode, before starting to encode a frame
	Before(data *PluginData)
	// called during Encoder.Encode, while encoding a frame
	Frame(data *PluginData)
	// called during Encoder.Encode, after encoding a frame
	After(data *PluginData)
}

type pluginInternal struct {
	cPlugin     C.xvid_enc_plugin_t
	free        func()
	destroyFree func()
}

func (p pluginInternal) Info() PluginFlag            { return 0 }
func (p pluginInternal) Init(create PluginInit) bool { return true }
func (p pluginInternal) Close(close PluginClose)     {}
func (p pluginInternal) Before(data *PluginData)     {}
func (p pluginInternal) Frame(data *PluginData)      {}
func (p pluginInternal) After(data *PluginData)      {}

// PluginRC1PassInit is a configuration for the PluginRC1Pass plugin (1-pass rate-control).
// To return a configuration initialized to default values, use NewPluginRC1PassInit.
type PluginRC1PassInit struct {
	// target bitrate in bits per second
	Bitrate int
	// reaction delay factor; defaults to 16
	ReactionDelayFactor int
	// averaging period; defaults to 100
	AveragingPeriod int
	// smoothing buffer; defaults to 100
	SmoothingBuffer int
}

// NewPluginRC1PassInit returns a PluginRC1PassInit initialized to default values.
func NewPluginRC1PassInit(bitrate int) PluginRC1PassInit {
	return PluginRC1PassInit{
		Bitrate:             bitrate,
		ReactionDelayFactor: 16,
		AveragingPeriod:     100,
		SmoothingBuffer:     100,
	}
}

// PluginRC1Pass returns an instance of the 1-pass rate-control plugin, to be used in NewEncoder.
// This plugin will choose specific quantizers to try to match the bitrate parameters.
func PluginRC1Pass(init PluginRC1PassInit) Plugin {
	return pluginInternal{
		cPlugin: C.xvid_enc_plugin_t{
			_func: &C.xvid_plugin_single,
			param: unsafe.Pointer(&C.xvid_plugin_single_t{
				version:               C.XVID_VERSION,
				bitrate:               C.int(init.Bitrate),
				reaction_delay_factor: C.int(init.ReactionDelayFactor),
				averaging_period:      C.int(init.AveragingPeriod),
				buffer:                C.int(init.SmoothingBuffer),
			}),
		},
	}
}

// PluginRC2Pass1 returns an instance of the 2-pass rate-control plugin for the first pass.
// To do 2-pass rate-control in Xvid, encode the same images twice, in the first run using the
// PluginRC2Pass1 plugin, and in the second run using the PluginRC2Pass2 plugin.
//
// PluginRC2Pass1 takes a filename which is used to store the rate-control information (the file will be overwritten).
// If the file writing fails, Xvid will not return errors, so you can check for the file existence yourself
// after the encoding ends.
func PluginRC2Pass1(filename string) Plugin {
	cFilename := C.CString(filename)
	return pluginInternal{
		cPlugin: C.xvid_enc_plugin_t{
			_func: &C.xvid_plugin_2pass1,
			param: unsafe.Pointer(&C.xvid_plugin_2pass1_t{
				version:  C.XVID_VERSION,
				filename: cFilename,
			}),
		},
		free: func() {
			C.free(unsafe.Pointer(cFilename))
		},
	}
}

// PluginRC2Pass2Init is a configuration for the PluginRC2Pass2 plugin (2-pass rate-control, pass 2).
// To return a configuration initialized to default values, use NewPluginRC2Pass2Init.
type PluginRC2Pass2Init struct {
	// target bitrate in bits per second; defaults to 700*1024
	Bitrate int
	// path to file to read rate-control info from, should be the same file as the first pass filename
	Filename string
	// I-frame boost percentage, range: [0..100]; defaults to 10
	IFrameBoost int
	// percentage of compression performed on the high part of the curve (above average); defaults to 0
	CurveCompressionHigh int
	// percentage of compression performed on the low  part of the curve (below average); defaults to 0
	CurveCompressionLow int
	// payback delay in number of frames; defaults to 5
	OverflowControlStrength int
	// percentage of allowed range for a frame that gets bigger because of overflow bonus; defaults to 5
	MaxOverflowImprovement int
	// percentage of allowed range for a frame that gets smaller because of overflow penalty; defaults to 5
	MaxOverflowDegradation int

	// maximum bitrate reduction applied to an I-frame under the kfthreshold distance limit; defaults to 20
	IFrameReduction int

	// if an iframe is closer to the next iframe than this distance, a quantity of bits
	// is substracted from its bit allocation. The reduction is computed as multiples of
	// kfreduction/kthreshold. It reaches kfreduction when the distance == kfthreshold,
	// 0 for 1<distance<kfthreshold
	// defaults to 1
	IFrameThreshold int

	// how many bytes the controller has to compensate per frame due to container format overhead; defaults to 0
	ContainerFrameOverhead int

	// Video Buffering Verifier buffer size in bits; 0 disables VBV check; defaults to 0
	VBVSize int
	// Video Buffering Verifier max processing bitrate in bits per second
	VBVMaxRate int
	// Video Buffering Verifier initial buffer occupancy in bits; defaults to 0
	VBVInitial int
}

// NewPluginRC2Pass2Init returns a PluginRC2Pass2Init initialized to default values.
// It takes a filename which is used to read the rate-control information from a previous first pass.
func NewPluginRC2Pass2Init(bitrate int, filename string) PluginRC2Pass2Init {
	return PluginRC2Pass2Init{
		Bitrate:                 bitrate,
		Filename:                filename,
		IFrameBoost:             10,
		OverflowControlStrength: 5,
		MaxOverflowImprovement:  5,
		MaxOverflowDegradation:  5,
		IFrameReduction:         20,
		IFrameThreshold:         1,
		ContainerFrameOverhead:  100,
	}
}

// PluginRC2Pass2 returns an instance of the 2-pass rate-control plugin for the second pass.
// To do 2-pass rate-control in Xvid, encode the same images twice, in the first run using the
// PluginRC2Pass1 plugin, and in the second run using the PluginRC2Pass2 plugin.
func PluginRC2Pass2(init PluginRC2Pass2Init) Plugin {
	filename := C.CString(init.Filename)
	return pluginInternal{
		cPlugin: C.xvid_enc_plugin_t{
			_func: &C.xvid_plugin_2pass2,
			param: unsafe.Pointer(&C.xvid_plugin_2pass2_t{
				version:                   C.XVID_VERSION,
				bitrate:                   C.int(init.Bitrate),
				filename:                  filename,
				keyframe_boost:            C.int(init.IFrameBoost),
				curve_compression_high:    C.int(init.CurveCompressionHigh),
				curve_compression_low:     C.int(init.CurveCompressionLow),
				overflow_control_strength: C.int(init.OverflowControlStrength),
				max_overflow_improvement:  C.int(init.MaxOverflowImprovement),
				max_overflow_degradation:  C.int(init.MaxOverflowDegradation),
				kfreduction:               C.int(init.IFrameReduction),
				kfthreshold:               C.int(init.IFrameThreshold),
				container_frame_overhead:  C.int(init.ContainerFrameOverhead),
				vbv_size:                  C.int(init.VBVSize),
				vbv_initial:               C.int(init.VBVInitial),
				vbv_maxrate:               C.int(init.VBVMaxRate),
			}),
		},
		free: func() {
			C.free(unsafe.Pointer(filename))
		},
	}
}

// MaskingMethod is a method used for lumi-masking (adaptive quantization).
type MaskingMethod uint

const (
	// luminance masking
	MaskingLuminance MaskingMethod = 0
	// variance masking
	MaskingVariance MaskingMethod = 1
)

// PluginAdaptiveQuantization returns an instance of the adaptive quantization plugin
// (also-called lumi-masking).
func PluginAdaptiveQuantization(method MaskingMethod) Plugin {
	return pluginInternal{
		cPlugin: C.xvid_enc_plugin_t{
			_func: &C.xvid_plugin_lumimasking,
			param: unsafe.Pointer(&C.xvid_plugin_lumimasking_t{
				version: C.XVID_VERSION,
				method:  C.int(method),
			}),
		},
	}
}

// PluginPSNR returns an instance of a plugin that writes PSNR values to the standard output.
func PluginPSNR() Plugin {
	return pluginInternal{
		cPlugin: C.xvid_enc_plugin_t{
			_func: &C.xvid_plugin_psnr,
			param: nil,
		},
	}
}

// PluginDump returns an instance of a plugin that writes original and encoded image data
// to files in YUV in PGM format in the working directory.
func PluginDump() Plugin {
	return pluginInternal{
		cPlugin: C.xvid_enc_plugin_t{
			_func: &C.xvid_plugin_dump,
			param: nil,
		},
	}
}

// PluginSSIMInit is a configuration for the PluginSSIM plugin (write SSIM values).
// The SSIM values can be written to the standard output or to a file.
type PluginSSIMInit struct {
	// whether to output stats to stdout
	PrintStats bool
	// output stats filename (file will be overwritten); or empty to not output to a file
	StatsFilename string

	// SSIM computation accuracy from 0 (gaussian weighted, very slow), 1 (unweighted, slow) to 4 (unweighted, very fast); default is 2
	Accuracy int
	// CPU flags to use for decoding, or nil to use the autodetected CPU features
	CpuFlags *CPUFlag

	// visualization is currently disabled in xvidcore
	// Visualize  bool
}

// PluginSSIM returns an instance of a plugin that writes SSIM values to the standard output
// or to a file.
func PluginSSIM(init PluginSSIMInit) Plugin {
	var filename *C.char = nil
	if init.StatsFilename != "" {
		filename = C.CString(init.StatsFilename)
	}
	var cpuFlags C.int = 0
	if init.CpuFlags != nil {
		cpuFlags = C.int(*init.CpuFlags | CPUFlag(C.CPU_FORCE))
	}
	return pluginInternal{
		cPlugin: C.xvid_enc_plugin_t{
			_func: &C.xvid_plugin_ssim,
			param: unsafe.Pointer(&C.xvid_plugin_ssim_t{
				b_printstat: cbool(init.PrintStats),
				stat_path:   filename,
				acc:         C.int(init.Accuracy),
				cpu_flags:   cpuFlags,
			}),
		},
		destroyFree: func() {
			if filename != nil {
				C.free(unsafe.Pointer(filename))
			}
		},
	}
}

// PluginPSNRHVSM returns an instance of a plugin that writes PSNRHVSM values
// to the standard output.
func PluginPSNRHVSM() Plugin {
	return pluginInternal{
		cPlugin: C.xvid_enc_plugin_t{
			_func: &C.xvid_plugin_psnrhvsm,
			param: nil,
		},
	}
}

// PluginInit stores general information for an encoder, used for reading by plugins
// in their Init callback.
type PluginInit struct {
	// encoder bitrate zones, that enforce a specific parameter for a range of frames; must be sorted in increasing frame start order
	Zones []EncoderZone
	// frame width in pixels
	Width int
	// frame height in pixels
	Height int
	// frame width in macro blocks
	WidthMacroBlocks int
	// frame height in macro blocks
	HeightMacroBlocks int
	// framerate; Numerator=0 means variable framerate
	FrameRate Fraction
}

// PluginClose stores information for an encoding session, used for reading by plugins
// in their Close callback.
type PluginClose struct {
	// total count of encoded frames
	NumFrames int
}

// PluginData stores information about an encoder and a specific frame to encode.
// It is used in plugins Before, Frame, and After callbacks.
//
// Depending on the callback ([B]efore, [F]rame, [A]fter), some fields can be
// [R]eadable or [W]ritable. To represent this, each field description starts
// with a list [<B(efore)/F(rame)/A(fter)><R(ead)/W(rite)>, ...].
// For example [AR,FW] means: writable during Frame, readable during After.
type PluginData struct {
	// [BR,FR,AR] current encoder zone, or nil if none
	Zone *EncoderZone
	// [BR,FR,AR] frame width in pixels
	Width int
	// [BR,FR,AR] frame height in pixels
	Height int
	// [BR,FR,AR] frame width in macro blocks
	WidthMacroBlocks int
	// [BR,FR,AR] frame height in macro blocks
	HeightMacroBlocks int
	// [BR,FR,AR] framerate; Numerator=0 means variable framerate
	FrameRate Fraction
	// [BR,FR,AR] quantizer range for I frames
	QuantizerI QuantizerRange
	// [BR,FR,AR] quantizer range for P frames
	QuantizerP QuantizerRange
	// [BR,FR,AR] quantizer range for B frames
	QuantizerB QuantizerRange
	// [BR,FR,AR] frame number
	FrameNum int
	// [BR,FR,AR] reference frame
	Reference Image
	// [BR,FR,AR] current frame
	Current Image
	// [AR] the original (uncompressed) copy of the current frame
	Original Image
	// [BR,FR,AR,BW] type of this frame
	Type FrameType
	// [BR,FR,AR,BW,FW] quantizer used for this frame
	Quantizer int
	// [AR,FW] diff quantizers for this frame, only present if PluginRequireDiffQuantizer was set during Info()
	DiffQuantizers []int
	// [FR,AR] diff quantizers stride (quantizers per row), only set if PluginRequireDiffQuantizer was set during Info()
	DiffQuantizersStride int
	// [BR,AR,BW] actual group of pictures flags
	VOLFlags VOLFlag
	// [BR,AR,BW] encoding flags for this frame
	VOPFlags VOPFlag
	// [BR,AR,BW] motion estimation flags for this frame
	MotionFlags MotionFlag
	// [FW] lambda table for this frame, only present if PluginRequireLambda was set during Info(); six floats for each macroblock
	Lambda []float32
	// [BR,FR,AR] B-frames quantizer multipier/offset; used to decide B-frames quantizer when automatic quantizer is used
	BFrameQuantizer BFrameQuantizer
	// [AR] frame statistics
	Stats EncoderStats
}

var encoderMutex = sync.Mutex{} // TODO use global map and int to avoid C referencing go memory
var encoder *Encoder

func internalImage(cImage C.xvid_image_t, width int, height int) (*Image, error) {
	if int(cImage.csp) != ColorSpacePlanar.value {
		return nil, fmt.Errorf("xvid: unexpected encoder internal image colorspace %d", int(cImage.csp))
	}
	colorspace := ColorSpacePlanar
	image := Image{
		Colorspace:   colorspace,
		VerticalFlip: false,
		Planes:       make([][]byte, colorspace.Planes),
		Strides:      make([]int, colorspace.Planes),
	}
	for i := range image.Planes {
		l := width * height * colorspace.BitsPerPixelPlanes[i] / 8
		sh := reflect.SliceHeader{
			Data: uintptr(cImage.plane[i]),
			Len:  l,
			Cap:  l,
		}
		image.Planes[i] = *(*[]byte)(unsafe.Pointer(&sh))
		image.Strides[i] = int(cImage.stride[i])
	}
	return &image, nil
}

//export pluginCallback
func pluginCallback(handle unsafe.Pointer, option int, param1 unsafe.Pointer, param2 unsafe.Pointer) int {
	switch option {
	case C.XVID_PLG_INFO:
		cInfo := (*C.xvid_plg_info_t)(param1)
		for {
			if _, ok := encoder.plugins[encoder.currentPlugin].(pluginInternal); ok {
				encoder.currentPlugin++
				continue
			}
			break
		}
		cInfo.flags = C.int(encoder.plugins[encoder.currentPlugin].Info())
		encoder.currentPlugin++
		return 0
	case C.XVID_PLG_DESTROY:
		cDestroy := (*C.xvid_plg_destroy_t)(param1)
		if cDestroy == nil {
			// can happen if oom during encoding init, ignore
			return 0
		}
		plugin := *(*Plugin)(handle)
		plugin.Close(PluginClose{
			NumFrames: int(cDestroy.num_frames),
		})
		return 0
	case C.XVID_PLG_CREATE:
		cCreate := (*C.xvid_plg_create_t)(param1)
		pluginInit := PluginInit{
			Zones:             encoder.zones,
			Width:             int(cCreate.width),
			Height:            int(cCreate.height),
			WidthMacroBlocks:  int(cCreate.mb_width),
			HeightMacroBlocks: int(cCreate.mb_height),
			FrameRate:         Fraction{int(cCreate.fbase), int(cCreate.fincr)},
		}
		plugin := (*Plugin)(cCreate.param)
		*(**Plugin)(param2) = plugin
		if !(*plugin).Init(pluginInit) {
			return -1
		}
		return 0
	case C.XVID_PLG_BEFORE:
		cData := (*C.xvid_plg_data_t)(param1)
		plugin := *(*Plugin)(handle)
		if data := pluginReadData(cData); data != nil {
			plugin.Before(data)
			pluginWriteData(cData, data)
		}
		return 0
	case C.XVID_PLG_FRAME:
		cData := (*C.xvid_plg_data_t)(param1)
		plugin := *(*Plugin)(handle)
		if data := pluginReadData(cData); data != nil {
			plugin.Frame(data)
			pluginWriteData(cData, data)
		}
		return 0
	case C.XVID_PLG_AFTER:
		cData := (*C.xvid_plg_data_t)(param1)
		plugin := *(*Plugin)(handle)
		if data := pluginReadData(cData); data != nil {
			plugin.After(data)
			pluginWriteData(cData, data)
		}
		return 0
	}
	// should not happen, ignore
	return 0
}

func pluginReadData(cData *C.xvid_plg_data_t) *PluginData {
	var zone *EncoderZone = nil
	if cData.zone != nil {
		zone = &EncoderZone{
			Frame: int(cData.zone.frame),
			Mode:  ZoneType(cData.zone.mode),
			Value: Fraction{int(cData.zone.increment), int(cData.zone.base)},
		}
	}
	pluginData := PluginData{
		Zone:                 zone,
		Width:                int(cData.width),
		Height:               int(cData.height),
		WidthMacroBlocks:     int(cData.mb_width),
		HeightMacroBlocks:    int(cData.mb_height),
		FrameRate:            Fraction{int(cData.fbase), int(cData.fincr)},
		QuantizerI:           QuantizerRange{int(cData.min_quant[0]), int(cData.max_quant[0])},
		QuantizerP:           QuantizerRange{int(cData.min_quant[1]), int(cData.max_quant[1])},
		QuantizerB:           QuantizerRange{int(cData.min_quant[2]), int(cData.max_quant[2])},
		FrameNum:             int(cData.frame_num),
		Type:                 FrameType(cData._type),
		Quantizer:            int(cData.quant),
		DiffQuantizersStride: int(cData.dquant_stride),
		VOLFlags:             VOLFlag(cData.vol_flags),
		VOPFlags:             VOPFlag(cData.vop_flags),
		MotionFlags:          MotionFlag(cData.motion_flags),
		BFrameQuantizer:      BFrameQuantizer{int(cData.bquant_ratio), int(cData.bquant_offset)},
		Stats: EncoderStats{
			FrameType:     FrameType(cData.stats._type),
			KeyFrame:      FrameType(cData.stats._type) == FrameTypeI,
			Quantizer:     int(cData.stats.quant),
			VOLFlags:      VOLFlag(cData.stats.vol_flags),
			VOPFlags:      VOPFlag(cData.stats.vop_flags),
			Length:        int(cData.stats.length),
			HeaderLength:  int(cData.stats.hlength),
			IntraBlocks:   int(cData.stats.kblks),
			InterBlocks:   int(cData.stats.mblks),
			UncodedBlocks: int(cData.stats.ublks),
			SSEY:          int(cData.stats.sse_y),
			SSEU:          int(cData.stats.sse_u),
			SSEV:          int(cData.stats.sse_v),
		},
	}
	referenceImage, err := internalImage(cData.reference, pluginData.Width, pluginData.Height)
	if err != nil {
		// TODO print to stderr?
		return nil
	}
	pluginData.Reference = *referenceImage
	currentImage, err := internalImage(cData.current, pluginData.Width, pluginData.Height)
	if err != nil {
		// TODO print to stderr?
		return nil
	}
	pluginData.Current = *currentImage
	if cData.original.csp != 0 {
		originalImage, err := internalImage(cData.original, pluginData.Width, pluginData.Height)
		if err != nil {
			// TODO print to stderr?
			return nil
		}
		pluginData.Original = *originalImage
	}
	if cData.dquant != nil {
		l := pluginData.WidthMacroBlocks * pluginData.HeightMacroBlocks
		if C.sizeof_int == unsafe.Sizeof(int(0)) { // avoid expensive copy if C.int == int
			pluginData.DiffQuantizers = *(*[]int)(unsafe.Pointer(&reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(cData.dquant)),
				Len:  l,
				Cap:  l,
			}))
		} else {
			pluginData.DiffQuantizers = make([]int, l)
			cDiffQuantizers := *(*[]C.int)(unsafe.Pointer(&reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(cData.dquant)),
				Len:  l,
				Cap:  l,
			}))
			for i, v := range cDiffQuantizers {
				pluginData.DiffQuantizers[i] = int(v)
			}
		}
	}
	if cData.lambda != nil {
		n := 6 * pluginData.WidthMacroBlocks * pluginData.HeightMacroBlocks
		pluginData.Lambda = *(*[]float32)(unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(cData.lambda)),
			Len:  n,
			Cap:  n,
		}))
	}
	return &pluginData
}

func pluginWriteData(cData *C.xvid_plg_data_t, pluginData *PluginData) {
	cData._type = C.int(pluginData.Type)
	cData.quant = C.int(pluginData.Quantizer)
	if pluginData.DiffQuantizers != nil {
		if C.sizeof_int != unsafe.Sizeof(int(0)) { // only copy back if we had copied before
			cDiffQuantizers := *(*[]C.int)(unsafe.Pointer(&reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(cData.dquant)),
				Len:  len(pluginData.DiffQuantizers),
				Cap:  len(pluginData.DiffQuantizers),
			}))
			for i, v := range pluginData.DiffQuantizers {
				cDiffQuantizers[i] = C.int(v)
			}
		}
	}
	cData.vol_flags = C.int(pluginData.VOLFlags)
	cData.vop_flags = C.int(pluginData.VOPFlags)
	cData.motion_flags = C.int(pluginData.MotionFlags)
}

// BufferSize returns the minimal output buffer size for encoding a frame.
// The Output buffer of an EncoderFrame will automatically be reallocated to
// this size if it is smaller.
func BufferSize(width int, height int) int {
	// inspired by FFMPEG libxvid encoder
	return 16384 + width*height*30*3/8 + 120 + 8
}

// BFrameQuantizer stores parameters for choosing B-frames quantizers.
// The actual formula used is:
//   quantizer = (average(pastReferenceQuantizer, futureReferenceQuantizer) * Ratio + Offset) / 100
type BFrameQuantizer struct {
	// ratio in percent (see formula); default is 150
	Ratio int
	// offset in 0.01 units (see formula); default is 100
	Offset int
}

// Encoder is an initialized Xvid encoder.
// To create a Encoder, use NewEncoder.
// An Encoder must be closed after use, by calling its Close method.
// To encode a frame, use the Encode method.
type Encoder struct {
	handle        unsafe.Pointer
	width         int
	height        int
	zones         []EncoderZone
	plugins       []Plugin
	currentPlugin int
	closed        bool
	err           error
}

// EncoderInit is information used to create an Encoder in NewEncoder.
// To initialize an EncoderInit with default values, use NewEncoderInit.
type EncoderInit struct {
	// frame width in pixels
	Width int
	// frame height in pixels
	Height int

	// optional encoder profile; default is EncoderAuto for automatic profile selection
	Profile EncoderProfile
	// optional encoder bitrate zones, that enforce a specific parameter for a range of frames; must be sorted in increasing frame start order
	Zones []EncoderZone
	// optional encoder plugins
	Plugins []Plugin
	// optional number of threads to use for encoding, 0 means single-threaded; default is GetGlobalInfo().NumThreads-1
	NumThreads int
	// optional maximum sequential B-frames, 0 means disabling B-frames; default is 2
	MaxBFrames int
	// optional global encoder flags; default is no flags
	Flags EncoderFlag

	// framerate; Numerator=0 means variable framerate; only the Denominator can be changed after initialization
	FrameRate Fraction

	// optional maximum interval between key frames; default is 300
	MaxKeyFrameInterval int
	// optional frame dropping ratio in percent between 0 (drop none) to 100 (drop all); default is 0
	FrameDropRatio int

	// optional B-frames quantizer multipier/offset; used to decide B-frames quantizer when automatic quantizer is used
	BFrameQuantizer BFrameQuantizer

	// optional quantizer range for I frames
	QuantizerI QuantizerRange
	// optional quantizer range for P frames
	QuantizerP QuantizerRange
	// optional quantizer range for B frames
	QuantizerB QuantizerRange

	// optional starting frame number, relative to the zones start frames; default is 0
	StartFrameNumber int
	// optional number of slices to encode for each frame; default is 0, meaning 1 slice
	NumSlices int
}

// EncoderZone is a bitrate enforcement zone used for encoding, which applies during
// a range of frames, starting on its Frame (inclusive) and ending on the next EncoderZone frame
// (exclusive).
type EncoderZone struct {
	// start frame (inclusive) of the zone
	Frame int
	// zone type
	Mode ZoneType
	// value, meaning depends on the ZoneType used
	Value Fraction
}

// EncoderFrame is information used when encoding a frame in Encoder.Encode.
// Its only required fields are the Input Image and its Output buffer.
type EncoderFrame struct {
	// input image to encode
	Input *Image
	// buffer to store the encoded frame data into, if pointing to a nil or too small slice, will realloc it to the minimum buffer size as returned by BufferSize
	Output *[]byte

	// optional flags for the next group of pictures; the encoder will not react to any changes until the next VOL (keyframe)
	VOLFlags VOLFlag
	// optional 8x8 row-major quantizer matrix for intraframe encoding
	QuantizerIntraMatrix []uint8
	// optional 8x8 row-major quantizer matrix for interframe encoding
	QuantizerInterMatrix []uint8
	// optional pixel aspect ratio, defaults to square pixel
	PixelAspectRatio PixelAspectRatio

	// optional; sets the frame rate by changing the Denominator of the frame rate fraction defined in Init; default means unchanged frame rate
	FrameRateDenominator int
	// optional encoding flags for this frame
	VOPFlags VOPFlag
	// optional motion estimation flags for this frame
	MotionFlags MotionFlag

	// optional forced type for this frame, defaults to FrameTypeAuto
	Type FrameType
	// optional quantizer for this frame, 0 defaults to automatic rate-controlled quantizer, recommended range is 2-31
	Quantizer int
	// optional adjustment for choosing between encoding a P-frame or a B-frame; > 0 means more B-frames, <0 means less B-frames
	BFrameThreshold int
}

// EncoderStats is information about an encoded frame, returned by Encoder.Encode.
type EncoderStats struct {
	// frame type of the encoded frame
	FrameType FrameType
	// whether this frame was encoded as an I frame
	KeyFrame bool
	// quantizer used for the frame
	Quantizer int
	// actual VOL flags used for the frame
	VOLFlags VOLFlag
	// actual VOP flags used for the frame
	VOPFlags VOPFlag
	// length of frame in bytes
	Length int
	// length of frame header in bytes
	HeaderLength int
	// number of blocks coded as intra
	IntraBlocks int
	// number of blocks coded as inter
	InterBlocks int
	// number of blocks not coded
	UncodedBlocks int

	// only present if VOLExtraStats is set; Y plane SSE
	SSEY int
	// only present if VOLExtraStats is set; U plane SSE
	SSEU int
	// only present if VOLExtraStats is set; V plane SSE
	SSEV int
}

// NewEncoderInit returns an EncoderInit initialized with the default encoding parameters.
//
// In Xvid rate-control is done with plugins: either 1-pass with PluginRC1Pass, or 2-pass
// with PluginRC2Pass1 (on the first pass) and PluginRC2Pass2 (on the second pass).
func NewEncoderInit(width int, height int, frameRate Fraction, plugins []Plugin) *EncoderInit {
	nThreads := 1
	if info, err := GetGlobalInfo(); err == nil && info.NumThreads > 2 {
		nThreads = info.NumThreads - 1
	}
	return &EncoderInit{
		Width:  width,
		Height: height,

		Profile:    EncoderProfileAuto,
		Zones:      nil,
		Plugins:    plugins,
		NumThreads: nThreads,
		MaxBFrames: 2,
		Flags:      0,

		FrameRate: Fraction{frameRate.Numerator, frameRate.Denominator},

		MaxKeyFrameInterval: 300,
		FrameDropRatio:      0,

		BFrameQuantizer: BFrameQuantizer{150, 100},

		QuantizerI: QuantizerRange{},
		QuantizerP: QuantizerRange{},
		QuantizerB: QuantizerRange{},

		StartFrameNumber: 0,
		NumSlices:        1,
	}
}

// NewEncoder creates a new Encoder based on a EncoderInit configuration. Init (or InitWithFlags) must be called once before calling this function.
// Once created and finished using, an Encoder must be freed by calling Encoder.Close().
// The Encoder is non-nil if and only if the returned error is nil.
// An internal error can be returned by Xvid, in which case the Encoder won't be created.
func NewEncoder(init *EncoderInit) (*Encoder, error) {
	if init == nil {
		return nil, errors.New("EncoderInit must not be nil")
	}
	e := Encoder{
		width:  init.Width,
		height: init.Height,
	}
	var cZonesPtr *C.xvid_enc_zone_t = nil
	if len(init.Zones) > 0 {
		cZones := make([]C.xvid_enc_zone_t, len(init.Zones))
		for i, z := range init.Zones {
			cZones[i] = C.xvid_enc_zone_t{
				frame:     C.int(z.Frame),
				mode:      C.int(z.Mode),
				increment: C.int(z.Value.Numerator),
				base:      C.int(z.Value.Denominator),
			}
		}
		cZonesPtr = &cZones[0]
	}
	var cPluginsPtr *C.xvid_enc_plugin_t = nil
	if len(init.Plugins) > 0 {
		cPlugins := make([]C.xvid_enc_plugin_t, len(init.Plugins))
		cPluginsPtr = &cPlugins[0]
		e.plugins = make([]Plugin, len(init.Plugins))
		copy(e.plugins, init.Plugins)
		for i, v := range init.Plugins {
			if pi, ok := v.(pluginInternal); ok {
				cPlugins[i] = pi.cPlugin
			} else {
				cPlugins[i] = C.xvid_enc_plugin_t{
					_func: (*C.xvid_plugin_func)(unsafe.Pointer(C.pluginCallback_cgo)),
					param: unsafe.Pointer(&e.plugins[i]),
				}
			}
		}
	}

	cEncoreCreate := C.xvid_enc_create_t{
		version:          C.XVID_VERSION,
		profile:          C.int(init.Profile),
		width:            C.int(init.Width),
		height:           C.int(init.Height),
		num_zones:        C.int(len(init.Zones)),
		zones:            cZonesPtr,
		num_plugins:      C.int(len(init.Plugins)),
		plugins:          cPluginsPtr,
		num_threads:      C.int(init.NumThreads),
		max_bframes:      C.int(init.MaxBFrames),
		global:           C.int(init.Flags),
		fincr:            C.int(init.FrameRate.Denominator),
		fbase:            C.int(init.FrameRate.Numerator),
		max_key_interval: C.int(init.MaxKeyFrameInterval),
		frame_drop_ratio: C.int(init.FrameDropRatio),
		bquant_ratio:     C.int(init.BFrameQuantizer.Ratio),
		bquant_offset:    C.int(init.BFrameQuantizer.Offset),
		min_quant:        [3]C.int{C.int(init.QuantizerI.Min), C.int(init.QuantizerB.Min), C.int(init.QuantizerP.Min)},
		max_quant:        [3]C.int{C.int(init.QuantizerI.Max), C.int(init.QuantizerB.Max), C.int(init.QuantizerP.Max)},
		start_frame_num:  C.int(init.StartFrameNumber),
		num_slices:       C.int(init.NumSlices),
	}
	encoderMutex.Lock()
	encoder = &e
	code := C.xvid_encore(nil, C.XVID_ENC_CREATE, unsafe.Pointer(&cEncoreCreate), nil)
	encoder = nil
	encoderMutex.Unlock()
	for _, v := range init.Plugins {
		if pi, ok := v.(pluginInternal); ok && pi.free != nil {
			pi.free()
		}
	}

	if code != 0 {
		return nil, xvidErr(code)
	}
	e.handle = cEncoreCreate.handle
	return &e, nil
}

// Encode encodes a single Image to an encoded Xvid stream.
//
// Encode returns an int, which is the length in bytes of the frame that was written.
// The Encoder might write data and return a non-zero int even if no frame was written
// (even if EncoderStats is nil).
//
// Encode returns an EncoderStats, which stores information about the encoded frame,
// which can be either a VOL (metadata), an actual frame, or nil. nil means that no
// frame was encoded, even though some data may be written (and int could be > 0) as
// Xvid can sometimes buffer frame data internally or write part of a frame to the stream.
//
// Encode returns an error, which if not nil can be due to invalid images, or internal
// Xvid errors.
//
// In most cases, the Encode should not be used after an error occurs. The Encode must
// still be closed with Close.
func (e *Encoder) Encode(frame EncoderFrame) (int, *EncoderStats, error) {
	if e.closed {
		return 0, nil, fmt.Errorf("xvid: encoder is closed")
	}
	var quantIntraMatrix *C.uchar = nil
	if frame.QuantizerIntraMatrix != nil {
		if len(frame.QuantizerIntraMatrix) != 64 {
			return 0, nil, fmt.Errorf("xvid: expected quantization intra table of 64 coefficients, got %d", len(frame.QuantizerIntraMatrix))
		}
		quantIntraMatrix = (*C.uchar)(unsafe.Pointer(&frame.QuantizerIntraMatrix[0]))
	}
	var quantInterMatrix *C.uchar = nil
	if frame.QuantizerInterMatrix != nil {
		if len(frame.QuantizerInterMatrix) != 64 {
			return 0, nil, fmt.Errorf("xvid: expected quantization inter table of 64 coefficients, got %d", len(frame.QuantizerIntraMatrix))
		}
		quantInterMatrix = (*C.uchar)(unsafe.Pointer(&frame.QuantizerInterMatrix[0]))
	}
	cInput, err := frame.Input.nativeInput(e.width, e.height)
	if err != nil {
		return 0, nil, err
	}
	if l := BufferSize(e.width, e.height); len(*frame.Output) < l {
		*frame.Output = make([]byte, l)
	}
	bitstream := unsafe.Pointer(&(*frame.Output)[0])
	cEncoreFrame := C.xvid_enc_frame_t{
		version:            C.XVID_VERSION,
		vol_flags:          C.int(frame.VOLFlags),
		quant_intra_matrix: quantIntraMatrix,
		quant_inter_matrix: quantInterMatrix,
		par:                C.int(frame.PixelAspectRatio.value),
		par_width:          C.int(frame.PixelAspectRatio.Width),
		par_height:         C.int(frame.PixelAspectRatio.Height),
		fincr:              C.int(frame.FrameRateDenominator),
		vop_flags:          C.int(frame.VOPFlags),
		motion:             C.int(frame.MotionFlags),
		input:              *cInput,
		_type:              C.int(frame.Type),
		quant:              C.int(frame.Quantizer),
		bframe_threshold:   C.int(frame.BFrameThreshold),
		bitstream:          bitstream,
		length:             C.int(len(*frame.Output)),
	}
	cEncodeStats := C.xvid_enc_stats_t{
		version: C.XVID_VERSION,
	}
	code := C.xvid_encore(e.handle, C.XVID_ENC_ENCODE, unsafe.Pointer(&cEncoreFrame), unsafe.Pointer(&cEncodeStats))
	if code < 0 {
		return 0, nil, xvidErr(code)
	}
	keyframe := cEncoreFrame.out_flags&C.XVID_KEYFRAME != 0
	var stats *EncoderStats = nil
	frameType := FrameType(cEncodeStats._type)
	if frameType != C.XVID_TYPE_NOTHING {
		stats = &EncoderStats{
			FrameType:     frameType,
			KeyFrame:      keyframe,
			Quantizer:     int(cEncodeStats.quant),
			VOLFlags:      VOLFlag(cEncodeStats.vol_flags),
			VOPFlags:      VOPFlag(cEncodeStats.vop_flags),
			Length:        int(cEncodeStats.length),
			HeaderLength:  int(cEncodeStats.hlength),
			IntraBlocks:   int(cEncodeStats.kblks),
			InterBlocks:   int(cEncodeStats.mblks),
			UncodedBlocks: int(cEncodeStats.ublks),
			SSEY:          int(cEncodeStats.sse_y),
			SSEU:          int(cEncodeStats.sse_u),
			SSEV:          int(cEncodeStats.sse_v),
		}
	}
	return int(code), stats, nil
}

// Close closes any internal resources specific to the Encoder.
// It must be called exactly once per Encoder and no other methods of the Encoder
// must be called after Close.
func (e *Encoder) Close() {
	if e.closed {
		return
	}
	e.closed = true
	C.xvid_encore(e.handle, C.XVID_ENC_DESTROY, nil, nil)
	for _, plugin := range e.plugins {
		if pi, ok := plugin.(pluginInternal); ok && pi.destroyFree != nil {
			pi.destroyFree()
		}
	}
}
