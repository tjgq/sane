// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test program for sane package
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path"
	"sane"
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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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
			pos += len(word)
		}
		print("\n")
	}
}

func printIntList(l []int) {
	first := true
	for _, i := range l {
		if first {
			print(" %d", i)
			first = false
		} else {
			print("|%d", i)
		}
	}
}

func printStrList(l []string) {
	first := true
	for _, s := range l {
		if first {
			print(" %s", s)
			first = false
		} else {
			print("|%s", s)
		}
	}
}

func printRange(r *sane.Range) {
	if r != nil {
		print(" %d..%d (in steps of %d)", r.Min, r.Max, r.Quant)
	}
}

func printOption(o sane.Option, v interface{}) {
	// Print option name
	print("    -%s", o.Name)

	// Print constraints
	printIntList(o.IntConstr)
	printStrList(o.StrConstr)
	printRange(o.RangeConstr)

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

func parseOptions(c *sane.Conn, args []string) error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	boolOpt := make(map[string]*bool)
	intOpt := make(map[string]*int)
	strOpt := make(map[string]*string)
	for _, o := range c.Options() {
		switch o.Type {
		case sane.TYPE_BOOL:
			boolOpt[o.Name] = fs.Bool(o.Name, false, "")
		case sane.TYPE_INT, sane.TYPE_FIXED:
			intOpt[o.Name] = fs.Int(o.Name, 0, "")
		case sane.TYPE_STRING:
			strOpt[o.Name] = fs.String(o.Name, "", "")
		}
	}
	fs.Usage = func() {} // don't print usage notice on parse error
	if err := fs.Parse(args); err != nil {
		return err
	}
	fs.Visit(func(f *flag.Flag) {
		for _, o := range c.Options() {
			if o.Name == f.Name {
				var err error
				switch o.Type {
				case sane.TYPE_BOOL:
					_, err = c.SetOption(o.Name, *boolOpt[o.Name])
				case sane.TYPE_INT, sane.TYPE_FIXED:
					_, err = c.SetOption(o.Name, *intOpt[o.Name])
				case sane.TYPE_STRING:
					_, err = c.SetOption(o.Name, *strOpt[o.Name])
				}
				if err != nil {
					die(err)
				}
				return
			}
		}
	})
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
	var err error

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
		die()
	}

	img, err := c.ReadImage()
	if err != nil {
		die(err)
	}

	if err = png.Encode(f, img); err != nil {
		die(err)
	}

	os.Exit(0)
}
