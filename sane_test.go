// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sane

import (
	"fmt"
	"image/color"
	"reflect"
	"testing"
)

const TestDevice = "test" // the sane test device

var not = map[bool]string{
	true:  "",
	false: "not ",
}

var typeMap = map[Type]string{
	TypeBool:   "bool",
	TypeInt:    "int",
	TypeFloat:  "float",
	TypeString: "string",
	TypeButton: "button",
}

func typeName(t Type) string {
	s, ok := typeMap[t]
	if ok {
		return s
	}
	return fmt.Sprintf("(unknown type with value %d)", int(t))
}

var unitMap = map[Unit]string{
	UnitNone:    "none",
	UnitPixel:   "pixel",
	UnitBit:     "bit",
	UnitMm:      "mm",
	UnitDpi:     "dpi",
	UnitPercent: "percent",
	UnitUsec:    "milliseconds",
}

func unitName(u Unit) string {
	s, ok := unitMap[u]
	if ok {
		return s
	}
	return fmt.Sprintf("(unknown unit with value %d)", int(u))
}

// Test options provided by the sane test device.
var testOpts = []Option{
	{
		Name:         "bool-soft-select-soft-detect",
		Type:         TypeBool,
		Unit:         UnitNone,
		Length:       1,
		IsSettable:   true,
		IsDetectable: true,
		IsAdvanced:   true,
	},
	{
		Name:         "bool-hard-select-soft-detect",
		Type:         TypeBool,
		Unit:         UnitNone,
		Length:       1,
		IsDetectable: true,
		IsAdvanced:   true,
	},
	{
		Name:       "bool-hard-select",
		Type:       TypeBool,
		Unit:       UnitNone,
		Length:     1,
		IsAdvanced: true,
	},
	{
		Name:         "bool-soft-detect",
		Type:         TypeBool,
		Unit:         UnitNone,
		Length:       1,
		IsDetectable: true,
		IsAdvanced:   true,
	},
	{
		Name:         "bool-soft-select-soft-detect-auto",
		Type:         TypeBool,
		Unit:         UnitNone,
		Length:       1,
		IsSettable:   true,
		IsDetectable: true,
		IsAdvanced:   true,
		IsAutomatic:  true,
	},
	{
		Name:         "bool-soft-select-soft-detect-emulated",
		Type:         TypeBool,
		Unit:         UnitNone,
		Length:       1,
		IsSettable:   true,
		IsDetectable: true,
		IsAdvanced:   true,
		IsEmulated:   true,
	},
	{
		Name:         "int",
		Type:         TypeInt,
		Unit:         UnitNone,
		Length:       1,
		IsSettable:   true,
		IsDetectable: true,
		IsAdvanced:   true,
	},
	{
		Name:   "int-constraint-range",
		Type:   TypeInt,
		Unit:   UnitPixel,
		Length: 1,
		ConstrRange: &Range{
			Min:   4,
			Max:   192,
			Quant: 2,
		},
		IsSettable:   true,
		IsDetectable: true,
		IsAdvanced:   true,
	},
	{
		Name:         "int-constraint-word-list",
		Type:         TypeInt,
		Unit:         UnitBit,
		Length:       1,
		ConstrSet:    []interface{}{-42, -8, 0, 17, 42, 256, 65536, 16777216, 1073741824},
		IsSettable:   true,
		IsDetectable: true,
		IsAdvanced:   true,
	},
	{
		Name:         "int-constraint-array-constraint-word-list",
		Type:         TypeInt,
		Unit:         UnitPercent,
		Length:       6,
		ConstrSet:    []interface{}{-42, -8, 0, 17, 42, 256, 65536, 16777216, 1073741824},
		IsSettable:   true,
		IsDetectable: true,
		IsAdvanced:   true,
	},
	{
		Name:         "fixed",
		Type:         TypeFloat,
		Unit:         UnitNone,
		Length:       1,
		IsSettable:   true,
		IsDetectable: true,
		IsAdvanced:   true,
	},
	{
		Name:   "fixed-constraint-range",
		Type:   TypeFloat,
		Unit:   UnitUsec,
		Length: 1,
		ConstrRange: &Range{
			Min:   -42.16999816894531,
			Max:   32767.999893188477,
			Quant: 2.0,
		},
		IsSettable:   true,
		IsDetectable: true,
		IsAdvanced:   true,
	},
	{
		Name:         "fixed-constraint-word-list",
		Type:         TypeFloat,
		Unit:         UnitNone,
		Length:       1,
		ConstrSet:    []interface{}{-32.69999694824219, 12.099990844726562, 42.0, 129.5},
		IsSettable:   true,
		IsDetectable: true,
		IsAdvanced:   true,
	},
	{
		Name:         "string",
		Type:         TypeString,
		Unit:         UnitNone,
		Length:       1,
		IsSettable:   true,
		IsDetectable: true,
	},
	{
		Name:   "string-constraint-string-list",
		Type:   TypeString,
		Unit:   UnitNone,
		Length: 1,
		ConstrSet: []interface{}{
			"First entry",
			"Second entry",
			"This is the very long third entry. Maybe the frontend has an idea how to display it",
		},
		IsSettable:   true,
		IsDetectable: true,
	},
}

