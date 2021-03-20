// Generated by go-wayland-scanner
// https://github.com/rajveermalviya/go-wayland/cmd/go-wayland-scanner
// XML file : https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/d10d18f3d49374d2e3eb96d63511f32795aab5f7/unstable/fullscreen-shell/fullscreen-shell-unstable-v1.xml
//
// FullscreenShellUnstableV1 Protocol Copyright:
//
// Copyright © 2016 Yong Bakos
// Copyright © 2015 Jason Ekstrand
// Copyright © 2015 Jonas Ådahl
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice (including the next
// paragraph) shall be included in all copies or substantial portions of the
// Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package fullscreen_shell

import (
	"sync"

	"github.com/rajveermalviya/go-wayland/client"
)

// ZwpFullscreenShellV1 : displays a single surface per output
//
// Displays a single surface per output.
//
// This interface provides a mechanism for a single client to display
// simple full-screen surfaces.  While there technically may be multiple
// clients bound to this interface, only one of those clients should be
// shown at a time.
//
// To present a surface, the client uses either the present_surface or
// present_surface_for_mode requests.  Presenting a surface takes effect
// on the next wl_surface.commit.  See the individual requests for
// details about scaling and mode switches.
//
// The client can have at most one surface per output at any time.
// Requesting a surface to be presented on an output that already has a
// surface replaces the previously presented surface.  Presenting a null
// surface removes its content and effectively disables the output.
// Exactly what happens when an output is "disabled" is
// compositor-specific.  The same surface may be presented on multiple
// outputs simultaneously.
//
// Once a surface is presented on an output, it stays on that output
// until either the client removes it or the compositor destroys the
// output.  This way, the client can update the output's contents by
// simply attaching a new buffer.
//
// Warning! The protocol described in this file is experimental and
// backward incompatible changes may be made. Backward compatible changes
// may be added together with the corresponding interface version bump.
// Backward incompatible changes are done by bumping the version number in
// the protocol and interface names and resetting the interface version.
// Once the protocol is to be declared stable, the 'z' prefix and the
// version number in the protocol and interface names are removed and the
// interface version number is reset.
type ZwpFullscreenShellV1 struct {
	client.BaseProxy
	mu                 sync.RWMutex
	capabilityHandlers []ZwpFullscreenShellV1CapabilityHandler
}

// NewZwpFullscreenShellV1 : displays a single surface per output
//
// Displays a single surface per output.
//
// This interface provides a mechanism for a single client to display
// simple full-screen surfaces.  While there technically may be multiple
// clients bound to this interface, only one of those clients should be
// shown at a time.
//
// To present a surface, the client uses either the present_surface or
// present_surface_for_mode requests.  Presenting a surface takes effect
// on the next wl_surface.commit.  See the individual requests for
// details about scaling and mode switches.
//
// The client can have at most one surface per output at any time.
// Requesting a surface to be presented on an output that already has a
// surface replaces the previously presented surface.  Presenting a null
// surface removes its content and effectively disables the output.
// Exactly what happens when an output is "disabled" is
// compositor-specific.  The same surface may be presented on multiple
// outputs simultaneously.
//
// Once a surface is presented on an output, it stays on that output
// until either the client removes it or the compositor destroys the
// output.  This way, the client can update the output's contents by
// simply attaching a new buffer.
//
// Warning! The protocol described in this file is experimental and
// backward incompatible changes may be made. Backward compatible changes
// may be added together with the corresponding interface version bump.
// Backward incompatible changes are done by bumping the version number in
// the protocol and interface names and resetting the interface version.
// Once the protocol is to be declared stable, the 'z' prefix and the
// version number in the protocol and interface names are removed and the
// interface version number is reset.
func NewZwpFullscreenShellV1(ctx *client.Context) *ZwpFullscreenShellV1 {
	zwpFullscreenShellV1 := &ZwpFullscreenShellV1{}
	ctx.Register(zwpFullscreenShellV1)
	return zwpFullscreenShellV1
}

// Release : release the wl_fullscreen_shell interface
//
// Release the binding from the wl_fullscreen_shell interface.
//
// This destroys the server-side object and frees this binding.  If
// the client binds to wl_fullscreen_shell multiple times, it may wish
// to free some of those bindings.
//
func (i *ZwpFullscreenShellV1) Release() error {
	err := i.Context().SendRequest(i, 0)
	return err
}

// PresentSurface : present surface for display
//
// Present a surface on the given output.
//
// If the output is null, the compositor will present the surface on
// whatever display (or displays) it thinks best.  In particular, this
// may replace any or all surfaces currently presented so it should
// not be used in combination with placing surfaces on specific
// outputs.
//
// The method parameter is a hint to the compositor for how the surface
// is to be presented.  In particular, it tells the compositor how to
// handle a size mismatch between the presented surface and the
// output.  The compositor is free to ignore this parameter.
//
// The "zoom", "zoom_crop", and "stretch" methods imply a scaling
// operation on the surface.  This will override any kind of output
// scaling, so the buffer_scale property of the surface is effectively
// ignored.
//
func (i *ZwpFullscreenShellV1) PresentSurface(surface *client.WlSurface, method uint32, output *client.WlOutput) error {
	err := i.Context().SendRequest(i, 1, surface, method, output)
	return err
}

// PresentSurfaceForMode : present surface for display at a particular mode
//
// Presents a surface on the given output for a particular mode.
//
// If the current size of the output differs from that of the surface,
// the compositor will attempt to change the size of the output to
// match the surface.  The result of the mode-switch operation will be
// returned via the provided wl_fullscreen_shell_mode_feedback object.
//
// If the current output mode matches the one requested or if the
// compositor successfully switches the mode to match the surface,
// then the mode_successful event will be sent and the output will
// contain the contents of the given surface.  If the compositor
// cannot match the output size to the surface size, the mode_failed
// will be sent and the output will contain the contents of the
// previously presented surface (if any).  If another surface is
// presented on the given output before either of these has a chance
// to happen, the present_cancelled event will be sent.
//
// Due to race conditions and other issues unknown to the client, no
// mode-switch operation is guaranteed to succeed.  However, if the
// mode is one advertised by wl_output.mode or if the compositor
// advertises the ARBITRARY_MODES capability, then the client should
// expect that the mode-switch operation will usually succeed.
//
// If the size of the presented surface changes, the resulting output
// is undefined.  The compositor may attempt to change the output mode
// to compensate.  However, there is no guarantee that a suitable mode
// will be found and the client has no way to be notified of success
// or failure.
//
// The framerate parameter specifies the desired framerate for the
// output in mHz.  The compositor is free to ignore this parameter.  A
// value of 0 indicates that the client has no preference.
//
// If the value of wl_output.scale differs from wl_surface.buffer_scale,
// then the compositor may choose a mode that matches either the buffer
// size or the surface size.  In either case, the surface will fill the
// output.
//
func (i *ZwpFullscreenShellV1) PresentSurfaceForMode(surface *client.WlSurface, output *client.WlOutput, framerate int32) (*ZwpFullscreenShellModeFeedbackV1, error) {
	zwpFullscreenShellModeFeedbackV1 := NewZwpFullscreenShellModeFeedbackV1(i.Context())
	err := i.Context().SendRequest(i, 2, surface, output, framerate, zwpFullscreenShellModeFeedbackV1)
	return zwpFullscreenShellModeFeedbackV1, err
}

// ZwpFullscreenShellV1Capability : capabilities advertised by the compositor
//
// Various capabilities that can be advertised by the compositor.  They
// are advertised one-at-a-time when the wl_fullscreen_shell interface is
// bound.  See the wl_fullscreen_shell.capability event for more details.
//
// ARBITRARY_MODES:
// This is a hint to the client that indicates that the compositor is
// capable of setting practically any mode on its outputs.  If this
// capability is provided, wl_fullscreen_shell.present_surface_for_mode
// will almost never fail and clients should feel free to set whatever
// mode they like.  If the compositor does not advertise this, it may
// still support some modes that are not advertised through wl_global.mode
// but it is less likely.
//
// CURSOR_PLANE:
// This is a hint to the client that indicates that the compositor can
// handle a cursor surface from the client without actually compositing.
// This may be because of a hardware cursor plane or some other mechanism.
// If the compositor does not advertise this capability then setting
// wl_pointer.cursor may degrade performance or be ignored entirely.  If
// CURSOR_PLANE is not advertised, it is recommended that the client draw
// its own cursor and set wl_pointer.cursor(NULL).
const (
	// ZwpFullscreenShellV1CapabilityArbitraryModes : compositor is capable of almost any output mode
	ZwpFullscreenShellV1CapabilityArbitraryModes = 1
	// ZwpFullscreenShellV1CapabilityCursorPlane : compositor has a separate cursor plane
	ZwpFullscreenShellV1CapabilityCursorPlane = 2
)

// ZwpFullscreenShellV1PresentMethod : different method to set the surface fullscreen
//
// Hints to indicate to the compositor how to deal with a conflict
// between the dimensions of the surface and the dimensions of the
// output. The compositor is free to ignore this parameter.
const (
	// ZwpFullscreenShellV1PresentMethodDefault : no preference, apply default policy
	ZwpFullscreenShellV1PresentMethodDefault = 0
	// ZwpFullscreenShellV1PresentMethodCenter : center the surface on the output
	ZwpFullscreenShellV1PresentMethodCenter = 1
	// ZwpFullscreenShellV1PresentMethodZoom : scale the surface, preserving aspect ratio, to the largest size that will fit on the output
	ZwpFullscreenShellV1PresentMethodZoom = 2
	// ZwpFullscreenShellV1PresentMethodZoomCrop : scale the surface, preserving aspect ratio, to fully fill the output cropping if needed
	ZwpFullscreenShellV1PresentMethodZoomCrop = 3
	// ZwpFullscreenShellV1PresentMethodStretch : scale the surface to the size of the output ignoring aspect ratio
	ZwpFullscreenShellV1PresentMethodStretch = 4
)

// ZwpFullscreenShellV1Error : wl_fullscreen_shell error values
//
// These errors can be emitted in response to wl_fullscreen_shell requests.
const (
	// ZwpFullscreenShellV1ErrorInvalidMethod : present_method is not known
	ZwpFullscreenShellV1ErrorInvalidMethod = 0
)

// ZwpFullscreenShellV1CapabilityEvent : advertises a capability of the compositor
//
// Advertises a single capability of the compositor.
//
// When the wl_fullscreen_shell interface is bound, this event is emitted
// once for each capability advertised.  Valid capabilities are given by
// the wl_fullscreen_shell.capability enum.  If clients want to take
// advantage of any of these capabilities, they should use a
// wl_display.sync request immediately after binding to ensure that they
// receive all the capability events.
type ZwpFullscreenShellV1CapabilityEvent struct {
	Capability uint32
}

type ZwpFullscreenShellV1CapabilityHandler interface {
	HandleZwpFullscreenShellV1Capability(ZwpFullscreenShellV1CapabilityEvent)
}

// AddCapabilityHandler : advertises a capability of the compositor
//
// Advertises a single capability of the compositor.
//
// When the wl_fullscreen_shell interface is bound, this event is emitted
// once for each capability advertised.  Valid capabilities are given by
// the wl_fullscreen_shell.capability enum.  If clients want to take
// advantage of any of these capabilities, they should use a
// wl_display.sync request immediately after binding to ensure that they
// receive all the capability events.
func (i *ZwpFullscreenShellV1) AddCapabilityHandler(h ZwpFullscreenShellV1CapabilityHandler) {
	if h == nil {
		return
	}

	i.mu.Lock()
	i.capabilityHandlers = append(i.capabilityHandlers, h)
	i.mu.Unlock()
}

func (i *ZwpFullscreenShellV1) RemoveCapabilityHandler(h ZwpFullscreenShellV1CapabilityHandler) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for j, e := range i.capabilityHandlers {
		if e == h {
			i.capabilityHandlers = append(i.capabilityHandlers[:j], i.capabilityHandlers[j+1:]...)
			break
		}
	}
}

