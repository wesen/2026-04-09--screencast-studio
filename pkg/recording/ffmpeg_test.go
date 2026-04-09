package recording

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

func TestBuildVideoRecordArgsForRegion(t *testing.T) {
	job := dsl.VideoJob{
		Source: dsl.EffectiveVideoSource{
			ID:   "top_left_region",
			Name: "Top Left Region",
			Type: "region",
			Target: dsl.VideoTarget{
				Display: ":0.0",
				Rect: &dsl.Rect{
					X: 0,
					Y: 0,
					W: 1280,
					H: 720,
				},
			},
			Capture: dsl.VideoCaptureSettings{FPS: 24},
			Output: dsl.VideoOutputSettings{
				Container:  "mov",
				VideoCodec: "h264",
				Quality:    75,
			},
		},
		OutputPath: "./recordings/demo/top-left.mov",
	}

	args, err := buildVideoRecordArgs(job, 0)
	require.NoError(t, err)
	require.Contains(t, args, "-f")
	require.Contains(t, args, "x11grab")
	require.Contains(t, args, "-video_size")
	require.Contains(t, args, "1280x720")
	require.Contains(t, args, ":0.0+0,0")
	require.Equal(t, "./recordings/demo/top-left.mov", args[len(args)-1])
}

func TestBuildAudioMixArgs(t *testing.T) {
	job := dsl.AudioMixJob{
		Name: "audio-mix",
		Sources: []dsl.EffectiveAudioSource{
			{
				ID:      "mic",
				Name:    "Mic",
				Enabled: true,
				Device:  "default",
				Settings: dsl.AudioSourceSettings{
					Gain: 1.25,
				},
			},
		},
		Output: dsl.AudioOutputSettings{
			Codec:        "pcm_s16le",
			SampleRateHz: 48000,
			Channels:     2,
		},
		OutputPath: "./recordings/demo/audio-mix.wav",
	}

	args, err := buildAudioMixArgs(job, 0)
	require.NoError(t, err)
	require.Contains(t, args, "pulse")
	require.Contains(t, args, "default")
	require.Contains(t, args, "[0:a]volume=1.25[a0];[a0]anull[aout]")
	require.Equal(t, "./recordings/demo/audio-mix.wav", args[len(args)-1])
}
