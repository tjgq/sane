// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sane

import (
	"bytes"
	"fmt"
)

type Frame struct {
	Format       Format // frame format
	Width        int    // width in pixels
	Height       int    // height in pixels
	Channels     int    // number of channels
	Depth        int    // bits per sample
	IsLast       bool   // whether this is the last frame
	bytesPerLine int    // bytes per line, including any padding
	data         []byte // raw data
}

// ReadFrame reads and returns a whole frame.
func (c *Conn) ReadFrame() (f *Frame, err error) {
	if err := c.Start(); err != nil {
		return nil, err
	}

	p, err := c.Params()
	if err != nil {
		return nil, err
	}

	if p.Depth != 1 && p.Depth != 8 && p.Depth != 16 {
		return nil, fmt.Errorf("unsupported bit depth: %d", p.Depth)
	}

	data := new(bytes.Buffer)
	if p.Lines > 0 {
		// Preallocate buffer with expected size
		data = bytes.NewBuffer(make([]byte, 0, p.Lines*p.BytesPerLine))
	}

	if _, err := data.ReadFrom(c); err != nil {
		return nil, err
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
		IsLast:       p.IsLast,
		bytesPerLine: p.BytesPerLine,
		data:         data.Bytes()}, nil
}

// At returns the sample at coordinates (x,y) for channel ch.
// Note that values are not normalized to the uint16 range,
// so you need to interpret them relative to the color depth.
func (f *Frame) At(x, y, ch int) uint16 {
	switch f.Depth {
	case 1:
		i := f.bytesPerLine*y + f.Channels*(x/8) + ch
		s := (f.data[i] >> uint8(x%8)) & 0x01
		if f.Format == FrameGray {
			// For B&W lineart, 0 is white and 1 is black
			return uint16(s ^ 0x1)
		} else {
			return uint16(s)
		}
	case 8:
		i := f.bytesPerLine*y + f.Channels*x + ch
		return uint16(f.data[i])
	case 16:
		i := f.bytesPerLine*y + 2*(f.Channels*x+ch)
		return uint16(f.data[i+1])<<8 + uint16(f.data[i])
	}
	return 0
}
