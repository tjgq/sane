// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sane provides access to version 1 of the SANE scanner API.
package sane

/*
 #cgo LDFLAGS: -lsane
 #include <sane/sane.h>

 // Helpers to avoid unnecessary fiddling around with package unsafe
 SANE_Word nth_word(SANE_Word *v, int n) { return v[n]; }
 SANE_String_Const nth_string(SANE_String_Const *v, int n) { return v[n]; }
 SANE_Device *nth_device(SANE_Device **v, int n) { return v[n]; }
*/
import "C"

import (
	"bytes"
	"fmt"
	"io"
	"unsafe"
)

type Type int

const (
	TYPE_BOOL   Type = C.SANE_TYPE_BOOL
	TYPE_INT         = C.SANE_TYPE_INT
	TYPE_FIXED       = C.SANE_TYPE_FIXED
	TYPE_STRING      = C.SANE_TYPE_STRING
	TYPE_BUTTON      = C.SANE_TYPE_BUTTON
	TYPE_GROUP       = C.SANE_TYPE_GROUP
)

type Unit int

const (
	UNIT_NONE        Unit = C.SANE_UNIT_NONE
	UNIT_PIXEL            = C.SANE_UNIT_PIXEL
	UNIT_BIT              = C.SANE_UNIT_BIT
	UNIT_MM               = C.SANE_UNIT_MM
	UNIT_DPI              = C.SANE_UNIT_DPI
	UNIT_PERCENT          = C.SANE_UNIT_PERCENT
	UNIT_MICROSECOND      = C.SANE_UNIT_MICROSECOND
)

type Format int

const (
	FRAME_GRAY  Format = C.SANE_FRAME_GRAY
	FRAME_RGB          = C.SANE_FRAME_RGB
	FRAME_RED          = C.SANE_FRAME_RED
	FRAME_GREEN        = C.SANE_FRAME_GREEN
	FRAME_BLUE         = C.SANE_FRAME_BLUE
)

type Info struct {
	Inexact      bool // option set to an approximate value
	ReloadOpts   bool // option affects value or availability of other options
	ReloadParams bool // option affects scanning parameters
}

type Range struct {
	Min   int // minimum value
	Max   int // maximum value
	Quant int // quantization step
}

type Option struct {
	Name        string   // option name
	Group       string   // option group
	Title       string   // option title
	Desc        string   // option description
	Type        Type     // option type
	Unit        Unit     // units
	Size        int      // option size
	StrConstr   []string // constraint set for string-valued option
	IntConstr   []int    // constraint set for integer-valued option
	RangeConstr *Range   // constraint range for integer-valued option
	IsActive    bool     // whether option is active
	IsSettable  bool     // whether option can be set
	IsAutomatic bool     // whether option has an auto value
	IsEmulated  bool     // whether option is emulated
	IsAdvanced  bool     // whether option is advanced
	index       int      // internal option index
}

type autoType int

var Auto = autoType(0) // automatic mode for SetOption

type Device struct {
	Name, Vendor, Model, Type string
}

type Conn struct {
	Device  string // device name
	handle  C.SANE_Handle
	options []Option
}

type Params struct {
	Format        Format // frame format
	IsLast        bool   // true if last frame in multi-frame image
	BytesPerLine  int    // bytes per line, including any padding
	PixelsPerLine int    // pixels per line
	Lines         int    // number of lines, -1 if unknown
	Depth         int    // bits per sample
}

type Frame struct {
	Format       Format // frame format
	Width        int    // width in pixels
	Height       int    // height in pixels
	Channels     int    // number of channels
	Depth        int    // bits per sample
	bytesPerLine int    // bytes per line, including any padding
	data         []byte // raw data
}

type Error int

