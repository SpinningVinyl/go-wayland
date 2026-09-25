package client

import (
	"bytes"
	"fmt"
	"io"
	"unsafe"

	"golang.org/x/sys/unix"
)

// ReadMsg reads a complete wire message and every descriptor received along the way.
// Descriptors form a separate ordered stream: they need not accompany their event.
// Only this reader touches conn; DispatchMessage consumes the FD queue on the UI thread.
func (ctx *Context) ReadMsg() (senderID uint32, opcode uint32, fds []int, msg []byte, err error) {
	defer func() {
		if err != nil {
			for _, fd := range fds {
				_ = unix.Close(fd)
			}
			fds = nil
		}
	}()
	readFull := func(dst []byte) error {
		// Linux allows up to 253 SCM_RIGHTS descriptors in one sendmsg.
		oob := make([]byte, unix.CmsgSpace(253*4))
		for len(dst) > 0 {
			n, oobn, flags, _, readErr := ctx.conn.ReadMsgUnix(dst, oob)
			if oobn > 0 {
				messages, parseErr := unix.ParseSocketControlMessage(oob[:oobn])
				if parseErr != nil {
					return parseErr
				}
				for _, message := range messages {
					rights, parseErr := unix.ParseUnixRights(&message)
					if parseErr != nil {
						return parseErr
					}
					fds = append(fds, rights...)
				}
			}
			if flags&unix.MSG_CTRUNC != 0 {
				return fmt.Errorf("truncated Wayland file descriptors")
			}
			if readErr != nil {
				return readErr
			}
			if n == 0 {
				return io.EOF
			}
			dst = dst[n:]
		}
		return nil
	}
	header := make([]byte, 8)
	if err = readFull(header); err != nil {
		return
	}
	senderID = Uint32(header[:4])
	opcodeAndSize := Uint32(header[4:])
	opcode = opcodeAndSize & 0xffff
	size := int(opcodeAndSize >> 16)
	if size < 8 || size%4 != 0 {
		err = fmt.Errorf("invalid Wayland message size %d", size)
		return
	}
	msg = make([]byte, size-8)
	err = readFull(msg)
	return
}

func Uint32(src []byte) uint32 {
	_ = src[3]
	return *(*uint32)(unsafe.Pointer(&src[0]))
}

func String(src []byte) string {
	idx := bytes.IndexByte(src, 0)
	src = src[:idx:idx]
	return *(*string)(unsafe.Pointer(&src))
}

func Fixed(src []byte) float64 {
	_ = src[3]
	fx := *(*int32)(unsafe.Pointer(&src[0]))
	return fixedToFloat64(fx)
}
