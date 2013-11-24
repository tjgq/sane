// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sane

import (
	"fmt"
	"image"
	"image/color"
)

var opaque = uint8(0xff) // no transparency 8-bit alpha value

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
	if m.fs[0].Format == FrameGray {
		return color.GrayModel
	} else {
		return color.RGBAModel
	}
}

// At returns the color of the pixel at (x, y).
func (m *Image) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(m.Bounds())) {
		return color.RGBA{}
	}
	switch m.fs[0].Format {
	case FrameRed, FrameGreen, FrameBlue:
		// non-interleaved RGB
		return color.RGBA{
			m.fs[0].At(x, y, 0),
			m.fs[1].At(x, y, 0),
			m.fs[2].At(x, y, 0),
			opaque}
	case FrameRgb:
		// interleaved RGB
		return color.RGBA{
			m.fs[0].At(x, y, 0),
			m.fs[0].At(x, y, 1),
			m.fs[0].At(x, y, 2),
			opaque}
	default:
		// grayscale
		return color.Gray{m.fs[0].At(x, y, 0)}
	}
}

// ReadImage reads an image from the connection.
func (c *Conn) ReadImage() (*Image, error) {
	if err := c.Start(); err != nil {
		return nil, err
	}
	defer c.Cancel()

	m := Image{}
	done := false
	for i := 0; !done && i < 3; i++ {
		f, err := c.ReadFrame()
		if err != nil {
			return nil, err
		}
		switch {
		case i == 0 && (f.Format == FrameGray || f.Format == FrameRgb):
			m.fs[0] = f
			done = true // single-frame image
		case f.Format == FrameRed && m.fs[0] == nil:
			m.fs[0] = f
		case f.Format == FrameGreen && m.fs[1] == nil:
			m.fs[1] = f
		case f.Format == FrameBlue && m.fs[2] == nil:
			m.fs[2] = f
		default:
			return nil, fmt.Errorf("sane: unexpected frame type %d", f.Format)
		}
	}
	return &m, nil
}
