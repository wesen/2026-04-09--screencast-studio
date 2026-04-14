package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
)

// This experiment builds a GStreamer preview pipeline in Go:
//   ximagesrc -> videoconvert -> videoscale -> videorate -> jpegenc -> appsink
//
// It mirrors the current FFmpeg preview: capture screen, downscale, reduce FPS,
// produce JPEG frames that Go code can consume.
//
// The appsink callback receives each JPEG buffer and prints its size.
// This proves we can get raw frame bytes from GStreamer into Go.

func buildPreviewPipeline() (*gst.Pipeline, *app.Sink, error) {
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, nil, fmt.Errorf("create pipeline: %w", err)
	}

	// Source: capture X11 screen (top-left 640x480 region)
	ximagesrc, err := gst.NewElement("ximagesrc")
	if err != nil {
		return nil, nil, fmt.Errorf("create ximagesrc: %w", err)
	}
	ximagesrc.Set("startx", 0)
	ximagesrc.Set("starty", 0)
	ximagesrc.Set("endx", 639)
	ximagesrc.Set("endy", 479)
	ximagesrc.Set("use-damage", false)

	// Convert colorspace
	videoconvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, nil, fmt.Errorf("create videoconvert: %w", err)
	}

	// Scale down
	videoscale, err := gst.NewElement("videoscale")
	if err != nil {
		return nil, nil, fmt.Errorf("create videoscale: %w", err)
	}

	// Caps: 640 wide, keep aspect ratio (GStreamer will compute height)
	capsScale := gst.NewCapsFromString("video/x-raw,width=640")
	capsfilterScale, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, nil, fmt.Errorf("create capsfilter: %w", err)
	}
	capsfilterScale.Set("caps", capsScale)

	// Reduce framerate to 5fps
	videorate, err := gst.NewElement("videorate")
	if err != nil {
		return nil, nil, fmt.Errorf("create videorate: %w", err)
	}
	capsRate := gst.NewCapsFromString("video/x-raw,framerate=5/1")
	capsfilterRate, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, nil, fmt.Errorf("create rate capsfilter: %w", err)
	}
	capsfilterRate.Set("caps", capsRate)

	// Encode to JPEG
	jpegenc, err := gst.NewElement("jpegenc")
	if err != nil {
		return nil, nil, fmt.Errorf("create jpegenc: %w", err)
	}
	jpegenc.Set("quality", 50)

	// AppSink: receive JPEG frames in Go
	sink, err := app.NewAppSink()
	if err != nil {
		return nil, nil, fmt.Errorf("create appsink: %w", err)
	}
	sink.SetCaps(gst.NewCapsFromString("image/jpeg"))
	sink.SetProperty("max-buffers", 2)
	sink.SetProperty("drop", true)

	// Add all elements to pipeline
	pipeline.AddMany(
		ximagesrc, videoconvert, videoscale, capsfilterScale,
		videorate, capsfilterRate, jpegenc, sink.Element,
	)

	// Link: ximagesrc -> videoconvert -> videoscale -> capsfilter -> videorate -> capsfilter -> jpegenc -> appsink
	if err := ximagesrc.Link(videoconvert); err != nil {
		return nil, nil, fmt.Errorf("link ximagesrc->videoconvert: %w", err)
	}
	if err := videoconvert.Link(videoscale); err != nil {
		return nil, nil, fmt.Errorf("link videoconvert->videoscale: %w", err)
	}
	if err := videoscale.Link(capsfilterScale); err != nil {
		return nil, nil, fmt.Errorf("link videoscale->capsfilter: %w", err)
	}
	if err := capsfilterScale.Link(videorate); err != nil {
		return nil, nil, fmt.Errorf("link capsfilter->videorate: %w", err)
	}
	if err := videorate.Link(capsfilterRate); err != nil {
		return nil, nil, fmt.Errorf("link videorate->capsfilter: %w", err)
	}
	if err := capsfilterRate.Link(jpegenc); err != nil {
		return nil, nil, fmt.Errorf("link capsfilter->jpegenc: %w", err)
	}
	if err := jpegenc.Link(sink.Element); err != nil {
		return nil, nil, fmt.Errorf("link jpegenc->appsink: %w", err)
	}

	return pipeline, sink, nil
}

func main() {
	gst.Init(nil)

	fmt.Println("=== go-gst Preview Pipeline Experiment ===")
	fmt.Println("Building pipeline: ximagesrc -> videoconvert -> videoscale -> videorate(5fps) -> jpegenc -> appsink")
	fmt.Println()

	pipeline, sink, err := buildPreviewPipeline()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build pipeline: %v\n", err)
		os.Exit(1)
	}

	frameCount := 0
	totalBytes := 0

	// Wire up the appsink callback — this is where Go receives JPEG frames
	sink.SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
			sample := s.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}

			buffer := sample.GetBuffer()
			if buffer == nil {
				return gst.FlowError
			}

			// Map the buffer to get the raw JPEG bytes
			mapInfo := buffer.Map(gst.MapRead)
			if mapInfo == nil {
				return gst.FlowError
			}
			defer buffer.Unmap()

			frameCount++
			totalBytes += int(mapInfo.Size())
			fmt.Printf("  Frame %04d: %d bytes (JPEG)\n", frameCount, mapInfo.Size())

			return gst.FlowOK
		},
	})

	// Handle bus messages
	bus := pipeline.GetPipelineBus()
	bus.AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageError:
			err := msg.ParseError()
			fmt.Fprintf(os.Stderr, "Pipeline error: %s\n", err.Error())
			if debug := err.DebugString(); debug != "" {
				fmt.Fprintf(os.Stderr, "Debug: %s\n", debug)
			}
			return false
		case gst.MessageEOS:
			fmt.Println("Received EOS — pipeline finished")
			return false
		case gst.MessageStateChanged:
			// We could log state changes here but it's noisy
		}
		return true
	})

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the pipeline
	fmt.Println("Starting pipeline (running for ~5 seconds, or press Ctrl+C)...")
	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start pipeline: %v\n", err)
		os.Exit(1)
	}

	// Run main loop in a goroutine
	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)
	go func() {
		mainLoop.Run()
	}()

	// Wait for signal or timeout
	select {
	case <-sigChan:
		fmt.Println("\nCaught signal, stopping...")
	case <-time.After(5 * time.Second):
		fmt.Println("\n5 second timeout reached, stopping...")
	}

	// Send EOS for clean shutdown
	pipeline.SendEvent(gst.NewEOSEvent())

	// Stop the pipeline
	pipeline.BlockSetState(gst.StateNull)
	mainLoop.Quit()

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Frames captured: %d\n", frameCount)
	fmt.Printf("Total bytes: %d\n", totalBytes)
	if frameCount > 0 {
		fmt.Printf("Average frame size: %d bytes\n", totalBytes/frameCount)
	}
}
