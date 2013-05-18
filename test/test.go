// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test program for sane package
package main

import (
	"fmt"
	"image/png"
	"os"
	"path"
	"sane"
	"strconv"
	"strings"
)

var unitName = map[sane.Unit]string{
	sane.UNIT_PIXEL:       "pixels",
	sane.UNIT_BIT:         "bits",
	sane.UNIT_MM:          "millimetres",
	sane.UNIT_DPI:         "dots per inch",
	sane.UNIT_PERCENT:     "percent",
	sane.UNIT_MICROSECOND: "microseconds",
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

func printConstrSet(o sane.Option) {
	first := true
	for _, v := range o.ConstrSet {
		if first {
			print(" %v", v)
			first = false
		} else {
			print("|%v", v)
		}
	}
}

func printConstrRange(o sane.Option) {
	if o.ConstrRange != nil {
		print(" %v..%v", o.ConstrRange.Min, o.ConstrRange.Max)
		if (o.Type == sane.TYPE_INT && o.ConstrRange.Quant != 0) ||
			(o.Type == sane.TYPE_FLOAT && o.ConstrRange.Quant != 0.0) {
			print(" in steps of %v", o.ConstrRange.Quant)
		}
	}
}

func printOption(o sane.Option, v interface{}) {
	// Print option name
	print("    -%s", o.Name)

	// Print constraints
	printConstrSet(o)
	printConstrRange(o)

	// Print current value
	if v != nil {
		print(" [%v]", v)
	} else {
		print(" [?]")
	}

	// Print unit
	if name, ok := unitName[o.Unit]; ok {
		print(" %s", name)
	}

	print("\n")

	// Print description
	printWrapped(o.Desc, 8, 70)
}

func printOptions(c *sane.Conn) {
	lastGroup := ""
	print("Available options for device %s:\n", c.Device)
	for _, o := range c.Options() {
		if o.Group != lastGroup {
			print("  %s:\n", o.Group)
			lastGroup = o.Group
		}
		v, _ := c.GetOption(o.Name)
		printOption(o, v)
	}
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
			case sane.TYPE_BOOL:
				if v, err = parseBool(args[i+1]); err != nil {
					return invalidArg // not a bool
				}
			case sane.TYPE_INT:
				if v, err = strconv.Atoi(args[i+1]); err != nil {
					return invalidArg // not an int
				}
			case sane.TYPE_FLOAT:
				if v, err = strconv.ParseFloat(args[i+1], 64); err != nil {
					return invalidArg // not a float
				}
			case sane.TYPE_STRING:
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
	devs, err := sane.Devices()
	if err != nil {
		return nil, err
	}
	for _, d := range devs {
		// A substring of the device name will do
		if strings.Contains(d.Name, name) {
			return sane.Open(d.Name)
		}
	}
	return nil, fmt.Errorf("no device named %s", name)
}

func help() {
	print("Usage: %s <device-name> <output-file> [OPTIONS...]\n\n", path.Base(os.Args[0]))

	devs, _ := sane.Devices()
	if len(devs) == 0 {
		print("No available devices.\n")
		return
	}

	for _, d := range devs {
		if c, err := sane.Open(d.Name); err == nil {
			printOptions(c)
			c.Close()
		}
	}
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

	if len(os.Args) < 3 {
		help()
		os.Exit(1)
	}

	f, err := os.Create(os.Args[2])
	if err != nil {
		die(err)
	}
	defer f.Close()

	c, err := openDevice(os.Args[1])
	if err != nil {
		die(err)
	}
	defer c.Close()

	if err := parseOptions(c, os.Args[3:]); err != nil {
		die(err)
	}

	img, err := c.ReadImage()
	if err != nil {
		die(err)
	}

	if err := png.Encode(f, img); err != nil {
		die(err)
	}

	os.Exit(0)
}
