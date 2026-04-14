package main

import (
	"fmt"

	"github.com/go-gst/go-gst/gst"
)

func main() {
	// Initialize GStreamer
	gst.Init(nil)

	// Check element factories we need for screencast-studio
	elements := []string{
		// Video sources
		"ximagesrc",
		"pipewiresrc",
		"v4l2src",
		// Audio sources
		"pulsesrc",
		// Sinks
		"appsink",
		"filesink",
		// Video processing
		"videoconvert",
		"videoscale",
		"videorate",
		// Audio processing
		"audioconvert",
		"audioresample",
		"audiomixer",
		"volume",
		// Encoders
		"jpegenc",
		"pngenc",
		"x264enc",
		"opusenc",
		"wavenc",
		// Muxers
		"mp4mux",
		"oggmux",
	}

	fmt.Println("=== Element Factory Availability ===")
	for _, name := range elements {
		factory := gst.Find(name)
		if factory != nil {
			desc := factory.GetMetadata("description")
			fmt.Printf("  ✓ %-18s : %s\n", name, desc)
		} else {
			fmt.Printf("  ✗ %-18s : NOT FOUND\n", name)
		}
	}

	// Detailed ximagesrc info
	fmt.Println("\n=== ximagesrc details ===")
	ximgFactory := gst.Find("ximagesrc")
	if ximgFactory != nil {
		fmt.Printf("Element: %s\n", ximgFactory.GetMetadata("long-name"))
		fmt.Printf("Description: %s\n", ximgFactory.GetMetadata("description"))
		fmt.Printf("Klass: %s\n", ximgFactory.GetMetadata("klass"))
		fmt.Printf("Author: %s\n", ximgFactory.GetMetadata("author"))

		// Pad templates
		elem, err := gst.NewElement("ximagesrc")
		if err == nil {
			for _, pt := range elem.GetPadTemplates() {
				fmt.Printf("  Pad: %s (direction: %v, presence: %v)\n", pt.Name(), pt.Direction(), pt.Presence())
			}
		}
	}

	// appsink details
	fmt.Println("\n=== appsink details ===")
	appFactory := gst.Find("appsink")
	if appFactory != nil {
		fmt.Printf("Element: %s\n", appFactory.GetMetadata("long-name"))
		fmt.Printf("Description: %s\n", appFactory.GetMetadata("description"))
	}

	// audiomixer details
	fmt.Println("\n=== audiomixer details ===")
	mixerFactory := gst.Find("audiomixer")
	if mixerFactory != nil {
		fmt.Printf("Element: %s\n", mixerFactory.GetMetadata("long-name"))
		fmt.Printf("Description: %s\n", mixerFactory.GetMetadata("description"))
	}
}