type testVal struct {
	name string
	typ  Type
	val  interface{}
}

// Values to test option setting.
var testVals = []testVal{
	{
		name: "enable-test-options",
		typ:  TypeBool,
		val:  true,
	},
	{
		name: "bool-soft-select-soft-detect",
		typ:  TypeBool,
		val:  true,
	},
	{
		name: "int",
		typ:  TypeInt,
		val:  1,
	},
	{
		name: "int-constraint-array",
		typ:  TypeInt,
		val:  []int{1, 2, 3, 4, 5, 6},
	},
	{
		name: "fixed",
		typ:  TypeFloat,
		val:  1.0,
	},
	{
		name: "string",
		typ:  TypeString,
		val:  "Hello world!",
	},
}

func checkOption(t *testing.T, actual, expected *Option) {
	if actual.Name != expected.Name {
		t.Errorf("option %s has wrong name: %v should be %v",
			actual.Name, actual.Name, expected.Name)
	}
	if actual.Type != expected.Type {
		t.Errorf("option %s has wrong type: %s should be %s",
			actual.Name, typeName(actual.Type), typeName(expected.Type))
	}
	if actual.Unit != expected.Unit {
		t.Errorf("option %s has wrong unit: %s should be %s",
			actual.Name, unitName(actual.Unit), unitName(expected.Unit))
	}
	if actual.Length != expected.Length {
		t.Errorf("option %s has wrong length: %d should be %d",
			actual.Name, actual.Length, expected.Length)
	}
	if actual.IsSettable != expected.IsSettable {
		t.Errorf("option %s should %sbe settable",
			actual.Name, not[expected.IsSettable])
	}
	if actual.IsDetectable != expected.IsDetectable {
		t.Errorf("option %s should %sbe detectable",
			actual.Name, not[expected.IsDetectable])
	}
	if actual.IsAutomatic != expected.IsAutomatic {
		t.Errorf("option %s should %sbe automatic",
			actual.Name, not[expected.IsAutomatic])
	}
	if actual.IsEmulated != expected.IsEmulated {
		t.Errorf("option %s should %sbe emulated",
			actual.Name, not[expected.IsEmulated])
	}
	if actual.IsAdvanced != expected.IsAdvanced {
		t.Errorf("option %s should %sbe advanced",
			actual.Name, not[expected.IsAdvanced])
	}
	if !reflect.DeepEqual(actual.ConstrSet, expected.ConstrSet) {
		t.Errorf("option %s has incorrect constraint set: %v should be %v",
			actual.Name, actual.ConstrSet, expected.ConstrSet)
	}
	if !reflect.DeepEqual(actual.ConstrRange, expected.ConstrRange) {
		t.Errorf("option %s has incorrect constraint range: %v should be %v",
			actual.Name, actual.ConstrRange, expected.ConstrRange)
	}
}

func checkOptionType(t *testing.T, o *Option, val interface{}) {
	var (
		ok      bool
		valType string
	)
	switch val.(type) {
	case bool:
		ok = o.Type == TypeBool
		valType = "bool"
	case int:
		ok = o.Type == TypeInt
		valType = "int"
	case float64:
		ok = o.Type == TypeFloat
		valType = "float"
	case string:
		ok = o.Type == TypeString
		valType = "string"
	default:
		ok = false
		valType = "unexpected type"
	}
	if !ok {
		t.Errorf("get option %s returned %s, should return %s",
			o.Name, valType, typeName(o.Type))
	}
}

func findOption(opts []Option, name string) *Option {
	for _, o := range opts {
		if o.Name == name {
			return &o
		}
	}
	return nil
}

