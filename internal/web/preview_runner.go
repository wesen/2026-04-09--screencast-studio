package web

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"

	"github.com/rs/zerolog/log"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

func computePreviewSignature(source dsl.EffectiveVideoSource) string {
	payload, err := json.Marshal(source)
	if err != nil {
		log.Error().
			Str("event", "preview.signature.marshal.error").
			Str("source_id", source.ID).
			Err(err).
			Msg("failed to marshal preview source for signature; falling back to source id")
		payload = []byte(source.ID)
	}
	sum := sha1.Sum(payload)
	return hex.EncodeToString(sum[:])
}
