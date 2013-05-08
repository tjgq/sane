// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sane

import (
	"fmt"
	"image"
	"image/color"
)

func grayImage(f *Frame) image.Image {
	img := image.NewGray(image.Rect(0, 0, f.Width, f.Height))
	for j := 0; j < f.Height; j++ {
		for i := 0; i < f.Width; i++ {
			c := f.SampleAt(i, j, 0)
			img.Set(i, j, color.Gray{c})
		}
	}
	return img
}

func rgbImage(red, green, blue *Frame) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, red.Width, red.Height))
	for j := 0; j < red.Height; j++ {
		for i := 0; i < red.Width; i++ {
			r := red.SampleAt(i, j, 0)
			g := blue.SampleAt(i, j, 0)
			b := green.SampleAt(i, j, 0)
			img.Set(i, j, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

func interleavedImage(f *Frame) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, f.Width, f.Height))
	for j := 0; j < f.Height; j++ {
		for i := 0; i < f.Width; i++ {
			r := f.SampleAt(i, j, 0)
			g := f.SampleAt(i, j, 1)
			b := f.SampleAt(i, j, 2)
			img.Set(i, j, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

// ReadImage reads a whole image, possibly corresponding to multiple frames.
func (c *Conn) ReadImage() (image.Image, error) {
	f, err := c.ReadFrame()
	if err != nil {
		return nil, err
	}
	switch f.Format {
	case FRAME_GRAY:
		return grayImage(f), nil
	case FRAME_RGB:
		return interleavedImage(f), nil
	case FRAME_RED:
		r := f
		g, err := c.ReadFrame()
		if err != nil {
			return nil, err
		}
		b, err := c.ReadFrame()
		if err != nil {
			return nil, err
		}
		if g.Format == FRAME_GREEN && b.Format == FRAME_BLUE {
			return rgbImage(r, g, b), nil
		}
	}
	return nil, fmt.Errorf("sane: unexpected frame type %d", int(f.Format))
}
