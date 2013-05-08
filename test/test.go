// Copyright (C) 2013 Tiago Quelhas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test program for sane package
package main

import (
	"fmt"
	"image/png"
	"os"
	"sane"
)

func main() {
	fmt.Fprintln(os.Stderr, "List devices...")
	devs, err := sane.Devices()
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(os.Stderr, "Open device", devs[0].Name, "...")
	conn, err := sane.Open(devs[0].Name)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Fprintln(os.Stderr, "List options...")
	for i, o := range conn.Options() {
		fmt.Fprintln(os.Stderr, "Option", i, "is", o.Name)
	}

	fmt.Fprintln(os.Stderr, "Set DPI...")
	if _, err := conn.SetOption("resolution", 300); err != nil {
		panic(err)
	}

	fmt.Fprintln(os.Stderr, "Set default mode...")
	if _, err := conn.SetOption("mode", sane.Auto); err != nil {
		panic(err)
	}
	mode, _ := conn.GetOption("mode")
	fmt.Fprintln(os.Stderr, "Mode is", mode.(string))

	fmt.Fprintln(os.Stderr, "Read image...")
	img, err := conn.ReadImage()
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(os.Stderr, "Write output...")
	if err := png.Encode(os.Stdout, img); err != nil {
		panic(err)
	}
}
