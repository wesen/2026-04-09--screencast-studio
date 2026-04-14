package discovery

import "testing"

func TestDedupeCamerasPrefersLowestVideoNodePerCard(t *testing.T) {
	cameras := []Camera{
		{ID: "/dev/video2", Device: "/dev/video2", Label: "Laptop Camera", CardName: "Laptop Camera"},
		{ID: "/dev/video0", Device: "/dev/video0", Label: "Laptop Camera", CardName: "Laptop Camera"},
		{ID: "/dev/video4", Device: "/dev/video4", Label: "USB Camera", CardName: "USB Camera"},
		{ID: "/dev/video6", Device: "/dev/video6", Label: "USB Camera", CardName: "USB Camera"},
	}

	got := dedupeCameras(cameras)
	if len(got) != 2 {
		t.Fatalf("expected 2 deduped cameras, got %d: %+v", len(got), got)
	}
	if got[0].Device != "/dev/video0" {
		t.Fatalf("expected first deduped camera to prefer /dev/video0, got %q", got[0].Device)
	}
	if got[1].Device != "/dev/video4" {
		t.Fatalf("expected second deduped camera to prefer /dev/video4, got %q", got[1].Device)
	}
}

func TestVideoDeviceIndex(t *testing.T) {
	index, ok := videoDeviceIndex("/dev/video7")
	if !ok || index != 7 {
		t.Fatalf("expected /dev/video7 -> (7, true), got (%d, %v)", index, ok)
	}
	if _, ok := videoDeviceIndex("/dev/not-video"); ok {
		t.Fatalf("expected non-video path to fail parsing")
	}
}
