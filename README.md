# Wayland implementation in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/SpinningVinyl/go-wayland/wayland.svg)](https://pkg.go.dev/github.com/SpinningVinyl/go-wayland/wayland)

This module contains pure Go implementation of the Wayland protocol.
Currently only wayland-client functionality is supported.

Go code is generated from protocol XML files using
[`go-wayland-scanner`](cmd/go-wayland-scanner/scanner.go).
The bindings are generated from Wayland 1.26.0 and Wayland Protocols 1.49.
Stable, staging, and unstable protocols are included; experimental protocols
are omitted because their interfaces may change incompatibly.
Each package's `doc.go` pins the XML used to regenerate it.

To load cursor, minimal port of `wayland-cursor` & `xcursor` in pure Go
is located at [`wayland/cursor`](wayland/cursor) & [`wayland/cursor/xcursor`](wayland/cursor/xcursor)
respectively.

To demonstrate the functionality of this module
[`examples/imageviewer`](examples/imageviewer) contains a simple image
viewer. It demos displaying a top-level window, resizing of window,
cursor themes, pointer & keyboard. Because it's in pure Go, it can be
compiled without CGO. From a checkout, run:

```sh
CGO_ENABLED=0 go run ./examples/imageviewer file.jpg
```
