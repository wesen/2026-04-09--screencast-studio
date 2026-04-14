#!/usr/bin/env bash
# 05-go-gst-basic-test.sh
# Build and run a minimal go-gst program to verify native bindings work.
# This creates a temporary Go module outside the main project.
set -euo pipefail

WORKDIR="/tmp/gst-go-experiment"
rm -rf "$WORKDIR"
mkdir -p "$WORKDIR"
cd "$WORKDIR"

echo "=== Setting up Go module ==="
go mod init gst-experiment
go get github.com/go-gst/go-gst@latest
go get github.com/go-gst/go-glib@latest
go get github.com/mattn/go-pointer@latest

cat > main.go << 'GOEOF'
package main

import (
	"fmt"
	"os"

	"github.com/go-gst/go-gst/gst"
)

func main() {
	// Initialize GStreamer
	gst.Init(nil)

	// Print version
	major, minor, micro, nano := gst.Version()
	fmt.Printf("GStreamer version: %d.%d.%d.%d\n", major, minor, micro, nano)

	// List all elements matching "ximagesrc"
	fmt.Println("\n=== ximagesrc details ===")
	factory := gst.FindElementFactory("ximagesrc")
	if factory == nil {
		fmt.Println("ximagesrc factory not found!")
		os.Exit(1)
	}
	fmt.Printf("Element: %s\n", factory.GetMetadata("long-name"))
	fmt.Printf("Description: %s\n", factory.GetMetadata("description"))
	fmt.Printf("Klass: %s\n", factory.GetMetadata("klass"))

	// List pad templates
	fmt.Println("\nPad templates:")
	for _, pad := range factory.GetStaticPadTemplates() {
		fmt.Printf("  %s: %s (direction: %s)\n", pad.Name, pad.Caps.String(), pad.Direction)
	}

	// List properties
	fmt.Println("\nProperties:")
	for _, prop := range factory.ListProperties() {
		fmt.Printf("  %-20s : %s\n", prop.Name, prop.Blurb)
	}

	// Check appsink
	fmt.Println("\n=== appsink details ===")
	appFactory := gst.FindElementFactory("appsink")
	if appFactory != nil {
		fmt.Printf("Element: %s\n", appFactory.GetMetadata("long-name"))
	} else {
		fmt.Println("appsink not found!")
	}

	// Check pulsesrc
	fmt.Println("\n=== pulsesrc details ===")
	pulseFactory := gst.FindElementFactory("pulsesrc")
	if pulseFactory != nil {
		fmt.Printf("Element: %s\n", pulseFactory.GetMetadata("long-name"))
	} else {
		fmt.Println("pulsesrc not found!")
	}
}
GOEOF

echo ""
echo "=== Building ==="
go build -o gst-test . 2>&1

echo ""
echo "=== Running ==="
GST_DEBUG=2 ./gst-test 2>&1 | head -50
