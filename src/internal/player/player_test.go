package player

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestControllerPlayAndEvents(t *testing.T) {
	socketPath := filepath.Join(os.TempDir(), "tunez-player-test.sock")
	_ = os.Remove(socketPath)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, _ := ln.Accept()
		accepted <- conn
	}()

	ctrl := New(Options{
		MPVPath:        "mpv",
		IPCPath:        socketPath,
		DisableProcess: true,
	})
	if err := ctrl.Start(context.Background()); err != nil {
		t.Fatalf("start controller: %v", err)
	}
	conn := <-accepted
	defer conn.Close()

	if err := ctrl.Play("file:///tmp/test.mp3", nil); err != nil {
		t.Fatalf("play: %v", err)
	}

	go func() {
		evt := map[string]any{"event": "property-change", "name": "time-pos", "data": 12.5}
		b, _ := json.Marshal(evt)
		conn.Write(append(b, '\n'))
		end := map[string]any{"event": "end-file"}
		b, _ = json.Marshal(end)
		conn.Write(append(b, '\n'))
	}()

	timeout := time.After(2 * time.Second)
	receivedPos := false
	receivedEnd := false
loop:
	for {
		select {
		case evt := <-ctrl.Events():
			if evt.Err != nil {
				t.Fatalf("event err: %v", evt.Err)
			}
			if evt.TimePos == 12.5 {
				receivedPos = true
			}
			if evt.Ended {
				receivedEnd = true
				break loop
			}
		case <-timeout:
			t.Fatalf("timeout waiting for events")
		}
	}
	if !receivedPos || !receivedEnd {
		t.Fatalf("expected time-pos and end-file events, got pos=%v end=%v", receivedPos, receivedEnd)
	}
}
