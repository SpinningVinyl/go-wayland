package client

import (
	"errors"
	"fmt"
	"net"
	"os"

	"golang.org/x/sys/unix"
)

type Context struct {
	conn       *net.UnixConn
	objects    map[uint32]Proxy
	currentID  uint32
	pendingFDs []int
}

func (ctx *Context) Register(p Proxy) {
	ctx.currentID++
	p.SetID(ctx.currentID)
	p.SetContext(ctx)
	ctx.objects[ctx.currentID] = p
}

// RegisterServer records an object ID allocated by the compositor.
// Unlike Register, it must not advance the client ID sequence.
func (ctx *Context) RegisterServer(p Proxy, id uint32) {
	p.SetID(id)
	p.SetContext(ctx)
	ctx.objects[id] = p
}

func (ctx *Context) Unregister(p Proxy) {
	delete(ctx.objects, p.ID())
}

func (ctx *Context) GetProxy(id uint32) Proxy {
	return ctx.objects[id]
}

func (ctx *Context) Close() error {
	for _, fd := range ctx.pendingFDs {
		_ = unix.Close(fd)
	}
	ctx.pendingFDs = nil
	return ctx.conn.Close()
}

// Dispatch reads and processes incoming messages and calls [client.Dispatcher.Dispatch] on the
// respective wayland protocol.
// Dispatch must be called on the same goroutine as other interactions with the Context.
// If a multi goroutine approach is desired, use [Context.GetDispatch] instead.
// Dispatch blocks if there are no incoming messages.
// A Dispatch loop is usually used to handle incoming messages.
func (ctx *Context) Dispatch() error {
	return ctx.GetDispatch()()
}

var ErrDispatchSenderNotFound = errors.New("dispatch: unable to find sender")
var ErrDispatchSenderUnsupported = errors.New("dispatch: sender does not implement Dispatch method")
var ErrDispatchUnableToReadMsg = errors.New("dispatch: unable to read msg")

// GetDispatch reads incoming messages and returns the dispatch function which calls
// [client.Dispatcher.Dispatch] on the respective wayland protocol.
// While GetDispatch is usually called in a loop in a separate goroutine, the dispatch function it
// returns must be called in the same goroutine as other interactions with the Context.
// GetDispatch blocks if there are no incoming messages.
func (ctx *Context) GetDispatch() func() error {
	senderID, opcode, fd, data, err := ctx.ReadMsg() // Blocks if there are no incoming messages
	if err != nil {
		return func() error {
			return fmt.Errorf("%w: %w", ErrDispatchUnableToReadMsg, err)
		}
	}

	return func() error { return ctx.DispatchMessage(senderID, opcode, fd, data) }
}

// DispatchMessage must run on the same goroutine as other protocol operations.
func (ctx *Context) DispatchMessage(senderID, opcode uint32, fds []int, data []byte) error {
	ctx.pendingFDs = append(ctx.pendingFDs, fds...)
	sender, ok := ctx.objects[senderID]
	if !ok {
		return fmt.Errorf("%w (senderID=%d)", ErrDispatchSenderNotFound, senderID)
	}
	takesFD := false
	if receiver, ok := sender.(interface{ TakesFD(uint32) bool }); ok {
		takesFD = receiver.TakesFD(opcode)
	}
	fd := -1
	if takesFD {
		if len(ctx.pendingFDs) == 0 {
			return fmt.Errorf("missing descriptor for object %d opcode %d", senderID, opcode)
		}
		fd, ctx.pendingFDs = ctx.pendingFDs[0], ctx.pendingFDs[1:]
	}
	if dispatcher, ok := sender.(Dispatcher); ok {
		dispatcher.Dispatch(opcode, fd, data)
		return nil
	}
	if fd >= 0 {
		_ = unix.Close(fd)
	}
	return fmt.Errorf("%w (senderID=%d)", ErrDispatchSenderUnsupported, senderID)
}

func Connect(addr string) (*Display, error) {
	if addr == "" {
		runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
		if runtimeDir == "" {
			return nil, errors.New("env XDG_RUNTIME_DIR not set")
		}
		if addr == "" {
			addr = os.Getenv("WAYLAND_DISPLAY")
		}
		if addr == "" {
			addr = "wayland-0"
		}
		addr = runtimeDir + "/" + addr
	}

	ctx := &Context{
		objects: map[uint32]Proxy{},
	}

	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: addr, Net: "unix"})
	if err != nil {
		return nil, err
	}
	ctx.conn = conn

	return NewDisplay(ctx), nil
}