func (i *ZwpFullscreenShellV1) Dispatch(event *client.Event) {
	switch event.Opcode {
	case 0:
		i.mu.RLock()
		if len(i.capabilityHandlers) == 0 {
			i.mu.RUnlock()
			break
		}
		i.mu.RUnlock()

		e := ZwpFullscreenShellV1CapabilityEvent{
			Capability: event.Uint32(),
		}

		i.mu.RLock()
		for _, h := range i.capabilityHandlers {
			i.mu.RUnlock()

			h.HandleZwpFullscreenShellV1Capability(e)

			i.mu.RLock()
		}
		i.mu.RUnlock()
	}
}

// ZwpFullscreenShellModeFeedbackV1 :
type ZwpFullscreenShellModeFeedbackV1 struct {
	client.BaseProxy
	mu                       sync.RWMutex
	modeSuccessfulHandlers   []ZwpFullscreenShellModeFeedbackV1ModeSuccessfulHandler
	modeFailedHandlers       []ZwpFullscreenShellModeFeedbackV1ModeFailedHandler
	presentCancelledHandlers []ZwpFullscreenShellModeFeedbackV1PresentCancelledHandler
}

// NewZwpFullscreenShellModeFeedbackV1 :
func NewZwpFullscreenShellModeFeedbackV1(ctx *client.Context) *ZwpFullscreenShellModeFeedbackV1 {
	zwpFullscreenShellModeFeedbackV1 := &ZwpFullscreenShellModeFeedbackV1{}
	ctx.Register(zwpFullscreenShellModeFeedbackV1)
	return zwpFullscreenShellModeFeedbackV1
}

