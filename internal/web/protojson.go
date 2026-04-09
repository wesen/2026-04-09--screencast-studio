package web

import (
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var protoJSONMarshalOptions = protojson.MarshalOptions{
	EmitUnpopulated: true,
}

func decodeProtoJSON(w http.ResponseWriter, r *http.Request, msg proto.Message) bool {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return false
	}
	if err := protojson.Unmarshal(body, msg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return false
	}
	return true
}

func marshalProtoJSON(msg proto.Message) ([]byte, error) {
	return protoJSONMarshalOptions.Marshal(msg)
}

func writeProtoJSON(w http.ResponseWriter, statusCode int, msg proto.Message) {
	body, err := marshalProtoJSON(msg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "protojson_encode_failed", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(body)
	_, _ = w.Write([]byte("\n"))
}

func writeProtoWebsocketJSON(conn *websocket.Conn, msg proto.Message) error {
	body, err := marshalProtoJSON(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, body)
}
