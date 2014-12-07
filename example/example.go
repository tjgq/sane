// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test program for sane package
package main

import (
	"fmt"
	"github.com/tjgq/sane"
	"golang.org/x/image/tiff"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var unitName = map[sane.Unit]string{
	sane.UnitPixel:   "pixels",
	sane.UnitBit:     "bits",
	sane.UnitMm:      "millimetres",
	sane.UnitDpi:     "dots per inch",
	sane.UnitPercent: "percent",
	sane.UnitUsec:    "microseconds",
}

type EncodeFunc func(io.Writer, image.Image) error

func pathToEncoder(path string) (EncodeFunc, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return png.Encode, nil
	case ".jpg", ".jpeg":
		return func(w io.Writer, m image.Image) error {
			return jpeg.Encode(w, m, nil)
		}, nil
	case ".tif", ".tiff":
		return func(w io.Writer, m image.Image) error {
			return tiff.Encode(w, m, nil)
		}, nil
	default:
		return nil, fmt.Errorf("unrecognized extension")
	}
}

func print(f string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, f, v...)
}

func printWrapped(text string, indent, width int) {
	// Naïve implementation - won't work with fancy Unicode input
	indentStr := strings.Repeat(" ", indent)
	for _, line := range strings.Split(text, "\n") {
		pos := 0
		for _, word := range strings.Fields(line) {
			if pos+len(word) > width {
				// Will loop forever if there's a word with len > width!
				print("\n")
				pos = 0
			}
			if pos == 0 {
				print("%s%s", indentStr, word)
			} else {
				print(" %s", word)
			}
			pos += len(word) + 1
		}
		print("\n")
	}
}

func printConstraints(o sane.Option) {
	first := true
	if o.IsAutomatic {
		print(" auto")
		first = false
	}
	if o.ConstrRange != nil {
		if first {
			print(" %v..%v", o.ConstrRange.Min, o.ConstrRange.Max)
		} else {
			print("|%v..%v", o.ConstrRange.Min, o.ConstrRange.Max)
		}
		if (o.Type == sane.TypeInt && o.ConstrRange.Quant != 0) ||
			(o.Type == sane.TypeFloat && o.ConstrRange.Quant != 0.0) {
			print(" in steps of %v", o.ConstrRange.Quant)
		}
	} else {
		for _, v := range o.ConstrSet {
			if first {
				print(" %v", v)
				first = false
			} else {
				print("|%v", v)
			}
		}
	}
}

func printOption(o sane.Option, v interface{}) {
	// Print option name
	print("    -%s", o.Name)

	// Print constraints
	printConstraints(o)

	// Print current value
	if v != nil {
		print(" [%v]", v)
	} else {
		if !o.IsActive {
			print(" [inactive]")
		} else {
			print(" [?]")
		}
	}

	// Print unit
	if name, ok := unitName[o.Unit]; ok {
		print(" %s", name)
	}

	print("\n")

	// Print description
	printWrapped(o.Desc, 8, 70)
}

func findOption(opts []sane.Option, name string) (*sane.Option, error) {
	for _, o := range opts {
		if o.Name == name {
			return &o, nil
		}
	}
	return nil, fmt.Errorf("no such option")
}

func parseBool(s string) (interface{}, error) {
	if s == "yes" || s == "true" || s == "1" {
		return true, nil
	}
	if s == "no" || s == "false" || s == "0" {
		return false, nil
	}
	return nil, fmt.Errorf("not a boolean value")
}

func parseOptions(c *sane.Conn, args []string) error {
	invalidArg := fmt.Errorf("invalid argument")
	if len(args)%2 != 0 {
		return invalidArg // expect option/value pairs
	}
	for i := 0; i < len(args); i += 2 {
		if args[i][0] != '-' || args[i+1][0] == '-' {
			return invalidArg
		}
		o, err := findOption(c.Options(), args[i][1:])
		if err != nil {
			return invalidArg // no such option
		}
		var v interface{}
		if o.IsAutomatic && args[i+1] == "auto" {
			v = sane.Auto // set to auto value
		} else {
			switch o.Type {
			case sane.TypeBool:
				if v, err = parseBool(args[i+1]); err != nil {
					return invalidArg // not a bool
				}
			case sane.TypeInt:
				if v, err = strconv.Atoi(args[i+1]); err != nil {
					return invalidArg // not an int
				}
			case sane.TypeFloat:
				if v, err = strconv.ParseFloat(args[i+1], 64); err != nil {
					return invalidArg // not a float
				}
			case sane.TypeString:
				v = args[i+1]
			}
		}
		if _, err := c.SetOption(o.Name, v); err != nil {
			return err // can't set option
		}
	}
	return nil
}

func openDevice(name string) (*sane.Conn, error) {
	c, err := sane.Open(name)
	if err == nil {
		return c, nil
	}
	// Try a substring match over the available devices
	devs, err := sane.Devices()
	if err != nil {
		return nil, err
	}
	for _, d := range devs {
		if strings.Contains(d.Name, name) {
			return sane.Open(d.Name)
		}
	}
	return nil, fmt.Errorf("no device named %s", name)
}

func listDevices() {
	devs, _ := sane.Devices()
	if len(devs) == 0 {
		print("No available devices.\n")
	}
	for _, d := range devs {
		print("Device %s is a %s %s %s\n", d.Name, d.Vendor, d.Model, d.Type)
	}
}

func showOptions(name string) {
	c, err := openDevice(name)
	if err != nil {
		die(err)
	}
	defer c.Close()

	lastGroup := ""
	print("Options for device %s:\n", c.Device)
	for _, o := range c.Options() {
		if !o.IsSettable {
			continue
		}
		if o.Group != lastGroup {
			print("  %s:\n", o.Group)
			lastGroup = o.Group
		}
		v, _ := c.GetOption(o.Name)
		printOption(o, v)
	}
}

func doScan(deviceName string, fileName string, optargs []string) {
	enc, err := pathToEncoder(fileName)
	if err != nil {
		die(err)
	}

	f, err := os.Create(fileName)
	if err != nil {
		die(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			die(err)
		}
	}()

	c, err := openDevice(deviceName)
	if err != nil {
		die(err)
	}
	defer c.Close()

	if err := parseOptions(c, optargs); err != nil {
		die(err)
	}

	img, err := c.ReadImage()
	if err != nil {
		die(err)
	}

	if err := enc(f, img); err != nil {
		die(err)
	}
}

func usage() {
	exeName := path.Base(os.Args[0])
	print("Usage: %s list\n", exeName)
	print("       %s show <device-name>\n", exeName)
	print("       %s scan <device-name> <output-file> [OPTIONS...]\n", exeName)
	os.Exit(1)
}

func die(v ...interface{}) {
	if len(v) > 0 {
		fmt.Fprintln(os.Stderr, v...)
	}
	os.Exit(1)
}

func main() {
	if err := sane.Init(); err != nil {
		die(err)
	}
	defer sane.Exit()

	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "list":
		listDevices()
	case "show":
		if len(os.Args) != 3 {
			usage()
		}
		showOptions(os.Args[2])
	case "scan":
		if len(os.Args) < 4 {
			usage()
		}
		doScan(os.Args[2], os.Args[3], os.Args[4:])
	default:
		usage()
	}

	os.Exit(0)
}
