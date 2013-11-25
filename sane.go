// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sane provides access to version 1 of the SANE scanner API.
package sane

/*
 #cgo LDFLAGS: -lsane
 #include <sane/sane.h>

 #define SaneWordSize sizeof(SANE_Word)

 // Helpers to avoid unnecessary fiddling around with package unsafe
 SANE_Word nth_word(SANE_Word *v, int n) { return v[n]; }
 SANE_String_Const nth_string(SANE_String_Const *v, int n) { return v[n]; }
 SANE_Device *nth_device(SANE_Device **v, int n) { return v[n]; }
*/
import "C"

import (
	"fmt"
	"io"
	"unsafe"
)

type Type int

const (
	TypeBool   Type = C.SANE_TYPE_BOOL
	TypeInt         = C.SANE_TYPE_INT
	TypeFixed       = C.SANE_TYPE_FIXED
	TypeString      = C.SANE_TYPE_STRING
	TypeButton      = C.SANE_TYPE_BUTTON
	typeGroup       = C.SANE_TYPE_GROUP // internal use only
)

type Unit int

const (
	UnitNone    Unit = C.SANE_UNIT_NONE
	UnitPixel        = C.SANE_UNIT_PIXEL
	UnitBit          = C.SANE_UNIT_BIT
	UnitMm           = C.SANE_UNIT_MM
	UnitDpi          = C.SANE_UNIT_DPI
	UnitPercent      = C.SANE_UNIT_PERCENT
	UnitUsec         = C.SANE_UNIT_MICROSECOND
)

type Format int

const (
	FrameGray  Format = C.SANE_FRAME_GRAY
	FrameRgb          = C.SANE_FRAME_RGB
	FrameRed          = C.SANE_FRAME_RED
	FrameGreen        = C.SANE_FRAME_GREEN
	FrameBlue         = C.SANE_FRAME_BLUE
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
	Name        string        // option name
	Group       string        // option group
	Title       string        // option title
	Desc        string        // option description
	Type        Type          // option type
	Unit        Unit          // units
	Length      int           // vector length for vector-valued options
	ConstrSet   []interface{} // constraint set
	ConstrRange *Range        // constraint range
	IsActive    bool          // whether option is active
	IsSettable  bool          // whether option can be set
	IsAutomatic bool          // whether option has an auto value
	IsEmulated  bool          // whether option is emulated
	IsAdvanced  bool          // whether option is advanced
	index       int           // internal option index
	size        int           // internal option size in bytes
}

type autoType int

var Auto = autoType(0) // automatic mode for SetOption

type Device struct {
	Name, Vendor, Model, Type string
}

// A connection to a device, which can be used to get and set scanning
// options, or to read one or more frames.
//
// It implements the Reader interface, but note that it only makes sense to
// call Read after acquisition of a new frame is started by calling Start.
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

type Error error

var (
	ErrUnsupported = fmt.Errorf("operation not supported")
	ErrCancelled   = fmt.Errorf("operation cancelled")
	ErrBusy        = fmt.Errorf("device busy")
	ErrInvalid     = fmt.Errorf("invalid argument")
	ErrJammed      = fmt.Errorf("feeder jammed")
	ErrEmpty       = fmt.Errorf("feeder empty")
	ErrCoverOpen   = fmt.Errorf("cover open")
	ErrIo          = fmt.Errorf("input/output error")
	ErrNoMem       = fmt.Errorf("out of memory")
	ErrDenied      = fmt.Errorf("access denied")
)

var errMap = map[C.SANE_Status]Error{
	C.SANE_STATUS_UNSUPPORTED:   ErrUnsupported,
	C.SANE_STATUS_CANCELLED:     ErrCancelled,
	C.SANE_STATUS_DEVICE_BUSY:   ErrBusy,
	C.SANE_STATUS_INVAL:         ErrInvalid,
	C.SANE_STATUS_JAMMED:        ErrJammed,
	C.SANE_STATUS_NO_DOCS:       ErrEmpty,
	C.SANE_STATUS_COVER_OPEN:    ErrCoverOpen,
	C.SANE_STATUS_IO_ERROR:      ErrIo,
	C.SANE_STATUS_NO_MEM:        ErrNoMem,
	C.SANE_STATUS_ACCESS_DENIED: ErrDenied,
}

