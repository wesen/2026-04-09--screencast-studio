package web

import (
	"net/http"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

var websocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	events, unsubscribe := s.events.Subscribe(64)
	defer unsubscribe()

	if err := conn.WriteJSON(ServerEvent{
		Type:    "session.state",
		Payload: mapRecordingSessionResponse(s.recordings.Current()),
	}); err != nil {
		return
	}
	if err := conn.WriteJSON(ServerEvent{
		Type:    "preview.list",
		Payload: mapPreviewListPayload(s.previews.List()),
	}); err != nil {
		return
	}

	group, ctx := errgroup.WithContext(r.Context())
	group.Go(func() error {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				return nil
			}
		}
	})
	group.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case event, ok := <-events:
				if !ok {
					return nil
				}
				if err := conn.WriteJSON(event); err != nil {
					return err
				}
			}
		}
	})

	_ = group.Wait()
}

func mapPreviewListPayload(previews []previewSnapshot) apiPreviewListResponse {
	response := apiPreviewListResponse{
		Previews: make([]apiPreviewResponse, 0, len(previews)),
	}
	for _, preview := range previews {
		response.Previews = append(response.Previews, mapPreviewResponse(preview))
	}
	return response
}
