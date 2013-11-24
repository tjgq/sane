// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sane

import (
	"bytes"
)

type Frame struct {
	Format       Format // frame format
	Width        int    // width in pixels
	Height       int    // height in pixels
	Channels     int    // number of channels
	Depth        int    // bits per sample
	bytesPerLine int    // bytes per line, including any padding
	data         []byte // raw data
}

// ReadFrame reads and returns a whole frame,
// and whether it is the last frame in an image.
func (c *Conn) ReadFrame() (f *Frame, isLast bool, err error) {
	if err := c.Start(); err != nil {
		return nil, true, err
	}

	p, err := c.Params()
	if err != nil {
		return nil, true, err
	}

	data := new(bytes.Buffer)
	if p.Lines > 0 {
		// Preallocate buffer with expected size
		data = bytes.NewBuffer(make([]byte, 0, p.Lines*p.BytesPerLine))
	}

	if _, err := data.ReadFrom(c); err != nil {
		return nil, true, err
	}

	nch := 1
	if p.Format == FrameRgb {
		nch = 3
	}

	return &Frame{
		Format:       p.Format,
		Width:        p.PixelsPerLine,
		Height:       data.Len() / p.BytesPerLine, // p.Lines is unreliable
		Channels:     nch,
		Depth:        p.Depth,
		bytesPerLine: p.BytesPerLine,
		data:         data.Bytes()}, p.IsLast, nil
}

// At returns the sample at coordinates (x,y) for channel ch.
func (f *Frame) At(x, y, ch int) uint8 {
	if f.Depth == 8 {
		return uint8(f.data[f.bytesPerLine*y+f.Channels*x+ch])
	}
	return 0xFF // TODO: support other depths
}