// ZwpFullscreenShellModeFeedbackV1ModeSuccessfulEvent : mode switch succeeded
//
// This event indicates that the attempted mode switch operation was
// successful.  A surface of the size requested in the mode switch
// will fill the output without scaling.
//
// Upon receiving this event, the client should destroy the
// wl_fullscreen_shell_mode_feedback object.
type ZwpFullscreenShellModeFeedbackV1ModeSuccessfulEvent struct{}
type ZwpFullscreenShellModeFeedbackV1ModeSuccessfulHandler interface {
	HandleZwpFullscreenShellModeFeedbackV1ModeSuccessful(ZwpFullscreenShellModeFeedbackV1ModeSuccessfulEvent)
}

// AddModeSuccessfulHandler : mode switch succeeded
//
// This event indicates that the attempted mode switch operation was
// successful.  A surface of the size requested in the mode switch
// will fill the output without scaling.
//
// Upon receiving this event, the client should destroy the
// wl_fullscreen_shell_mode_feedback object.
func (i *ZwpFullscreenShellModeFeedbackV1) AddModeSuccessfulHandler(h ZwpFullscreenShellModeFeedbackV1ModeSuccessfulHandler) {
	if h == nil {
		return
	}

	i.mu.Lock()
	i.modeSuccessfulHandlers = append(i.modeSuccessfulHandlers, h)
	i.mu.Unlock()
}

