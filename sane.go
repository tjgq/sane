// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sane

// #cgo LDFLAGS: -lsane
// #include <sane/sane.h>
import "C"

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

var (
	boolType  = reflect.TypeOf(false)
	intType   = reflect.TypeOf(0)
	floatType = reflect.TypeOf(0.0)
)

const wordSize = unsafe.Sizeof(C.SANE_Word(0))

// Type represents the data type of an option.
type Type int

// Type constants.
const (
	TypeBool   Type = C.SANE_TYPE_BOOL
	TypeInt         = C.SANE_TYPE_INT
	TypeFloat       = C.SANE_TYPE_FIXED
	TypeString      = C.SANE_TYPE_STRING
	TypeButton      = C.SANE_TYPE_BUTTON
	typeGroup       = C.SANE_TYPE_GROUP // internal use only
)

// Unit represents the physical unit of an option.
type Unit int

// Unit constants.
const (
	UnitNone    Unit = C.SANE_UNIT_NONE
	UnitPixel        = C.SANE_UNIT_PIXEL
	UnitBit          = C.SANE_UNIT_BIT
	UnitMm           = C.SANE_UNIT_MM
	UnitDpi          = C.SANE_UNIT_DPI
	UnitPercent      = C.SANE_UNIT_PERCENT
	UnitUsec         = C.SANE_UNIT_MICROSECOND
)

// Format represents the format of a frame.
type Format int

// Format constants.
const (
	FrameGray  Format = C.SANE_FRAME_GRAY
	FrameRgb          = C.SANE_FRAME_RGB
	FrameRed          = C.SANE_FRAME_RED
	FrameGreen        = C.SANE_FRAME_GREEN
	FrameBlue         = C.SANE_FRAME_BLUE
)

// Info signals the side effects of setting an option.
type Info struct {
	Inexact      bool // option set to an approximate value
	ReloadOpts   bool // option affects value or availability of other options
	ReloadParams bool // option affects scanning parameters
}

// A Range is a set of discrete integer or fixed-point values. Value x is in
// the range if there is an integer k >= 0 such that Min <= k*Quant <= Max.
// The type of Min, Max and Quant is either int or float64 for all three.
type Range struct {
	Min   interface{} // minimum value
	Max   interface{} // maximum value
	Quant interface{} // quantization step
}

// Option represents a scanning option.
type Option struct {
	Name         string        // option name
	Group        string        // option group
	Title        string        // option title
	Desc         string        // option description
	Type         Type          // option type
	Unit         Unit          // units
	Length       int           // vector length for vector-valued options
	ConstrSet    []interface{} // constraint set
	ConstrRange  *Range        // constraint range
	IsActive     bool          // whether option is active
	IsSettable   bool          // whether option can be set
	IsDetectable bool          // whether option value can be detected
	IsAutomatic  bool          // whether option has an auto value
	IsEmulated   bool          // whether option is emulated
	IsAdvanced   bool          // whether option is advanced
	index        int           // internal option index
	size         int           // internal option size in bytes
}

type autoType int

// Auto is accepted by GetOption to set an option to its automatic value.
var Auto = autoType(0)

// Device represents a scanning device.
type Device struct {
	Name, Vendor, Model, Type string
}

// Conn is a connection to a scanning device. It can be used to get and set
// scanning options or to read one or more frames.
//
// Conn implements the Reader interface. However, it only makes sense to call
// Read after acquisition of a new frame is started by calling Start.
type Conn struct {
	Device  string // device name
	handle  C.SANE_Handle
	options []Option
}

// Params describes the properties of a frame.
type Params struct {
	Format        Format // frame format
	IsLast        bool   // true if last frame in multi-frame image
	BytesPerLine  int    // bytes per line, including any padding
	PixelsPerLine int    // pixels per line
	Lines         int    // number of lines, -1 if unknown
	Depth         int    // bits per sample
}