const (
	STATUS_GOOD          Error = Error(C.SANE_STATUS_GOOD)
	STATUS_UNSUPPORTED         = Error(C.SANE_STATUS_UNSUPPORTED)
	STATUS_CANCELLED           = Error(C.SANE_STATUS_CANCELLED)
	STATUS_DEVICE_BUSY         = Error(C.SANE_STATUS_DEVICE_BUSY)
	STATUS_INVAL               = Error(C.SANE_STATUS_INVAL)
	STATUS_EOF                 = Error(C.SANE_STATUS_EOF)
	STATUS_JAMMED              = Error(C.SANE_STATUS_JAMMED)
	STATUS_NO_DOCS             = Error(C.SANE_STATUS_NO_DOCS)
	STATUS_COVER_OPEN          = Error(C.SANE_STATUS_COVER_OPEN)
	STATUS_IO_ERROR            = Error(C.SANE_STATUS_IO_ERROR)
	STATUS_NO_MEM              = Error(C.SANE_STATUS_NO_MEM)
	STATUS_ACCESS_DENIED       = Error(C.SANE_STATUS_ACCESS_DENIED)
)

var errText = map[Error]string{
	STATUS_GOOD:          "successful",
	STATUS_UNSUPPORTED:   "operation not supported",
	STATUS_CANCELLED:     "operation cancelled",
	STATUS_DEVICE_BUSY:   "device busy",
	STATUS_INVAL:         "invalid argument",
	STATUS_EOF:           "no more data",
	STATUS_JAMMED:        "feeder jammed",
	STATUS_NO_DOCS:       "feeder empty",
	STATUS_COVER_OPEN:    "cover open",
	STATUS_IO_ERROR:      "input/output error",
	STATUS_NO_MEM:        "out of memory",
	STATUS_ACCESS_DENIED: "access denied",
}

func (e Error) Error() string {
	text, ok := errText[e]
	if ok {
		return fmt.Sprintf("sane: %s", text)
	}
	return fmt.Sprintf("sane: unknown error code %d", e)
}

func boolFromSane(b C.SANE_Bool) bool {
	return b == C.SANE_TRUE
}

func boolToSane(b bool) C.SANE_Bool {
	if b {
		return C.SANE_TRUE
	}
	return C.SANE_FALSE
}

func strFromSane(s C.SANE_String_Const) string {
	return C.GoString((*C.char)(unsafe.Pointer(s)))
}

func strToSane(s string) C.SANE_String_Const {
	str := make([]byte, len(s)+1) // +1 for null terminator
	copy(str, s)
	return C.SANE_String_Const(unsafe.Pointer(&str[0]))
}

// Devices lists all available devices.
func Devices() (devs []Device, err error) {
	var p **C.SANE_Device
	if s := C.sane_get_devices(&p, C.SANE_FALSE); s != C.SANE_STATUS_GOOD {
		return nil, Error(s)
	}
	for n := 0; C.nth_device(p, C.int(n)) != nil; n++ {
		p := C.nth_device(p, C.int(n))
		devs = append(devs, Device{
			strFromSane(p.name),
			strFromSane(p.vendor),
			strFromSane(p.model),
			strFromSane(p._type)})
	}
	return devs, nil
}

// Open opens a connection to a device. If successful, methods on the returned
// connection object can be used to get and set scanning options, or to read
// frames from the device.
func Open(name string) (*Conn, error) {
	var h C.SANE_Handle
	if s := C.sane_open(strToSane(name), &h); s != C.SANE_STATUS_GOOD {
		return nil, Error(s)
	}
	return &Conn{name, h, nil}, nil
}

// Start initiates the acquisition of a frame.
func (c *Conn) Start() error {
	if s := C.sane_start(c.handle); s != C.SANE_STATUS_GOOD {
		return Error(s)
	}
	return nil
}

func parseRangeConstr(d *C.SANE_Option_Descriptor, o *Option) {
	r := *(**C.SANE_Range)(unsafe.Pointer(&d.constraint))
	o.RangeConstr = &Range{int(r.min), int(r.max), int(r.quant)}
}

func parseIntConstr(d *C.SANE_Option_Descriptor, o *Option) {
	p := *(**C.SANE_Word)(unsafe.Pointer(&d.constraint))
	for i, n := 1, int(C.nth_word(p, C.int(0))); i <= n; i++ {
		o.IntConstr = append(o.IntConstr, int(C.nth_word(p, C.int(i))))
	}
}

func parseStrConstr(d *C.SANE_Option_Descriptor, o *Option) {
	p := *(**C.SANE_String_Const)(unsafe.Pointer(&d.constraint))
	for n := 0; C.nth_string(p, C.int(n)) != nil; n++ {
		o.StrConstr = append(o.StrConstr, strFromSane(C.nth_string(p, C.int(n))))
	}
}

