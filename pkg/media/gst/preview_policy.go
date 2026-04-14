package gst

import "github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"

type sharedPreviewLayout string

const (
	sharedPreviewLayoutScaleFirst sharedPreviewLayout = "scale-first"
	sharedPreviewLayoutRateFirst  sharedPreviewLayout = "rate-first"
)

type sharedPreviewMode string

const (
	sharedPreviewModeNormal    sharedPreviewMode = "normal"
	sharedPreviewModeRecording sharedPreviewMode = "recording"
)

type sharedPreviewStage string

const (
	sharedPreviewStageQueue     sharedPreviewStage = "queue"
	sharedPreviewStageScale     sharedPreviewStage = "scale"
	sharedPreviewStageScaleCaps sharedPreviewStage = "scale-caps"
	sharedPreviewStageRate      sharedPreviewStage = "rate"
	sharedPreviewStageRateCaps  sharedPreviewStage = "rate-caps"
	sharedPreviewStageJPEG      sharedPreviewStage = "jpeg"
	sharedPreviewStageSink      sharedPreviewStage = "sink"
)

type sharedPreviewRecipe struct {
	Layout  sharedPreviewLayout
	Profile sharedPreviewProfile
	Stages  []sharedPreviewStage
}

type sharedPreviewProfiles struct {
	Display sharedPreviewProfile
	Window  sharedPreviewProfile
	Region  sharedPreviewProfile
	Camera  sharedPreviewProfile
}

type sharedPreviewPolicy struct {
	Layout            sharedPreviewLayout
	NormalProfiles    sharedPreviewProfiles
	RecordingProfiles sharedPreviewProfiles
}

func defaultSharedPreviewPolicy() sharedPreviewPolicy {
	return sharedPreviewPolicy{
		Layout: sharedPreviewLayoutRateFirst,
		NormalProfiles: sharedPreviewProfiles{
			Display: sharedPreviewProfile{MaxWidth: 960, FPS: 10, JPEGQuality: 75},
			Window:  sharedPreviewProfile{MaxWidth: 1280, FPS: 10, JPEGQuality: 80},
			Region:  sharedPreviewProfile{MaxWidth: 1280, FPS: 10, JPEGQuality: 80},
			Camera:  sharedPreviewProfile{MaxWidth: 1280, FPS: 10, JPEGQuality: 85},
		},
		RecordingProfiles: sharedPreviewProfiles{
			Display: sharedPreviewProfile{MaxWidth: 640, FPS: 4, JPEGQuality: 50},
			Window:  sharedPreviewProfile{MaxWidth: 640, FPS: 4, JPEGQuality: 50},
			Region:  sharedPreviewProfile{MaxWidth: 640, FPS: 4, JPEGQuality: 50},
			Camera:  sharedPreviewProfile{MaxWidth: 960, FPS: 6, JPEGQuality: 70},
		},
	}
}

func (p sharedPreviewPolicy) normalizedLayout() sharedPreviewLayout {
	switch p.Layout {
	case sharedPreviewLayoutRateFirst:
		return sharedPreviewLayoutRateFirst
	default:
		return sharedPreviewLayoutScaleFirst
	}
}

func (p sharedPreviewPolicy) recipeFor(source dsl.EffectiveVideoSource, mode sharedPreviewMode) sharedPreviewRecipe {
	profile := p.profileForSource(source, mode)
	layout := p.normalizedLayout()
	return sharedPreviewRecipe{
		Layout:  layout,
		Profile: profile,
		Stages:  previewStagesForLayout(layout),
	}
}

func (p sharedPreviewPolicy) profileForSource(source dsl.EffectiveVideoSource, mode sharedPreviewMode) sharedPreviewProfile {
	profiles := p.NormalProfiles
	if mode == sharedPreviewModeRecording {
		profiles = p.RecordingProfiles
	}
	switch source.Type {
	case "camera":
		return normalizePreviewProfile(source, profiles.Camera)
	case "window":
		return normalizePreviewProfile(source, profiles.Window)
	case "region":
		return normalizePreviewProfile(source, profiles.Region)
	case "display":
		fallthrough
	default:
		return normalizePreviewProfile(source, profiles.Display)
	}
}

func normalizePreviewProfile(source dsl.EffectiveVideoSource, profile sharedPreviewProfile) sharedPreviewProfile {
	if profile.MaxWidth <= 0 {
		profile.MaxWidth = 960
	}
	profile.FPS = previewFPSForProfile(source.Capture.FPS, profile.FPS)
	if profile.JPEGQuality <= 0 {
		profile.JPEGQuality = 75
	}
	return profile
}

func previewStagesForLayout(layout sharedPreviewLayout) []sharedPreviewStage {
	switch layout {
	case sharedPreviewLayoutRateFirst:
		return []sharedPreviewStage{
			sharedPreviewStageQueue,
			sharedPreviewStageRate,
			sharedPreviewStageRateCaps,
			sharedPreviewStageScale,
			sharedPreviewStageScaleCaps,
			sharedPreviewStageJPEG,
			sharedPreviewStageSink,
		}
	default:
		return []sharedPreviewStage{
			sharedPreviewStageQueue,
			sharedPreviewStageScale,
			sharedPreviewStageScaleCaps,
			sharedPreviewStageRate,
			sharedPreviewStageRateCaps,
			sharedPreviewStageJPEG,
			sharedPreviewStageSink,
		}
	}
}