// mkError converts a libsane status code to an Error.
func mkError(s C.SANE_Status) Error {
	err, ok := errMap[s]
	if ok {
		return err
	}
	return fmt.Errorf("unknown error code %d", int(s))
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

// The scale factor used for fixed-point values.
var ScaleFactor = int(1 << C.SANE_FIXED_SCALE_SHIFT)

// FixedToFloat converts a fixed point value to floating point.
func FixedToFloat(f int) float64 {
	return float64(f) / float64(ScaleFactor)
}

// FloatToFixed converts a floating point value to fixed point.
func FloatToFixed(f float64) int {
	return int(f * float64(ScaleFactor))
}

// Init must be called before the package can be used.
func Init() error {
	if s := C.sane_init(nil, nil); s != C.SANE_STATUS_GOOD {
		return mkError(s)
	}
	return nil
}

// Exit releases all resources in use, closing any open connections. The
// package cannot be used after Exit returns and before Init is called again.
func Exit() {
	C.sane_exit()
}

// Devices lists all available devices.
func Devices() (devs []Device, err error) {
	var p **C.SANE_Device
	if s := C.sane_get_devices(&p, C.SANE_FALSE); s != C.SANE_STATUS_GOOD {
		return nil, mkError(s)
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

// Open opens a connection to a device.
func Open(name string) (*Conn, error) {
	var h C.SANE_Handle
	if s := C.sane_open(strToSane(name), &h); s != C.SANE_STATUS_GOOD {
		return nil, mkError(s)
	}
	return &Conn{name, h, nil}, nil
}

// Start initiates the acquisition of a frame.
func (c *Conn) Start() error {
	if s := C.sane_start(c.handle); s != C.SANE_STATUS_GOOD {
		return mkError(s)
	}
	return nil
}

func parseRangeConstr(d *C.SANE_Option_Descriptor, o *Option) {
	r := *(**C.SANE_Range)(unsafe.Pointer(&d.constraint))
	o.ConstrRange = &Range{int(r.min), int(r.max), int(r.quant)}
}

func parseIntConstr(d *C.SANE_Option_Descriptor, o *Option) {
	p := *(**C.SANE_Word)(unsafe.Pointer(&d.constraint))
	// First word is number of remaining words in array.
	for i, n := 1, int(C.nth_word(p, C.int(0))); i <= n; i++ {
		i := int(C.nth_word(p, C.int(i)))
		o.ConstrSet = append(o.ConstrSet, interface{}(i))
	}
}

func parseStrConstr(d *C.SANE_Option_Descriptor, o *Option) {
	p := *(**C.SANE_String_Const)(unsafe.Pointer(&d.constraint))
	// Array is null-terminated.
	for n := 0; C.nth_string(p, C.int(n)) != nil; n++ {
		s := strFromSane(C.nth_string(p, C.int(n)))
		o.ConstrSet = append(o.ConstrSet, interface{}(s))
	}
}

func parseOpt(d *C.SANE_Option_Descriptor) (o Option) {
	o.Name = strFromSane(d.name)
	o.Title = strFromSane(d.title)
	o.Desc = strFromSane(d.desc)
	o.Type = Type(d._type)
	o.Unit = Unit(d.unit)
	o.size = int(d.size)
	if o.Type == TypeInt || o.Type == TypeFixed {
		o.Length = int(d.size / C.SaneWordSize)
	} else {
		o.Length = 1
	}
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
		if opt.Type == typeGroup {
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
			if o.size > 0 {
				p = unsafe.Pointer(&make([]byte, o.size)[0])
			}
			s := C.sane_control_option(c.handle, C.SANE_Int(o.index),
				C.SANE_ACTION_GET_VALUE, p, nil)
			if s != C.SANE_STATUS_GOOD {
				return nil, mkError(s)
			}
			switch o.Type {
			case TypeBool:
				val = interface{}(boolFromSane(*(*C.SANE_Bool)(p)))
			case TypeInt, TypeFixed:
				val = interface{}(int(*(*C.SANE_Int)(p)))
			case TypeString:
				val = interface{}(strFromSane(C.SANE_String_Const(p)))
			}
			return val, err
		}
	}
	return nil, fmt.Errorf("no option named %s", name)
}

func fillOpt(o Option, val interface{}, v []byte) error {
	p := unsafe.Pointer(&v[0])
	switch o.Type {
	case TypeBool:
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("option %s expects bool arg", o.Name)
		}
		q := (*C.SANE_Bool)(p)
		*q = boolToSane(val.(bool))
	case TypeInt, TypeFixed:
		if _, ok := val.(int); !ok {
			return fmt.Errorf("option %s expects int arg", o.Name)
		}
		if o.Type == TypeInt {
			q := (*C.SANE_Int)(p)
			*q = C.SANE_Int(val.(int))
		} else {
			q := (*C.SANE_Fixed)(p)
			*q = C.SANE_Fixed(val.(int))
		}
	case TypeString:
		if _, ok := val.(string); !ok {
			return fmt.Errorf("option %s expects string arg", o.Name)
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
			v := make([]byte, o.size)
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
				return info, mkError(s)
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
	return info, fmt.Errorf("no option named %s", name)
}

// Params retrieves the current scanning parameters. The parameters are
// guaranteed to be accurate between the time the scan is started and the time
// the request is completed or cancelled. Outside that window, they are
// best-effort estimates for the next frame.
func (c *Conn) Params() (Params, error) {
	var p C.SANE_Parameters
	if s := C.sane_get_parameters(c.handle, &p); s != C.SANE_STATUS_GOOD {
		return Params{}, mkError(s)
	}
	return Params{
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
	s := C.sane_read(c.handle, (*C.SANE_Byte)(&b[0]), C.SANE_Int(len(b)), &n)
	if s == C.SANE_STATUS_EOF {
		return 0, io.EOF
	}
	if s != C.SANE_STATUS_GOOD {
		return 0, mkError(s)
	}
	return int(n), nil
}

// Cancel cancels the currently pending operation as soon as possible.
// It returns immediately; when the actual cancellation occurs, the canceled
// operation returns with ErrCancelled.
func (c *Conn) Cancel() {
	C.sane_cancel(c.handle)
}

// Close closes the connection, rendering it unusable for further operations.
func (c *Conn) Close() {
	C.sane_close(c.handle)
	c.handle = nil
	c.options = nil
}
