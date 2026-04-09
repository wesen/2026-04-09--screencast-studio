package dsl

import (
	"time"

	"github.com/pkg/errors"
)

func BuildPlan(cfg *EffectiveConfig, now time.Time) (*CompiledPlan, error) {
	outputs := []PlannedOutput{}

	for _, src := range cfg.VideoSources {
		if !src.Enabled {
			continue
		}
		ext := videoExtension(src.Output.Container)
		outPath, err := renderDestination(cfg.DestinationTemplates[src.DestinationTemplate], renderVars{
			SessionID:  cfg.SessionID,
			SourceID:   src.ID,
			SourceName: src.Name,
			SourceType: src.Type,
			Ext:        ext,
			Now:        now,
		})
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, PlannedOutput{
			Kind:     "video",
			SourceID: src.ID,
			Name:     src.Name,
			Path:     outPath,
		})
	}

	enabledAudio := make([]EffectiveAudioSource, 0, len(cfg.AudioSources))
	for _, src := range cfg.AudioSources {
		if src.Enabled {
			enabledAudio = append(enabledAudio, src)
		}
	}

	if len(enabledAudio) > 0 {
		templateName := cfg.AudioMixTemplate
		if templateName == "" {
			if _, ok := cfg.DestinationTemplates["audio_mix"]; ok {
				templateName = "audio_mix"
			} else {
				for k := range cfg.DestinationTemplates {
					templateName = k
					break
				}
			}
		}
		if templateName == "" {
			return nil, errors.New("no destination template available for mixed audio")
		}

		outPath, err := renderDestination(cfg.DestinationTemplates[templateName], renderVars{
			SessionID:  cfg.SessionID,
			SourceID:   "audio_mix",
			SourceName: "audio-mix",
			SourceType: "audio",
			Ext:        audioExtension(cfg.AudioOutput.Codec),
			Now:        now,
		})
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, PlannedOutput{
			Kind: "audio",
			Name: "audio-mix",
			Path: outPath,
		})
	}

	if len(outputs) == 0 {
		return nil, errors.New("no enabled sources to compile")
	}

	return &CompiledPlan{
		SessionID: cfg.SessionID,
		Outputs:   outputs,
		Warnings:  append([]string(nil), cfg.Warnings...),
	}, nil
}
