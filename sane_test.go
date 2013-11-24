// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sane

import (
	"image/color"
	"testing"
)

const TestDevice = "test"

var typeMap = map[Type]string{
	TypeBool:   "bool",
	TypeInt:    "int",
	TypeFloat:  "float",
	TypeString: "string",
	TypeButton: "button",
}

var unitMap = map[Unit]string{
	UnitNone:    "none",
	UnitPixel:   "pixel",
	UnitBit:     "bit",
	UnitMm:      "mm",
	UnitDpi:     "dpi",
	UnitPercent: "percent",
	UnitMsec:    "milliseconds",
}

func setOption(t *testing.T, c *Conn, name string, val interface{}) Info {
	i, err := c.SetOption(name, val)
	if err != nil {
		t.Fatalf("set option %s to %v failed: %v", name, val, err)
	}
	return i
}

func readImage(t *testing.T, c *Conn) *Image {
	m, err := c.ReadImage()
	if err != nil {
		t.Fatal("read image failed:", err)
	}
	b := m.Bounds()
	if b.Min.X != 0 || b.Min.Y != 0 || b.Max.X <= b.Min.X || b.Max.Y <= b.Min.Y {
		t.Fatal("bad bounds:", b)
	}
	return m
}

func checkGray(t *testing.T, m *Image) {
	// Areas of 4 x 4 pixels and a distance of 1 pixel between each other
	// and to the borders. Starting with black to white in a line of 256
	// areas. The next line is white to black. The background is medium
	// gray (0x55).
	if m.ColorModel() != color.GrayModel {
		t.Fatal("bad color model")
	}
	b := m.Bounds()
	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			var c color.Gray
			xPos, yPos := x/5, y/5
			switch {
			case x%5 == 0 || y%5 == 0:
				c = color.Gray{0x55}
			case yPos%2 == 0:
				c = color.Gray{uint8(xPos % 0xFF)}
			case yPos%2 == 1:
				c = color.Gray{0xFF - uint8(xPos%0xFF)}
			}
			if m.At(x, y) != c {
				t.Fatalf("bad pixel at (%d,%d): %v should be %v",
					x, y, xPos, yPos, m.At(x, y), c)
			}
		}
	}
}

func checkColor(t *testing.T, m *Image) {
	// Areas of 4 x 4 pixels and a distance of 1 pixel between each other
	// and to the borders. Starting with black to red in a line of 256
	// areas. The next line is red to black. The 3rd and 4th line is green,
	// the 5th and 6th blue. The background is medium gray (0x55).
	if m.ColorModel() != color.RGBAModel {
		t.Fatal("bad color model")
	}
	b := m.Bounds()
	for x := 0; x < b.Max.X; x++ {
		for y := 0; y < b.Max.Y; y++ {
			var (
				s uint8
				c color.RGBA
			)
			xPos, yPos := x/5, y/5
			if x%5 == 0 || y%5 == 0 {
				c = color.RGBA{0x55, 0x55, 0x55, 0xFF}
			} else {
				if yPos%2 == 0 {
					s = uint8(xPos % 0xFF)
				} else {
					s = uint8(0xFF - (xPos % 0xFF))
				}
				switch yPos % 6 {
				case 0, 1:
					c = color.RGBA{s, 0, 0, 0xFF}
				case 2, 3:
					c = color.RGBA{0, s, 0, 0xFF}
				case 4, 5:
					c = color.RGBA{0, 0, s, 0xFF}
				}
			}
			if m.At(x, y) != c {
				t.Fatalf("bad pixel at (%d,%d): %v should be %v",
					x, y, m.At(x, y), c)
			}
		}
	}
}

func checkOptionType(t *testing.T, o *Option, val interface{}) {
	typeName := typeMap[o.Type]
	switch val.(type) {
	case bool:
		if o.Type != TypeBool {
			t.Errorf("option %s has type bool, should be %s", o.Name, typeName)
		}
	case int:
		if o.Type != TypeInt {
			t.Errorf("options %s has type int, should be %s", o.Name, typeName)
		}
	case float64:
		if o.Type != TypeFloat {
			t.Errorf("option %s has type float, should be %s", o.Name, typeName)
		}
	case string:
		if o.Type != TypeString {
			t.Errorf("option %s has type string, should be %s", o.Name, typeName)
		}
	default:
		t.Errorf("option %s has unexpected type, should be %s", o.Name, typeName)
	}
}

func runTest(t *testing.T, f func(c *Conn)) {
	if err := Init(); err != nil {
		t.Fatal("init failed:", err)
	}
	defer Exit()
	c, err := Open(TestDevice)
	if err != nil {
		t.Fatal("open failed:", err)
	}
	defer c.Close()
	f(c)
}

func runGrayTest(t *testing.T, f func(c *Conn) *Image) {
	runTest(t, func(c *Conn) {
		setOption(t, c, "test-picture", "Color pattern")
		checkGray(t, f(c))
	})
}

func runColorTest(t *testing.T, f func(c *Conn) *Image) {
	runTest(t, func(c *Conn) {
		setOption(t, c, "mode", "Color")
		setOption(t, c, "test-picture", "Color pattern")
		checkColor(t, f(c))
	})
}

func TestDevices(t *testing.T) {
	if _, err := Devices(); err != nil {
		t.Fatal("list devices failed:", err)
	}
}

func TestOptions(t *testing.T) {
	runTest(t, func(c *Conn) {
		for _, o := range c.Options() {
			if _, ok := typeMap[o.Type]; !ok {
				t.Errorf("unknown type %d for option %s", o.Type, o.Name)
			}
			if _, ok := unitMap[o.Unit]; !ok {
				t.Errorf("unknown unit %d for option %s", o.Unit, o.Name)
			}
			if !o.IsActive {
				continue
			}
			if o.Type == TypeButton {
				return
			}
			val, err := c.GetOption(o.Name)
			if err != nil {
				t.Errorf("get option %s failed: %v", o.Name, err)
			} else {
				checkOptionType(t, &o, val)
			}
		}
	})
}

func TestGrayImage(t *testing.T) {
	runGrayTest(t, func(c *Conn) *Image {
		return readImage(t, c)
	})
}

func TestColorImage(t *testing.T) {
	runColorTest(t, func(c *Conn) *Image {
		return readImage(t, c)
	})
}

func TestThreePass(t *testing.T) {
	order := []string{"RGB", "RBG", "GBR", "GRB", "BRG", "BGR"}
	for _, o := range order {
		runColorTest(t, func(c *Conn) *Image {
			setOption(t, c, "three-pass", true)
			setOption(t, c, "three-pass-order", o)
			return readImage(t, c)
		})
	}
}