func getOption(t *testing.T, c *Conn, name string) interface{} {
	v, err := c.GetOption(name)
	if err != nil {
		t.Fatalf("get option %s failed: %v", name, err)
	}
	return v
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

func runTest(t *testing.T, n int, f func(i int, c *Conn)) {
	if err := Init(); err != nil {
		t.Fatal("init failed:", err)
	}
	defer Exit()
	c, err := Open(TestDevice)
	if err != nil {
		t.Fatal("open failed:", err)
	}
	defer c.Close()
	for i := 0; i < n; i++ {
		if f != nil {
			f(i, c)
		}
	}
}

func runGrayTest(t *testing.T, n int, f func(i int, c *Conn)) {
	runTest(t, n, func(i int, c *Conn) {
		setOption(t, c, "mode", "Gray")
		setOption(t, c, "test-picture", "Color pattern")
		if f != nil {
			f(i, c)
		}
		checkGray(t, readImage(t, c))
	})
}

func runColorTest(t *testing.T, n int, f func(i int, c *Conn)) {
	runTest(t, n, func(i int, c *Conn) {
		setOption(t, c, "mode", "Color")
		setOption(t, c, "test-picture", "Color pattern")
		if f != nil {
			f(i, c)
		}
		checkColor(t, readImage(t, c))
	})
}

func TestDevices(t *testing.T) {
	if _, err := Devices(); err != nil {
		t.Fatal("list devices failed:", err)
	}
}

func TestListOptions(t *testing.T) {
	runTest(t, 1, func(i int, c *Conn) {
		opts := c.Options()
		for _, o := range testOpts {
			opt := findOption(opts, o.Name)
			if opt == nil {
				t.Errorf("option %s missing from list", o.Name)
			} else {
				checkOption(t, opt, &o)
			}
		}
	})
}

func TestGetOptions(t *testing.T) {
	runTest(t, 1, func(i int, c *Conn) {
		for _, o := range c.Options() {
			if !o.IsActive || !o.IsDetectable || o.Type == TypeButton {
				continue
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

func TestSetOptions(t *testing.T) {
	runTest(t, len(testVals), func(i int, c *Conn) {
		optName := testVals[i].name
		//optType := testVals[i].typ
		optVal := testVals[i].val
		setOption(t, c, optName, optVal)
		v := getOption(t, c, optName)
		if !reflect.DeepEqual(interface{}(optVal), v) {
			t.Errorf("get option %s returned wrong value: %v should be %v",
				optName, v, optVal)
		}
	})
}

func TestGray(t *testing.T) {
	runGrayTest(t, 1, nil)
}

func TestGrayTwice(t *testing.T) {
	runGrayTest(t, 2, nil)
}

func TestColor(t *testing.T) {
	runColorTest(t, 1, nil)
}

func TestColorTwice(t *testing.T) {
	runColorTest(t, 2, nil)
}

func TestThreePass(t *testing.T) {
	order := []string{"RGB", "RBG", "GBR", "GRB", "BRG", "BGR"}
	runColorTest(t, len(order), func(i int, c *Conn) {
		setOption(t, c, "three-pass", true)
		setOption(t, c, "three-pass-order", order[i])
	})
}

func TestHandScanner(t *testing.T) {
	runColorTest(t, 1, func(i int, c *Conn) {
		setOption(t, c, "hand-scanner", true)
	})
}

func TestPadding(t *testing.T) {
	runColorTest(t, 1, func(i int, c *Conn) {
		setOption(t, c, "ppl-loss", 7)
	})
}

func TestFuzzyParams(t *testing.T) {
	runColorTest(t, 1, func(i int, c *Conn) {
		setOption(t, c, "fuzzy-parameters", true)
	})
}

func TestReadError(t *testing.T) {
	errList := []struct {
		s string
		e Error
	}{
		{"SANE_STATUS_UNSUPPORTED", ErrUnsupported},
		{"SANE_STATUS_CANCELLED", ErrCancelled},
		{"SANE_STATUS_DEVICE_BUSY", ErrBusy},
		{"SANE_STATUS_INVAL", ErrInvalid},
		{"SANE_STATUS_JAMMED", ErrJammed},
		{"SANE_STATUS_NO_DOCS", ErrEmpty},
		{"SANE_STATUS_COVER_OPEN", ErrCoverOpen},
		{"SANE_STATUS_IO_ERROR", ErrIo},
		{"SANE_STATUS_NO_MEM", ErrNoMem},
		{"SANE_STATUS_ACCESS_DENIED", ErrDenied},
	}
	runTest(t, len(errList), func(i int, c *Conn) {
		setOption(t, c, "read-return-value", errList[i].s)
		_, err := c.ReadImage()
		if err != errList[i].e {
			t.Fatalf("ReadImage returned wrong error: %v should be %v",
				err, errList[i].e)
		}
	})
}

func TestFeeder(t *testing.T) {
	// Feeder has 10 pages
	runTest(t, 11, func(i int, c *Conn) {
		if i == 0 {
			setOption(t, c, "source", "Automatic Document Feeder")
			setOption(t, c, "mode", "Color")
			setOption(t, c, "test-picture", "Color pattern")
		}
		if i < 10 {
			checkColor(t, readImage(t, c))
		} else if _, err := c.ReadImage(); err != ErrEmpty {
			t.Fatalf("feeder not empty after 10 pages")
		}
	})
}

func TestFeederThreePass(t *testing.T) {
	// Feeder has 10 pages
	runTest(t, 11, func(i int, c *Conn) {
		if i == 0 {
			setOption(t, c, "source", "Automatic Document Feeder")
			setOption(t, c, "mode", "Color")
			setOption(t, c, "test-picture", "Color pattern")
			setOption(t, c, "three-pass", true)
		}
		if i < 10 {
			checkColor(t, readImage(t, c))
		} else if _, err := c.ReadImage(); err != ErrEmpty {
			t.Fatalf("feeder not empty after 10 pages")
		}
	})
}

func TestCancel(t *testing.T) {
	runTest(t, 1, func(i int, c *Conn) {
		b := make([]byte, 10)
		if err := c.Start(); err != nil {
			t.Fatalf("start failed: %v", err)
		}
		c.Cancel()
		_, err := c.Read(b)
		if err != ErrCancelled {
			t.Fatalf("read returned wrong error: %v should be %v",
				err, ErrCancelled)
		}
	})
}