// Error represents a scanning error.
type Error error

// Error constants.
var (
	ErrUnsupported = errors.New("sane: operation not supported")
	ErrCancelled   = errors.New("sane: operation cancelled")
	ErrBusy        = errors.New("sane: device busy")
	ErrInvalid     = errors.New("sane: invalid argument")
	ErrJammed      = errors.New("sane: feeder jammed")
	ErrEmpty       = errors.New("sane: feeder empty")
	ErrCoverOpen   = errors.New("sane: cover open")
	ErrIo          = errors.New("sane: input/output error")
	ErrNoMem       = errors.New("sane: out of memory")
	ErrDenied      = errors.New("sane: access denied")
)

// mkError converts a libsane status code to an Error.
func mkError(s C.SANE_Status) Error {
	switch s {
	case C.SANE_STATUS_UNSUPPORTED:
		return ErrUnsupported
	case C.SANE_STATUS_CANCELLED:
		return ErrCancelled
	case C.SANE_STATUS_DEVICE_BUSY:
		return ErrBusy
	case C.SANE_STATUS_INVAL:
		return ErrInvalid
	case C.SANE_STATUS_JAMMED:
		return ErrJammed
	case C.SANE_STATUS_NO_DOCS:
		return ErrEmpty
	case C.SANE_STATUS_COVER_OPEN:
		return ErrCoverOpen
	case C.SANE_STATUS_IO_ERROR:
		return ErrIo
	case C.SANE_STATUS_NO_MEM:
		return ErrNoMem
	case C.SANE_STATUS_ACCESS_DENIED:
		return ErrDenied
	default:
		return fmt.Errorf("unknown error code %d", int(s))
	}
}

func boolFromSane(b C.SANE_Word) bool {
	return b != C.SANE_FALSE
}

func boolToSane(b bool) C.SANE_Word {
	if b {
		return C.SANE_TRUE
	}
	return C.SANE_FALSE
}

func intFromSane(i C.SANE_Word) int {
	return int(i)
}

func intToSane(i int) C.SANE_Word {
	return C.SANE_Word(i)
}

func strFromSane(s C.SANE_String_Const) string {
	return C.GoString((*C.char)(unsafe.Pointer(s)))
}

func strToSane(s string) C.SANE_String_Const {
	str := make([]byte, len(s)+1) // +1 for null terminator
	copy(str, s)
	return C.SANE_String_Const(unsafe.Pointer(&str[0]))
}

func floatFromSane(f C.SANE_Word) float64 {
	return float64(f) / (1 << C.SANE_FIXED_SCALE_SHIFT)
}

func floatToSane(f float64) C.SANE_Word {
	return C.SANE_Word(f * (1 << C.SANE_FIXED_SCALE_SHIFT))
}

func nthWord(p *C.SANE_Word, i int) C.SANE_Word {
	a := (*[1 << 16]C.SANE_Word)(unsafe.Pointer(p))
	return a[i]
}

func setNthWord(p *C.SANE_Word, i int, w C.SANE_Word) {
	a := (*[1 << 16]C.SANE_Word)(unsafe.Pointer(p))
	a[i] = w
}

func nthString(p *C.SANE_String_Const, i int) C.SANE_String_Const {
	a := (*[1 << 16]C.SANE_String_Const)(unsafe.Pointer(p))
	return a[i]
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

func nthDevice(p **C.SANE_Device, i int) *C.SANE_Device {
	a := (*[1 << 16]*C.SANE_Device)(unsafe.Pointer(p))
	return a[i]
}

// Devices lists all available devices.
func Devices() (devs []Device, err error) {
	var p **C.SANE_Device
	if s := C.sane_get_devices(&p, C.SANE_FALSE); s != C.SANE_STATUS_GOOD {
		return nil, mkError(s)
	}
	for i := 0; nthDevice(p, i) != nil; i++ {
		p := nthDevice(p, i)
		devs = append(devs, Device{
			strFromSane(p.name),
			strFromSane(p.vendor),
			strFromSane(p.model),
			strFromSane(p._type),
		})
	}
	return devs, nil
}

// Open opens a connection to a device with a given name.
// The empty string opens the first available device.
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
	switch o.Type {
	case TypeInt:
		o.ConstrRange = &Range{
			intFromSane(r.min),
			intFromSane(r.max),
			intFromSane(r.quant)}
	case TypeFloat:
		o.ConstrRange = &Range{
			floatFromSane(r.min),
			floatFromSane(r.max),
			floatFromSane(r.quant)}
	}
}

