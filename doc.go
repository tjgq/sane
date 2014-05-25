// Package sane provides bindings to the SANE scanner API.
//
// The package exposes all of the SANE primitives. It also provides a somewhat
// higher-level interface for scanning one or more images, described below.
//
// Before anything else, you must call Init to initialize the library.
//
//   err := sane.Init()
//
// Call Devices to get a list of the available devices.
//
//   devs, err := sane.Devices()
//
// Open a connection to a device by calling Open with its name. The empty
// string opens the first available device.
//
//   c, err := sane.Open("")
//
// Call Options to retrieve the available options. An option may be set, or
// its current value retrieved, by calling SetOption or GetOption. Note that
// setting an option may affect the value or availability of other options.
//
//   opts := c.Options()
//   val, err := c.GetOption(name)
//   inf, err := c.SetOption(name, val)
//
// To scan an image with the current options, call ReadImage. The returned
// Image object implements the standard library image.Image interface.
//
//   i, err := c.ReadImage()
//
// Although ReadImage blocks, you may interrupt a scan in progress by calling
// Cancel from another goroutine.
//
//   c.Cancel()
//
// Additional images may be scanned while the connection is open. To close the
// connection, call Close.
//
//   c.Close()
//
// Finally, when you are done with the library, call Exit.
//
//   sane.Exit()
//
// If you need finer-grained control over the scanning process, use the
// low-level API, documented at http://www.sane-project.org/html/.
package sane
