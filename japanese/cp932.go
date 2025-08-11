package japanese

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var (
	errInvalidRune           = errors.New("Transform: invalid rune")
	errInconsistentByteCount = errors.New("Transform: inconsistent byte count returned")
)

// CP932 is the CP932 encoding, compatible with CPython's implementation.
var CP932 encoding.Encoding = &cp932{}

type cp932 struct{}

func (c *cp932) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: &cp932Decoder{decoder: japanese.ShiftJIS.NewDecoder()}}
}

func (c *cp932) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: &cp932Encoder{encoder: japanese.ShiftJIS.NewEncoder()}}
}

type cp932Decoder struct {
	transform.NopResetter

	decoder *encoding.Decoder
}

var _ transform.Transformer = (*cp932Decoder)(nil)

// Uses japanese.ShiftJIS decoder as base, with CPython code as reference for parts with different specifications
// https://github.com/python/cpython/blob/8a0c7f1e402768c7e806e2472e0a493c1800851f/Modules/cjkcodecs/_codecs_jp.c#L84
func (c *cp932Decoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		var r rune
		size := 0
		switch c0 := src[nSrc]; {
		case c0 == 0xA0:
			r, size = 0xF8F0, 1
		case c0 >= 0xF0 && c0 <= 0xF9:
			if nSrc+1 >= len(src) {
				return nDst, nSrc, transform.ErrShortSrc
			}

			switch c1 := src[nSrc+1]; {
			case c1 >= 0x40 && c1 <= 0x7E:
				r, size = rune(0xE000+188*(int(c0)-0xF0)+(int(c1)-0x40)), 2
			case c1 >= 0x80 && c1 <= 0xFC:
				r, size = rune(0xE000+188*(int(c0)-0xF0)+(int(c1)-0x41)), 2
			default:
				r, size = '\ufffd', 1
			}
		case c0 >= 0xFD:
			r, size = rune(0xF8F1-0xFD+int(c0)), 1
		default:
			size = min(len(src)-nSrc, 2)
			eof := atEOF && nSrc+size == len(src)
			d, s, err := c.decoder.Transform(dst[nDst:], src[nSrc:nSrc+size], eof)
			nDst += d
			nSrc += s
			// Continue if ErrShortSrc occurs but there's more data to process
			if err != nil && (err != transform.ErrShortSrc || eof) {
				return nDst, nSrc, err
			}
			if s == 0 {
				return nDst, nSrc, errInconsistentByteCount
			}
			continue
		}
		if size > 0 {
			runeLen := utf8.RuneLen(r)
			if runeLen == -1 {
				return nDst, nSrc, errInvalidRune
			}
			if nDst+runeLen > len(dst) {
				return nDst, nSrc, transform.ErrShortDst
			}
			nDst += utf8.EncodeRune(dst[nDst:], r)
			nSrc += size
		}
	}
	return nDst, nSrc, nil
}

type cp932Encoder struct {
	transform.NopResetter

	encoder *encoding.Encoder
}

var _ transform.Transformer = (*cp932Encoder)(nil)

// Uses japanese.ShiftJIS encoder as base, with CPython code as reference for parts with different specifications
// https://github.com/python/cpython/blob/8a0c7f1e402768c7e806e2472e0a493c1800851f/Modules/cjkcodecs/_codecs_jp.c#L20
func (c *cp932Encoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		r, size := rune(src[nSrc]), 1
		if r >= utf8.RuneSelf {
			r, size = utf8.DecodeRune(src[nSrc:])
			if size == 1 {
				if !atEOF && !utf8.FullRune(src[nSrc:]) {
					return nDst, nSrc, transform.ErrShortSrc
				}
			}
		}

		var d0, d1 byte
		switch {
		case r == 0x80:
			dst[nDst] = 0x80
			nDst++
			nSrc += size
			continue
		case r == 0xA2: // cent sign
			d0, d1 = 0x81, 0x91
		case r == 0xA3: // pound sign
			d0, d1 = 0x81, 0x92
		case r == 0xAC: // not sign
			d0, d1 = 0x81, 0xCA
		case r == 0x2016: // double vertical line
			d0, d1 = 0x81, 0x61
		case r == 0x2212: // minus sign
			d0, d1 = 0x81, 0x7c
		case r == 0x301C: // wave dash
			d0, d1 = 0x81, 0x60
		case r >= 0xE000 && r < 0xE758:
			d0 = 0xF0 + byte((r-0xE000)/188)
			d1 = byte((r - 0xE000) % 188)
			if d1 < 0x3F {
				d1 += 0x40
			} else {
				d1 += 0x41
			}
		case r == 0xF8F0:
			dst[nDst] = 0xA0
			nDst++
			nSrc += size
			continue
		case r >= 0xF8F1 && r <= 0xF8F3:
			dst[nDst] = 0xFD + byte(r-0xF8F1)
			nDst++
			nSrc += size
			continue
		default:
			eof := atEOF && nSrc+size == len(src)
			d, s, err := c.encoder.Transform(dst[nDst:], src[nSrc:nSrc+size], eof)
			nDst += d
			nSrc += s
			if err != nil {
				return nDst, nSrc, err
			}
			continue
		}

		// Write 2 bytes
		if nDst+2 > len(dst) {
			return nDst, nSrc, transform.ErrShortDst
		}
		dst[nDst] = d0
		dst[nDst+1] = d1
		nDst += 2
		nSrc += size
	}
	return nDst, nSrc, nil
}