func parseIntConstr(d *C.SANE_Option_Descriptor, o *Option) {
	p := *(**C.SANE_Word)(unsafe.Pointer(&d.constraint))
	n := intFromSane(nthWord(p, 0))
	// First word is number of remaining words in array.
	for i := 1; i <= n; i++ {
		o.ConstrSet = append(o.ConstrSet, intFromSane(nthWord(p, i)))
	}
}

func parseFloatConstr(d *C.SANE_Option_Descriptor, o *Option) {
	p := *(**C.SANE_Word)(unsafe.Pointer(&d.constraint))
	n := intFromSane(nthWord(p, 0))
	// First word is number of remaining words in array.
	for i := 1; i <= n; i++ {
		o.ConstrSet = append(o.ConstrSet, floatFromSane(nthWord(p, i)))
	}
}

func parseStrConstr(d *C.SANE_Option_Descriptor, o *Option) {
	p := *(**C.SANE_String_Const)(unsafe.Pointer(&d.constraint))
	// Array is null-terminated.
	for i := 0; nthString(p, i) != nil; i++ {
		s := strFromSane(nthString(p, i))
		o.ConstrSet = append(o.ConstrSet, s)
	}
}

func parseOpt(d *C.SANE_Option_Descriptor) (o Option) {
	o.Name = strFromSane(d.name)
	o.Title = strFromSane(d.title)
	o.Desc = strFromSane(d.desc)
	o.Type = Type(d._type)
	o.Unit = Unit(d.unit)
	o.size = int(d.size)
	if o.Type == TypeInt || o.Type == TypeFloat {
		o.Length = o.size / int(wordSize)
	} else {
		o.Length = 1
	}
	switch d.constraint_type {
	case C.SANE_CONSTRAINT_RANGE:
		parseRangeConstr(d, &o)
	case C.SANE_CONSTRAINT_WORD_LIST:
		if o.Type == TypeInt {
			parseIntConstr(d, &o)
		} else {
			parseFloatConstr(d, &o)
		}
	case C.SANE_CONSTRAINT_STRING_LIST:
		parseStrConstr(d, &o)
	}
	o.IsActive = (d.cap & C.SANE_CAP_INACTIVE) == 0
	o.IsSettable = (d.cap & C.SANE_CAP_SOFT_SELECT) != 0
	o.IsDetectable = (d.cap & C.SANE_CAP_SOFT_DETECT) != 0
	o.IsAutomatic = (d.cap & C.SANE_CAP_AUTOMATIC) != 0
	o.IsEmulated = (d.cap & C.SANE_CAP_EMULATED) != 0
	o.IsAdvanced = (d.cap & C.SANE_CAP_ADVANCED) != 0
	return
}

// Options returns a list of available scanning options.
// The list of options usually remains valid until the connection is closed,
// but setting some options may affect the value or availability of others.
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

func readArrayAt(p unsafe.Pointer, i int, t reflect.Type) interface{} {
	ptr := (*C.SANE_Word)(p)
	switch t.Kind() {
	case reflect.Bool:
		return boolFromSane(nthWord(ptr, i))
	case reflect.Int:
		return intFromSane(nthWord(ptr, i))
	case reflect.Float64:
		return floatFromSane(nthWord(ptr, i))
	default:
		return nil
	}
}

