package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestGeneratedServerObjectAndFDType(t *testing.T) {
	prefix = "wl"
	defer func() { prefix = "" }()
	protocol = Protocol{Name: "wayland", Interfaces: []Interface{{Name: "wl_data_offer"}}}
	var output bytes.Buffer
	writeEventDispatcher(&output, "DataDevice", Interface{Events: []Event{{Name: "data_offer", Args: []Arg{{Name: "id", Type: "new_id", Interface: "wl_data_offer"}}}}})
	s := output.String()
	for _, want := range []string{"e.Id = new(DataOffer)", "RegisterServer(e.Id", "if i.dataOfferHandler != nil"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
	if strings.Contains(s, "if i.dataOfferHandler == nil") {
		t.Fatal("unhandled new_id was dropped")
	}
	output.Reset()
	writeEventDispatcher(&output, "DataSource", Interface{Events: []Event{{Name: "target"}, {Name: "send", Args: []Arg{{Name: "fd", Type: "fd"}}}}})
	if !strings.Contains(output.String(), "TakesFD") || !strings.Contains(output.String(), "case 1: return true") {
		t.Fatalf("FD event type missing: %s", output.String())
	}
}

func TestCurrentProtocolNames(t *testing.T) {
	if got := toLowerCamel("range"); got != "_range" {
		t.Fatalf("range argument generated as %q", got)
	}
	if got := interfaceGoType("zwp_tablet_tool_v2"); got != "tablet.TabletTool" {
		t.Fatalf("tablet reference generated as %q", got)
	}
}

func TestArrayPaddingAndMultipleFDs(t *testing.T) {
	protocol = Protocol{Name: "test"}
	var output bytes.Buffer
	writeRequest(&output, "Manager", 0, Request{Name: "send", Args: []Arg{
		{Name: "data", Type: "array"},
		{Name: "listen_fd", Type: "fd"},
		{Name: "close_fd", Type: "fd"},
	}})
	s := output.String()
	for _, want := range []string{"client.PaddedLen(len(data))", "(4 + dataLen)", "unix.UnixRights(listenFd, closeFd)"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
	}
}
