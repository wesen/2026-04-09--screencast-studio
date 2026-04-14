package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
)

// This experiment builds a GStreamer audio recording pipeline in Go:
//   pulsesrc -> audioconvert -> volume(1.5) -> wavenc -> filesink
//
// It captures 3 seconds of audio with gain adjustment and writes to a WAV file.

func buildAudioPipeline(outputPath string) (*gst.Pipeline, error) {
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, fmt.Errorf("create pipeline: %w", err)
	}

	// Source: PulseAudio capture
	pulsesrc, err := gst.NewElement("pulsesrc")
	if err != nil {
		return nil, fmt.Errorf("create pulsesrc: %w", err)
	}
	pulsesrc.Set("device", "default")

	// Request specific format
	capsfilter, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, fmt.Errorf("create capsfilter: %w", err)
	}
	capsfilter.Set("caps", gst.NewCapsFromString("audio/x-raw,rate=48000,channels=2"))

	// Convert format if needed
	audioconvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, fmt.Errorf("create audioconvert: %w", err)
	}

	// Volume/gain element
	volume, err := gst.NewElement("volume")
	if err != nil {
		return nil, fmt.Errorf("create volume: %w", err)
	}
	volume.Set("volume", 1.0)

	// WAV encoder
	wavenc, err := gst.NewElement("wavenc")
	if err != nil {
		return nil, fmt.Errorf("create wavenc: %w", err)
	}

	// File output
	filesink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, fmt.Errorf("create filesink: %w", err)
	}
	filesink.Set("location", outputPath)

	// Add and link
	pipeline.AddMany(pulsesrc, capsfilter, audioconvert, volume, wavenc, filesink)

	if err := pulsesrc.Link(capsfilter); err != nil {
		return nil, fmt.Errorf("link pulsesrc->capsfilter: %w", err)
	}
	if err := capsfilter.Link(audioconvert); err != nil {
		return nil, fmt.Errorf("link capsfilter->audioconvert: %w", err)
	}
	if err := audioconvert.Link(volume); err != nil {
		return nil, fmt.Errorf("link audioconvert->volume: %w", err)
	}
	if err := volume.Link(wavenc); err != nil {
		return nil, fmt.Errorf("link volume->wavenc: %w", err)
	}
	if err := wavenc.Link(filesink); err != nil {
		return nil, fmt.Errorf("link wavenc->filesink: %w", err)
	}

	return pipeline, nil
}

func main() {
	gst.Init(nil)

	outputPath := "/tmp/gst-go-audio-test.wav"
	fmt.Println("=== go-gst Audio Capture Pipeline Experiment ===")
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Println()

	pipeline, err := buildAudioPipeline(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build pipeline: %v\n", err)
		os.Exit(1)
	}

	// Bus messages
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
			fmt.Println("Received EOS — recording finished")
			return false
		}
		return true
	})

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start recording
	fmt.Println("Starting audio recording (3 seconds)...")
	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start pipeline: %v\n", err)
		os.Exit(1)
	}

	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)
	go mainLoop.Run()

	// Wait for signal or timeout
	select {
	case <-sigChan:
		fmt.Println("\nCaught signal, stopping...")
	case <-time.After(3 * time.Second):
		fmt.Println("3 second recording complete, sending EOS...")
	}

	// Send EOS for clean WAV finalization
	pipeline.SendEvent(gst.NewEOSEvent())

	// Wait for EOS to be processed (bus watch will quit the loop)
	time.Sleep(500 * time.Millisecond)

	// Stop
	pipeline.BlockSetState(gst.StateNull)
	mainLoop.Quit()

	// Check output
	stat, err := os.Stat(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Output file not found: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Output: %s (%d bytes)\n", outputPath, stat.Size())
}
