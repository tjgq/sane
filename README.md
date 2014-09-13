# sane

This package provides bindings to the SANE scanner API for the Go programming
language.

## INSTALLING

Run `go get github.com/tjgq/sane`.

The bindings are generated against `libsane` using `cgo`.
You will need to have the appropriate development packages installed.

## USING

Read the package documentation at [GoDoc.org](http://godoc.org/github.com/tjgq/sane).

A sample program is provided in the `example` subdirectory.
It (mostly) mimics the `scanimage` utility shipped with SANE.

Further information about the SANE API can be found at the
[SANE Project website](http://www.sane-project.org).

## BUGS

All SANE functionality is supported except authentication callbacks.

The package contains a test suite that runs against the SANE test device.
However, more testing with real-world devices is always welcome.

## LICENSE

This library is available under the BSD 3-clause license. See `LICENSE` for details.

## FEEDBACK

Feel free to report bugs, make suggestions, or contribute improvements!
