package discovery

import (
	"net"
	"testing"
	"time"
)

func TestDeviceMerge(t *testing.T) {
	base := Device{
		IP:           net.ParseIP("10.0.0.1"),
		DisplayName:  "host",
		Manufacturer: "",
		Services:     map[string]int{"svc": 1},
		Sources:      map[string]struct{}{"a": {}},
		ExtraData:    map[string]string{"k1": "v1"},
		FirstSeen:    time.Unix(100, 0),
		LastSeen:     time.Unix(200, 0),
	}

	other := Device{
		IP:           net.ParseIP("10.0.0.1"),
		MAC:          "aa:bb",
		DisplayName:  "new-host",
		Manufacturer: "manu",
		Services:     map[string]int{"svc": 0, "svc2": 2},
		Sources:      map[string]struct{}{"b": {}},
		ExtraData:    map[string]string{"k2": "v2"},
		FirstSeen:    time.Unix(50, 0),
		LastSeen:     time.Unix(300, 0),
	}

	base.Merge(&other)

	if base.MAC != "aa:bb" {
		t.Fatalf("expected MAC merged, got %s", base.MAC)
	}
	if base.DisplayName != "host" {
		t.Fatalf("DisplayName should remain original when non-empty, got %s", base.DisplayName)
	}
	if base.Manufacturer != "manu" {
		t.Fatalf("Manufacturer merge failed, got %s", base.Manufacturer)
	}
	if base.Services["svc"] != 1 || base.Services["svc2"] != 2 {
		t.Fatalf("services merge failed: %+v", base.Services)
	}
	if _, ok := base.Sources["a"]; !ok {
		t.Fatalf("source a missing")
	}
	if _, ok := base.Sources["b"]; !ok {
		t.Fatalf("source b missing")
	}
	if base.ExtraData["k1"] != "v1" || base.ExtraData["k2"] != "v2" {
		t.Fatalf("extra data merge failed: %+v", base.ExtraData)
	}
	if !base.FirstSeen.Equal(time.Unix(50, 0)) {
		t.Fatalf("FirstSeen should be earliest, got %v", base.FirstSeen)
	}
	if !base.LastSeen.Equal(time.Unix(300, 0)) {
		t.Fatalf("LastSeen should be latest, got %v", base.LastSeen)
	}
}

func TestDeviceMergeNilOther(t *testing.T) {
	d := Device{}
	d.Merge(nil)
}
