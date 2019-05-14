# go-xvid [![GoDoc](https://godoc.org/github.com/delthas/go-xvid?status.svg)](https://godoc.org/github.com/delthas/go-xvid) [![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)](https://github.com/emersion/stability-badges#experimental)

Go bindings for Xvid (libxvidcore) 1.3.X

Run with environment variable `GODEBUG=cgocheck=0`.

## Usage

The API is well-documented in its [![GoDoc](https://godoc.org/github.com/delthas/go-xvid?status.svg)](https://godoc.org/github.com/delthas/go-xvid)

A [decoder](https://github.com/delthas/go-xvid/tree/master/examples/decoder/main.go) and [encoder](https://github.com/delthas/go-xvid/tree/master/examples/encoder/main.go) example are available in `examples/` (must be run from the repo main directory, with `GODEBUG=cgocheck=0`).

You can also check the library source code and the [Xvid source code](https://labs.xvid.com/source/) (please open an issue if the library lacks documentation for your use case).

## Status

Some tests run locally, not used in production environments yet.

The API could be slightly changed in backwards-incompatible ways for now.

- [X] Nearly all of xvidcore
- [ ] Frame slice rendering

#### Known issues

- The BGRA, ABGR, RGBA, ARGB colorspaces have their alpha channel set to 0 (fully transparent) instead of 255 (fully opaque) when used as output of conversion/encoding

## License

*Disclaimer: IANAL/TINLA*

TL;DR go-xvid is MIT-licensed, but if you build and redistribute the binaries of a program that uses xvid through go-xvid you must redistribute it as GPLv2.

The go-xvid source code files themselves do not copy or use any significant part of libxvidcore. By themselves the source code files are MIT-licensed as stated in the LICENSE file and solely belong to the copyright owners listed in the LICENSE files.

A piece of software that uses (statically links to) the go-xvid bindings will probably link (dynamically) to libxvidcore. If that is the case, and the program is to be redistributed, then per the GPLv2 license, that piece of software must be redistributed under the GPLv2 license (which includes distributing the source code of the program). *(Actually under a license that is compatible with GPLv2, but there are almost none.)*

Note that if you build a program that links against libxvidcore but you do not redistribute it (typically you use it as part of a server backend), you can use libxvidcore and go-xvid, even for commercial use, without sharing the source code of your program.