func (i *ZwpFullscreenShellModeFeedbackV1) RemoveModeSuccessfulHandler(h ZwpFullscreenShellModeFeedbackV1ModeSuccessfulHandler) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for j, e := range i.modeSuccessfulHandlers {
		if e == h {
			i.modeSuccessfulHandlers = append(i.modeSuccessfulHandlers[:j], i.modeSuccessfulHandlers[j+1:]...)
			break
		}
	}
}

// ZwpFullscreenShellModeFeedbackV1ModeFailedEvent : mode switch failed
//
// This event indicates that the attempted mode switch operation
// failed.  This may be because the requested output mode is not
// possible or it may mean that the compositor does not want to allow it.
//
// Upon receiving this event, the client should destroy the
// wl_fullscreen_shell_mode_feedback object.
type ZwpFullscreenShellModeFeedbackV1ModeFailedEvent struct{}
type ZwpFullscreenShellModeFeedbackV1ModeFailedHandler interface {
	HandleZwpFullscreenShellModeFeedbackV1ModeFailed(ZwpFullscreenShellModeFeedbackV1ModeFailedEvent)
}

// AddModeFailedHandler : mode switch failed
//
// This event indicates that the attempted mode switch operation
// failed.  This may be because the requested output mode is not
// possible or it may mean that the compositor does not want to allow it.
//
// Upon receiving this event, the client should destroy the
// wl_fullscreen_shell_mode_feedback object.
func (i *ZwpFullscreenShellModeFeedbackV1) AddModeFailedHandler(h ZwpFullscreenShellModeFeedbackV1ModeFailedHandler) {
	if h == nil {
		return
	}

	i.mu.Lock()
	i.modeFailedHandlers = append(i.modeFailedHandlers, h)
	i.mu.Unlock()
}