func parseOpt(d *C.SANE_Option_Descriptor) (o Option) {
	o.Name = strFromSane(d.name)
	o.Title = strFromSane(d.title)
	o.Desc = strFromSane(d.desc)
	o.Type = Type(d._type)
	o.Unit = Unit(d.unit)
	o.Size = int(d.size)
	switch d.constraint_type {
	case C.SANE_CONSTRAINT_RANGE:
		parseRangeConstr(d, &o)
	case C.SANE_CONSTRAINT_WORD_LIST:
		parseIntConstr(d, &o)
	case C.SANE_CONSTRAINT_STRING_LIST:
		parseStrConstr(d, &o)
	}
	o.IsActive = (d.cap & C.SANE_CAP_INACTIVE) == 0
	o.IsSettable = (d.cap & C.SANE_CAP_SOFT_SELECT) != 0
	o.IsAutomatic = (d.cap & C.SANE_CAP_AUTOMATIC) != 0
	o.IsEmulated = (d.cap & C.SANE_CAP_EMULATED) != 0
	o.IsAdvanced = (d.cap & C.SANE_CAP_ADVANCED) != 0
	return
}

// Options returns a list of available scanning options.

// The list of options usually remains valid until the connection is closed,
// but setting some options may affect the availability of others.
func (c *Conn) Options() (opts []Option) {
	if c.options != nil {
		return c.options // use cached value
	}
	curgroup := ""
	for i := 1; ; i++ {
		desc := C.sane_get_option_descriptor(c.handle, C.SANE_Int(i))
		if desc == nil {
			break
		}
		opt := parseOpt(desc)
		if opt.Type == TYPE_GROUP {
			curgroup = opt.Title
			continue
		}
		opt.Group = curgroup
		opt.index = i
		opts = append(opts, opt)
	}
	c.options = opts
	return
}

// GetOption gets the current value for the named option. If successful, it
// returns a value of the appropriate type for the option.
func (c *Conn) GetOption(name string) (val interface{}, err error) {
	for _, o := range c.Options() {
		if o.Name == name {
			var p unsafe.Pointer
			if o.Size > 0 {
				p = unsafe.Pointer(&make([]byte, o.Size)[0])
			}
			s := C.sane_control_option(c.handle, C.SANE_Int(o.index),
				C.SANE_ACTION_GET_VALUE, p, nil)
			if s != C.SANE_STATUS_GOOD {
				return nil, Error(s)
			}
			switch o.Type {
			case TYPE_BOOL:
				val = interface{}(boolFromSane(*(*C.SANE_Bool)(p)))
			case TYPE_INT, TYPE_FIXED:
				val = interface{}(int(*(*C.SANE_Int)(p)))
			case TYPE_STRING:
				val = interface{}(strFromSane(C.SANE_String_Const(p)))
			}
			return val, err
		}
	}
	return nil, fmt.Errorf("sane: no option named %s", name)
}

func fillOpt(o Option, val interface{}, v []byte) error {
	p := unsafe.Pointer(&v[0])
	switch o.Type {
	case TYPE_BOOL:
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("sane: option %s expects bool arg", o.Name)
		}
		q := (*C.SANE_Bool)(p)
		*q = boolToSane(val.(bool))
	case TYPE_INT, TYPE_FIXED:
		if _, ok := val.(int); !ok {
			return fmt.Errorf("sane: option %s expects int arg", o.Name)
		}
		q := (*C.SANE_Int)(p)
		*q = C.SANE_Int(val.(int))
	case TYPE_STRING:
		if _, ok := val.(string); !ok {
			return fmt.Errorf("sane: option %s expects string arg", o.Name)
		}
		copy(v, val.(string))
		v[len(v)-1] = byte(0) // ensure null terminator when len(s) == len(v)
	}
	return nil
}

