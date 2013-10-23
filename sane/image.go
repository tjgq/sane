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
	frames []*Frame
}

// Bounds returns the domain for which At returns valid pixels.
func (i *Image) Bounds() image.Rectangle {
	f := i.frames[0]
	return image.Rect(0, 0, f.Width, f.Height)
}

// ColorModel returns the Image's color model.
func (i *Image) ColorModel() color.Model {
	if i.frames[0].Format == FrameGray {
		return color.GrayModel
	} else {
		return color.RGBAModel
	}
}

// At returns the color of the pixel at (x, y).
func (i *Image) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(i.Bounds())) {
		return color.RGBA{}
	}
	fs := i.frames
	switch {
	case len(fs) == 3:
		// non-interleaved RGB
		return color.RGBA{
			fs[0].At(x, y, 0),
			fs[1].At(x, y, 0),
			fs[2].At(x, y, 0),
			opaque}
	case fs[0].Format == FrameRgb:
		// interleaved RGB
		return color.RGBA{
			fs[0].At(x, y, 0),
			fs[0].At(x, y, 1),
			fs[0].At(x, y, 2),
			opaque}
	default:
		// grayscale
		return color.Gray{fs[0].At(x, y, 0)}
	}
}

// ReadImage reads an image from the connection.
func (c *Conn) ReadImage() (*Image, error) {
	if err := c.Start(); err != nil {
		return nil, err
	}
	defer c.Cancel()

	frames := make([]*Frame, 0, 3)
	done := false
	for i := 0; !done && i < 3; i++ {
		f, err := c.ReadFrame()
		if err != nil {
			return nil, err
		}
		frames = append(frames, f)
		switch {
		case i == 0 && (f.Format == FrameGray || f.Format == FrameRgb):
			done = true // single-frame image
		case (i == 0 && f.Format != FrameRed) ||
			(i == 1 && f.Format != FrameGreen) ||
			(i == 2 && f.Format != FrameBlue):
			// Make sure red/green/blue appear in the right order
			return nil, fmt.Errorf("sane: unexpected frame type %d", f.Format)
		}
	}
	return &Image{frames}, nil
}