func (i *ZwpFullscreenShellModeFeedbackV1) RemoveModeFailedHandler(h ZwpFullscreenShellModeFeedbackV1ModeFailedHandler) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for j, e := range i.modeFailedHandlers {
		if e == h {
			i.modeFailedHandlers = append(i.modeFailedHandlers[:j], i.modeFailedHandlers[j+1:]...)
			break
		}
	}
}

// ZwpFullscreenShellModeFeedbackV1PresentCancelledEvent : mode switch cancelled
//
// This event indicates that the attempted mode switch operation was
// cancelled.  Most likely this is because the client requested a
// second mode switch before the first one completed.
//
// Upon receiving this event, the client should destroy the
// wl_fullscreen_shell_mode_feedback object.
type ZwpFullscreenShellModeFeedbackV1PresentCancelledEvent struct{}
type ZwpFullscreenShellModeFeedbackV1PresentCancelledHandler interface {
	HandleZwpFullscreenShellModeFeedbackV1PresentCancelled(ZwpFullscreenShellModeFeedbackV1PresentCancelledEvent)
}

// AddPresentCancelledHandler : mode switch cancelled
//
// This event indicates that the attempted mode switch operation was
// cancelled.  Most likely this is because the client requested a
// second mode switch before the first one completed.
//
// Upon receiving this event, the client should destroy the
// wl_fullscreen_shell_mode_feedback object.
func (i *ZwpFullscreenShellModeFeedbackV1) AddPresentCancelledHandler(h ZwpFullscreenShellModeFeedbackV1PresentCancelledHandler) {
	if h == nil {
		return
	}

	i.mu.Lock()
	i.presentCancelledHandlers = append(i.presentCancelledHandlers, h)
	i.mu.Unlock()
}

func (i *ZwpFullscreenShellModeFeedbackV1) RemovePresentCancelledHandler(h ZwpFullscreenShellModeFeedbackV1PresentCancelledHandler) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for j, e := range i.presentCancelledHandlers {
		if e == h {
			i.presentCancelledHandlers = append(i.presentCancelledHandlers[:j], i.presentCancelledHandlers[j+1:]...)
			break
		}
	}
}

func (i *ZwpFullscreenShellModeFeedbackV1) Dispatch(event *client.Event) {
	switch event.Opcode {
	case 0:
		i.mu.RLock()
		if len(i.modeSuccessfulHandlers) == 0 {
			i.mu.RUnlock()
			break
		}
		i.mu.RUnlock()

		e := ZwpFullscreenShellModeFeedbackV1ModeSuccessfulEvent{}

		i.mu.RLock()
		for _, h := range i.modeSuccessfulHandlers {
			i.mu.RUnlock()

			h.HandleZwpFullscreenShellModeFeedbackV1ModeSuccessful(e)

			i.mu.RLock()
		}
		i.mu.RUnlock()
	case 1:
		i.mu.RLock()
		if len(i.modeFailedHandlers) == 0 {
			i.mu.RUnlock()
			break
		}
		i.mu.RUnlock()

		e := ZwpFullscreenShellModeFeedbackV1ModeFailedEvent{}

		i.mu.RLock()
		for _, h := range i.modeFailedHandlers {
			i.mu.RUnlock()

			h.HandleZwpFullscreenShellModeFeedbackV1ModeFailed(e)

			i.mu.RLock()
		}
		i.mu.RUnlock()
	case 2:
		i.mu.RLock()
		if len(i.presentCancelledHandlers) == 0 {
			i.mu.RUnlock()
			break
		}
		i.mu.RUnlock()

		e := ZwpFullscreenShellModeFeedbackV1PresentCancelledEvent{}

		i.mu.RLock()
		for _, h := range i.presentCancelledHandlers {
			i.mu.RUnlock()

			h.HandleZwpFullscreenShellModeFeedbackV1PresentCancelled(e)

			i.mu.RLock()
		}
		i.mu.RUnlock()
	}
}
