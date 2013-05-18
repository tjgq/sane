# go-sane

This package provides bindings to version 1 of the SANE scanner API
for the Go programming language.

Please note that this work is still at a very early stage.
In particular, do not rely on the provided API not to change in the future.

## INSTALLING

Just run `go get github.com/tjgq/go-sane/sane`.

The bindings are generated against `libsane` using `cgo`.
You will need to have the appropriate development packages installed.

## USING

There is a sample program in the `test` subdirectory.

Further information about the SANE API can be found at the
[SANE Project website](http://www.sane-project.org).

## BUGS

Very incomplete testing, due to the fact that I only have a single device
to test with. Incomplete API coverage; refer to the bug tracker for the missing
stuff.

## LICENSE

This library is available under a BSD-style license.
See `LICENSE` for details.

## FEEDBACK

Feel free to report bugs, make suggestions, or contribute improvements!
You can reach me at `tiagoq at gmail dot com`.
