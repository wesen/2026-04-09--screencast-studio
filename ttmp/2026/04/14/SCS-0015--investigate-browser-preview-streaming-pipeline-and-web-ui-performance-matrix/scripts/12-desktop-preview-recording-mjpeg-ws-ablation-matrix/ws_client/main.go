package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

type summary struct {
	URL          string         `json:"url"`
	StartedAt    time.Time      `json:"startedAt"`
	FinishedAt   time.Time      `json:"finishedAt"`
	MessageCount int            `json:"messageCount"`
	ByteCount    int            `json:"byteCount"`
	ByKind       map[string]int `json:"byKind"`
	Error        string         `json:"error,omitempty"`
}

func main() {
	url := flag.String("url", "ws://127.0.0.1:7777/ws", "websocket url")
	out := flag.String("output", "", "output json path")
	flag.Parse()
	if strings.TrimSpace(*out) == "" {
		panic("--output is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := summary{URL: *url, StartedAt: time.Now(), ByKind: map[string]int{}}
	defer func() {
		s.FinishedAt = time.Now()
		b, _ := json.MarshalIndent(s, "", "  ")
		_ = os.WriteFile(*out, b, 0o644)
	}()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, *url, nil)
	if err != nil {
		s.Error = err.Error()
		return
	}
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			s.Error = err.Error()
			return
		}
		s.MessageCount++
		s.ByteCount += len(msg)
		var parsed map[string]any
		if err := json.Unmarshal(msg, &parsed); err == nil {
			kind := "unknown"
			for k := range parsed {
				if k == "timestamp" {
					continue
				}
				kind = k
				break
			}
			s.ByKind[kind]++
		} else {
			s.ByKind["unparsed"]++
		}
	}
}
