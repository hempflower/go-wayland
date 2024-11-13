# Wayland implementation in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/hempflower/go-wayland/wayland.svg)](https://pkg.go.dev/github.com/hempflower/go-wayland/wayland)

This module contains pure Go implementation of the Wayland protocol.
Currently only wayland-client functionality is supported.

Go code is generated from protocol XML files using
[`go-wayland-scanner`](cmd/go-wayland-scanner/scanner.go).

To load cursor, minimal port of `wayland-cursor` & `xcursor` in pure Go
is located at [`wayland/cursor`](wayland/cursor) & [`wayland/cursor/xcursor`](wayland/cursor/xcursor)
respectively.
