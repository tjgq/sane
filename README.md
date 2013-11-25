# go-sane

This package provides bindings to version 1 of the SANE scanner API
for the Go programming language.

## INSTALLING

Just run `go get github.com/tjgq/go-sane`.

The bindings are generated against `libsane` using `cgo`.
You will need to have the appropriate development packages installed.

## USING

A sample program is provided in the `example` subdirectory.
It (mostly) mimics the `scanimage` utility shipped with SANE.

Further information about the SANE API can be found at the
[SANE Project website](http://www.sane-project.org).

## BUGS

API coverage is not yet complete; refer to the bug tracker for the missing stuff.

The package contains a test suite that runs against the SANE test device.
However, more testing with real-world devices is always welcome.

## LICENSE

This library is available under a BSD-style license.
See `LICENSE` for details.

## FEEDBACK

Feel free to report bugs, make suggestions, or contribute improvements!
You can reach me at `tiagoq at gmail dot com`.