func readArray(p unsafe.Pointer, t reflect.Type, n int) interface{} {
	if n == 1 {
		return readArrayAt(p, 0, t)
	}
	v := reflect.MakeSlice(reflect.SliceOf(t), 0, n)
	for i := 0; i < n; i++ {
		v = reflect.Append(v, reflect.ValueOf(readArrayAt(p, i, t)))
	}
	return v.Interface()
}

// GetOption gets the current value for the named option. If successful, it
// returns a value of the appropriate type for the option.
func (c *Conn) GetOption(name string) (interface{}, error) {
	var p unsafe.Pointer
	for _, o := range c.Options() {
		if o.Name == name {
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
				return readArray(p, boolType, o.Length), nil
			case TypeInt:
				return readArray(p, intType, o.Length), nil
			case TypeFloat:
				return readArray(p, floatType, o.Length), nil
			case TypeString:
				return strFromSane(C.SANE_String_Const(p)), nil
			}
		}
	}
	return nil, fmt.Errorf("no option named %s", name)
}

func fillOpt(o Option, v interface{}) (unsafe.Pointer, error) {
	b := make([]byte, o.size)
	p := unsafe.Pointer(&b[0])
	l := o.size / int(wordSize)

	s := ""
	if l > 1 {
		s = "[]"
	}

	switch o.Type {
	case TypeBool:
		if !writeArray(p, boolType, l, v) {
			return nil, fmt.Errorf("option %s expects %sbool arg", o.Name, s)
		}
	case TypeInt:
		if !writeArray(p, intType, l, v) {
			return nil, fmt.Errorf("option %s expects %sint arg", o.Name, s)
		}
	case TypeFloat:
		if !writeArray(p, floatType, l, v) {
			return nil, fmt.Errorf("option %s expects %sfloat64 arg", o.Name, s)
		}
	case TypeString:
		if _, ok := v.(string); !ok {
			return nil, fmt.Errorf("option %s expects string arg", o.Name)
		}
		copy(b, v.(string))
		b[len(b)-1] = byte(0) // ensure null terminator when len(s) == len(v)
	}

	return p, nil
}

func writeArrayAt(p unsafe.Pointer, i int, v reflect.Value) {
	ptr := (*C.SANE_Word)(p)
	switch v.Type().Kind() {
	case reflect.Bool:
		setNthWord(ptr, i, boolToSane(v.Bool()))
	case reflect.Int:
		setNthWord(ptr, i, intToSane(int(v.Int())))
	case reflect.Float64:
		setNthWord(ptr, i, floatToSane(v.Float()))
	}
}

func writeArray(p unsafe.Pointer, t reflect.Type, n int, v interface{}) bool {
	if n == 1 {
		if reflect.TypeOf(v) != t {
			return false
		}
		writeArrayAt(p, 0, reflect.ValueOf(v))
	} else {
		if reflect.TypeOf(v) != reflect.SliceOf(t) {
			return false
		}
		v := reflect.ValueOf(v)
		if v.Len() != n {
			return false
		}
		for i := 0; i < n; i++ {
			writeArrayAt(p, i, v.Index(i))
		}
	}
	return true
}

// SetOption sets the value of the named option, which should be either of the
// corresponding type, or Auto for automatic mode. If successful, info contains
// information on the effects of setting the option.
func (c *Conn) SetOption(name string, v interface{}) (info Info, err error) {
	var (
		s C.SANE_Status
		i C.SANE_Int
	)
	for _, o := range c.Options() {
		if o.Name == name {
			if _, ok := v.(autoType); ok {
				// automatic mode
				s = C.sane_control_option(c.handle, C.SANE_Int(o.index),
					C.SANE_ACTION_SET_AUTO, nil, &i)
			} else {
				p, err := fillOpt(o, v)
				if err != nil {
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
		IsLast:        boolFromSane(C.SANE_Word(p.last_frame)),
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
