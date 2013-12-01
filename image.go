// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sane

import (
	"fmt"
	"image"
	"image/color"
)

var (
	opaque8  = uint8(0xff)
	opaque16 = uint16(0xffff)
)

// Image is a scanned image, corresponding to one or more frames.
//
// It implements the image.Image interface.
type Image struct {
	fs [3]*Frame // multiple frames must be in RGB order
}

// Bounds returns the domain for which At returns valid pixels.
func (m *Image) Bounds() image.Rectangle {
	f := m.fs[0]
	return image.Rect(0, 0, f.Width, f.Height)
}

// ColorModel returns the Image's color model.
func (m *Image) ColorModel() color.Model {
	f := m.fs[0]
	switch {
	case f.Depth != 16 && f.Format == FrameGray:
		return color.GrayModel
	case f.Depth == 16 && f.Format == FrameGray:
		return color.Gray16Model
	case f.Depth != 16 && f.Format != FrameGray:
		return color.RGBAModel
	case f.Depth == 16 && f.Format != FrameGray:
		return color.RGBA64Model
	}
	return color.RGBAModel
}

// At returns the color of the pixel at (x, y).
func (m *Image) At(x, y int) color.Color {
	if x < 0 || x >= m.fs[0].Width || y < 0 || y >= m.fs[0].Height {
		return color.RGBA{}
	}
	if m.fs[0].Format == FrameGray {
		// grayscale
		switch m.fs[0].Depth {
		case 1:
			return color.Gray{uint8(0xFF * m.fs[0].At(x, y, 0))}
		case 8:
			return color.Gray{uint8(m.fs[0].At(x, y, 0))}
		case 16:
			return color.Gray16{m.fs[0].At(x, y, 0)}
		}
	} else {
		// color
		var r, g, b uint16
		if m.fs[0].Format == FrameRgb {
			// interleaved
			r = m.fs[0].At(x, y, 0)
			g = m.fs[0].At(x, y, 1)
			b = m.fs[0].At(x, y, 2)
		} else {
			// non-interleaved
			r = m.fs[0].At(x, y, 0)
			g = m.fs[1].At(x, y, 0)
			b = m.fs[2].At(x, y, 0)
		}
		switch m.fs[0].Depth {
		case 1:
			return color.RGBA{uint8(0xFF * r), uint8(0xFF * g), uint8(0xFF * b), opaque8}
		case 8:
			return color.RGBA{uint8(r), uint8(g), uint8(b), opaque8}
		case 16:
			return color.RGBA64{r, g, b, opaque16}
		}
	}
	return color.RGBA{} // shouldn't happen
}

// ReadImage reads an image from the connection.
func (c *Conn) ReadImage() (*Image, error) {
	defer c.Cancel()

	m := Image{}
	for {
		f, err := c.ReadFrame()
		if err != nil {
			return nil, err
		}
		switch f.Format {
		case FrameGray, FrameRgb, FrameRed:
			m.fs[0] = f
		case FrameGreen:
			m.fs[1] = f
		case FrameBlue:
			m.fs[2] = f
		default:
			return nil, fmt.Errorf("unknown frame type %d", f.Format)
		}
		if f.IsLast {
			break
		}
	}
	return &m, nil
}
