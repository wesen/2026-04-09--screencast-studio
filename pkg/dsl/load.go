package dsl

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func LoadFile(path string) ([]byte, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "read DSL file")
	}
	return body, nil
}

func decodeDSL(body []byte, out any) error {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return errors.New("empty config")
	}
	if json.Valid([]byte(trimmed)) {
		if err := json.Unmarshal([]byte(trimmed), out); err != nil {
			return errors.Wrap(err, "decode JSON DSL")
		}
		return nil
	}
	if err := yaml.Unmarshal([]byte(trimmed), out); err != nil {
		return errors.Wrap(err, "decode YAML DSL")
	}
	return nil
}
