package system

import (
	"strings"
	"testing"
)

func TestParseDefaultRoute(t *testing.T) {
	t.Parallel()

	input := `Iface Destination Gateway Flags RefCnt Use Metric Mask MTU Window IRTT
eth0 0000FEA9 00000000 0001 0 0 0 00FFFFFF 0 0 0
wlan0 00000000 0102A8C0 0003 0 0 100 00000000 0 0 0
eth1 00000000 0100000A 0000 0 0 200 00000000 0 0 0
`
	got, err := parseDefaultRoute(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if got != "wlan0" {
		t.Fatalf("default route = %q, want wlan0", got)
	}
}

func TestParseDefaultRouteRejectsMissingRoute(t *testing.T) {
	t.Parallel()

	if _, err := parseDefaultRoute(strings.NewReader("Iface Destination Gateway Flags\n")); err == nil {
		t.Fatal("parseDefaultRoute returned no error")
	}
}