// SetOption sets the value of the named option, which should be either of the
// corresponding type, or Auto for automatic mode. If successful, info contains
// information on the effects of setting the option.
func (c *Conn) SetOption(name string, val interface{}) (info Info, err error) {
	for _, o := range c.Options() {
		if o.Name == name {
			var s C.SANE_Status
			var i C.SANE_Int
			v := make([]byte, o.Size)
			p := unsafe.Pointer(&v[0])

			if _, ok := val.(autoType); ok {
				// automatic mode
				s = C.sane_control_option(c.handle, C.SANE_Int(o.index),
					C.SANE_ACTION_SET_AUTO, nil, &i)
			} else {
				if err = fillOpt(o, val, v); err != nil {
					return info, err
				}
				s = C.sane_control_option(c.handle, C.SANE_Int(o.index),
					C.SANE_ACTION_SET_VALUE, p, &i)
			}

			if s != C.SANE_STATUS_GOOD {
				return info, Error(s)
			}

			if int(i)&C.SANE_INFO_INEXACT != 0 {
				info.Inexact = true
			}
			if int(i)&C.SANE_INFO_RELOAD_OPTIONS != 0 {
				info.ReloadOpts = true
				c.options = nil // cached options are no longer valid
			}
			if int(i)&C.SANE_INFO_RELOAD_PARAMS != 0 {
				info.ReloadParams = true
			}
			return info, nil
		}
	}
	return info, fmt.Errorf("sane: no option named %s", name)
}

// Params retrieves the current scanning parameters. The parameters are
// guaranteed to be accurate between the time the scan is started and the time
// the request is completed or cancelled. Outside that window, they are
// best-effort estimates for the next frame.
func (c *Conn) Params() (*Params, error) {
	var p C.SANE_Parameters
	if s := C.sane_get_parameters(c.handle, &p); s != C.SANE_STATUS_GOOD {
		return nil, Error(s)
	}
	return &Params{
		Format:        Format(p.format),
		IsLast:        boolFromSane(p.last_frame),
		BytesPerLine:  int(p.bytes_per_line),
		PixelsPerLine: int(p.pixels_per_line),
		Lines:         int(p.lines),
		Depth:         int(p.depth)}, nil
}

// Read reads up to len(b) bytes of data from the current frame.
// It returns the number of bytes read and an error, if any. If the frame is
// complete, a zero count is returned together with an io.EOF error.
func (c *Conn) Read(b []byte) (int, error) {
	var n C.SANE_Int
	s := C.sane_read(c.handle, (*C.SANE_Byte)(&b[0]), C.SANE_Int(cap(b)), &n)
	if s == C.SANE_STATUS_EOF {
		return 0, io.EOF
	}
	if s != C.SANE_STATUS_GOOD {
		return 0, Error(s)
	}
	return int(n), nil
}

// ReadFrame reads and returns a whole frame.
//
// It automatically calls Start before reading, and Cancel when it's done.
func (c *Conn) ReadFrame() (*Frame, error) {
	if err := c.Start(); err != nil {
		return nil, err
	}
	defer c.Cancel()

	p, err := c.Params()
	if err != nil {
		return nil, err
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
	if p.Format == FRAME_RGB {
		nch = 3
	}

	return &Frame{
		Format:       p.Format,
		Width:        p.PixelsPerLine,
		Height:       data.Len() / p.BytesPerLine, // p.Lines is unreliable
		Channels:     nch,
		Depth:        p.Depth,
		bytesPerLine: p.BytesPerLine,
		data:         data.Bytes()}, nil
}

// Cancel cancels the currently pending operation as soon as possible.
// It merely initiates the cancellation; cancellation is only guaranteed to
// have occurred when the cancelled operation returns.
//
// Note that Cancel must be called after reading an entire frame successfully,
// before another frame can be read.
func (c *Conn) Cancel() {
	C.sane_cancel(c.handle)
}

// Close closes the connection, rendering it unusable for further operations.
func (c *Conn) Close() {
	C.sane_close(c.handle)
	c.handle = nil
	c.options = nil
}

// SampleAt returns the sample at coordinates (x,y) for channel ch.
func (f *Frame) SampleAt(x, y, ch int) uint8 {
	if f.Depth == 8 {
		return uint8(f.data[f.bytesPerLine*y+f.Channels*x+ch])
	}
	return 0xFF // TODO: support other depths
}

func init() {
	if C.sane_init(nil, nil) != C.SANE_STATUS_GOOD {
		panic("sane: can't init")
	}
}
