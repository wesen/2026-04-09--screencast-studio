package dsl

type Config struct {
	Schema                string            `json:"schema" yaml:"schema"`
	SessionID             string            `json:"session_id" yaml:"session_id"`
	DestinationTemplates  map[string]string `json:"destination_templates" yaml:"destination_templates"`
	ScreenCaptureDefaults VideoDefaults     `json:"screen_capture_defaults" yaml:"screen_capture_defaults"`
	CameraCaptureDefaults VideoDefaults     `json:"camera_capture_defaults" yaml:"camera_capture_defaults"`
	AudioDefaults         AudioDefaults     `json:"audio_defaults" yaml:"audio_defaults"`
	AudioMix              AudioMixConfig    `json:"audio_mix" yaml:"audio_mix"`
	VideoSources          []VideoSource     `json:"video_sources" yaml:"video_sources"`
	AudioSources          []AudioSource     `json:"audio_sources" yaml:"audio_sources"`
}

type VideoDefaults struct {
	Capture VideoCaptureSettings `json:"capture" yaml:"capture"`
	Output  VideoOutputSettings  `json:"output" yaml:"output"`
}

type VideoCaptureSettings struct {
	FPS          int    `json:"fps" yaml:"fps"`
	Cursor       *bool  `json:"cursor" yaml:"cursor"`
	FollowResize *bool  `json:"follow_resize" yaml:"follow_resize"`
	Mirror       *bool  `json:"mirror" yaml:"mirror"`
	Size         string `json:"size" yaml:"size"`
}

type VideoOutputSettings struct {
	Container  string `json:"container" yaml:"container"`
	VideoCodec string `json:"video_codec" yaml:"video_codec"`
	Quality    int    `json:"quality" yaml:"quality"`
}

type VideoSource struct {
	ID                  string        `json:"id" yaml:"id"`
	Name                string        `json:"name" yaml:"name"`
	Type                string        `json:"type" yaml:"type"`
	Enabled             *bool         `json:"enabled" yaml:"enabled"`
	Target              VideoTarget   `json:"target" yaml:"target"`
	Settings            VideoDefaults `json:"settings" yaml:"settings"`
	DestinationTemplate string        `json:"destination_template" yaml:"destination_template"`
}

type VideoTarget struct {
	Display  string `json:"display" yaml:"display"`
	WindowID string `json:"window_id" yaml:"window_id"`
	Device   string `json:"device" yaml:"device"`
	Rect     *Rect  `json:"rect" yaml:"rect"`
}

type Rect struct {
	X int `json:"x" yaml:"x"`
	Y int `json:"y" yaml:"y"`
	W int `json:"w" yaml:"w"`
	H int `json:"h" yaml:"h"`
}

type AudioDefaults struct {
	Output AudioOutputSettings `json:"output" yaml:"output"`
}

type AudioOutputSettings struct {
	Codec        string `json:"codec" yaml:"codec"`
	SampleRateHz int    `json:"sample_rate_hz" yaml:"sample_rate_hz"`
	Channels     int    `json:"channels" yaml:"channels"`
}

type AudioMixConfig struct {
	DestinationTemplate string `json:"destination_template" yaml:"destination_template"`
}

type AudioSource struct {
	ID       string              `json:"id" yaml:"id"`
	Name     string              `json:"name" yaml:"name"`
	Device   string              `json:"device" yaml:"device"`
	Enabled  *bool               `json:"enabled" yaml:"enabled"`
	Settings AudioSourceSettings `json:"settings" yaml:"settings"`
}

type AudioSourceSettings struct {
	Gain      float64 `json:"gain" yaml:"gain"`
	NoiseGate bool    `json:"noise_gate" yaml:"noise_gate"`
	Denoise   bool    `json:"denoise" yaml:"denoise"`
}

type EffectiveConfig struct {
	Schema               string
	SessionID            string
	DestinationTemplates map[string]string
	AudioMixTemplate     string
	AudioOutput          AudioOutputSettings
	VideoSources         []EffectiveVideoSource
	AudioSources         []EffectiveAudioSource
	RawDSL               string
	Warnings             []string
}

type EffectiveVideoSource struct {
	ID                  string
	Name                string
	Type                string
	Enabled             bool
	Target              VideoTarget
	Capture             VideoCaptureSettings
	Output              VideoOutputSettings
	DestinationTemplate string
}

type EffectiveAudioSource struct {
	ID       string
	Name     string
	Enabled  bool
	Device   string
	Settings AudioSourceSettings
}

type PlannedOutput struct {
	Kind     string `json:"kind"`
	SourceID string `json:"source_id,omitempty"`
	Name     string `json:"name"`
	Path     string `json:"path"`
}

type CompiledPlan struct {
	SessionID string          `json:"session_id"`
	Outputs   []PlannedOutput `json:"outputs"`
	Warnings  []string        `json:"warnings"`
}
