package client

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestPutStringUsesUnpaddedLength(t *testing.T) {
	for _, value := range []string{"x", "wl_shm", "text/uri-list"} {
		wire := make([]byte, 4+PaddedLen(len(value)+1))
		PutString(wire, value, len(wire)-4)
		if got := Uint32(wire); got != uint32(len(value)+1) {
			t.Fatalf("%q: wire length %d, want %d", value, got, len(value)+1)
		}
		if wire[4+len(value)] != 0 {
			t.Fatalf("%q has no NUL terminator", value)
		}
	}
}

func TestServerCreatedDataOffer(t *testing.T) {
	display, server := testWaylandPair(t)
	ctx := display.Context()
	device := NewDataDevice(ctx)
	initialID := ctx.currentID
	const serverID = 0xff000010
	message := make([]byte, 12)
	PutUint32(message, device.ID())
	PutUint32(message[4:], 12<<16)
	PutUint32(message[8:], serverID)
	if _, err := server.Write(message); err != nil {
		t.Fatal(err)
	}
	if err := ctx.Dispatch(); err != nil {
		t.Fatal(err)
	}
	if _, ok := ctx.GetProxy(serverID).(*DataOffer); !ok {
		t.Fatal("new data offer not registered")
	}
	if ctx.currentID != initialID {
		t.Fatal("server ID advanced client ID counter")
	}
}

func TestDescriptorsFollowEventTypes(t *testing.T) {
	display, server := testWaylandPair(t)
	ctx := display.Context()
	source := NewDataSource(ctx)
	source.SetSendHandler(func(e DataSourceSendEvent) {
		if e.MimeType != "text/uri-list" {
			t.Errorf("MIME: %q", e.MimeType)
		}
		fd := os.NewFile(uintptr(e.Fd), "transfer")
		if _, err := fd.WriteString("ok"); err != nil {
			t.Error(err)
		}
		if err := fd.Close(); err != nil {
			t.Error(err)
		}
	})
	var readers, writers []*os.File
	var fds []int
	for i := 0; i < 2; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		if err := r.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
			t.Fatal(err)
		}
		readers = append(readers, r)
		writers = append(writers, w)
		fds = append(fds, int(w.Fd()))
	}
	target := make([]byte, 12)
	PutUint32(target, source.ID())
	PutUint32(target[4:], 12<<16)
	send := make([]byte, 28)
	PutUint32(send, source.ID())
	PutUint32(send[4:], 28<<16|1)
	PutString(send[8:], "text/uri-list", 16)
	batch := append(append(target, send...), send...)
	if _, _, err := server.WriteMsgUnix(batch, unix.UnixRights(fds...), nil); err != nil {
		t.Fatal(err)
	}
	for _, w := range writers {
		_ = w.Close()
	}
	for i := 0; i < 3; i++ {
		if err := ctx.Dispatch(); err != nil {
			t.Fatal(err)
		}
	}
	for _, r := range readers {
		data, err := io.ReadAll(r)
		if err != nil || string(data) != "ok" {
			t.Fatalf("got %q, %v; want ok", data, err)
		}
	}
}

func testWaylandPair(t *testing.T) (*Display, *net.UnixConn) {
	t.Helper()
	address := filepath.Join(t.TempDir(), "wayland")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: address, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	display, err := Connect(address)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = display.Context().Close() })
	server, err := listener.AcceptUnix()
	if err != nil {
		t.Fatal(err)
	}
	if err := server.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return display, server
}
