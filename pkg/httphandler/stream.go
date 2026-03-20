package httphandler

import (
	"context"
	"encoding/binary"
	"net/http"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	pkg "github.com/mutablelogic/go-whisper/pkg"
	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	websocket "github.com/coder/websocket"
	wsjson "github.com/coder/websocket/wsjson"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RegisterStreamHandlers registers the WebSocket handler for streaming audio transcription
func RegisterStreamHandlers(router *http.ServeMux, prefix string, manager *pkg.Manager, middleware HTTPMiddlewareFuncs) {
	// Do NOT wrap with middleware — it may interfere with the WebSocket upgrade
	router.HandleFunc(joinPath(prefix, "stream"), func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
			return
		}
		_ = streamHandle(w, r, manager)
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// streamHandle upgrades the connection to WebSocket and runs the streaming loop
func streamHandle(w http.ResponseWriter, r *http.Request, manager *pkg.Manager) error {
	// Upgrade to WebSocket
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		return err
	}

	// Use a cancellable context derived from the request context
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Read first text message as StreamConfig JSON
	var cfg schema.StreamConfig
	if err := wsjson.Read(ctx, conn, &cfg); err != nil {
		conn.Close(websocket.StatusUnsupportedData, "expected StreamConfig JSON as first message")
		return err
	}

	// Create a new stream session from the manager
	session, cleanup, err := manager.NewStreamSession(ctx, &cfg)
	if err != nil {
		_ = wsjson.Write(ctx, conn, schema.StreamEvent{Type: schema.StreamEventError, Error: err.Error()})
		conn.Close(websocket.StatusInternalError, err.Error())
		return err
	}
	defer cleanup()

	// Notify client that we are ready
	if err := wsjson.Write(ctx, conn, schema.StreamEvent{Type: schema.StreamEventReady}); err != nil {
		conn.Close(websocket.StatusInternalError, err.Error())
		return err
	}

	// Main loop: read binary PCM frames from client
	for {
		msgType, data, err := conn.Read(ctx)
		if err != nil {
			// Client closed or error — run final processing
			break
		}

		if msgType != websocket.MessageBinary {
			// Unexpected message type — skip
			continue
		}

		// Convert 16-bit signed little-endian PCM to float32
		samples := make([]float32, len(data)/2)
		for i := range samples {
			raw := binary.LittleEndian.Uint16(data[i*2 : i*2+2])
			samples[i] = float32(int16(raw)) / 32768.0
		}

		session.AddSamples(samples)

		if session.ShouldProcess() {
			segments, err := session.Process(ctx)
			if err != nil {
				_ = wsjson.Write(ctx, conn, schema.StreamEvent{Type: schema.StreamEventError, Error: err.Error()})
				conn.Close(websocket.StatusInternalError, err.Error())
				return err
			}
			for _, seg := range segments {
				if err := wsjson.Write(ctx, conn, schema.StreamEvent{Type: schema.StreamEventSegment, Segment: seg}); err != nil {
					conn.Close(websocket.StatusInternalError, err.Error())
					return err
				}
			}
		}
	}

	// Process any remaining audio
	segments, err := session.ProcessFinal(ctx)
	if err != nil {
		_ = wsjson.Write(ctx, conn, schema.StreamEvent{Type: schema.StreamEventError, Error: err.Error()})
		conn.Close(websocket.StatusInternalError, err.Error())
		return err
	}
	for _, seg := range segments {
		if err := wsjson.Write(ctx, conn, schema.StreamEvent{Type: schema.StreamEventSegment, Segment: seg}); err != nil {
			conn.Close(websocket.StatusInternalError, err.Error())
			return err
		}
	}

	// Send done event and close cleanly
	_ = wsjson.Write(ctx, conn, schema.StreamEvent{Type: schema.StreamEventDone})
	conn.Close(websocket.StatusNormalClosure, "")
	return nil
}
