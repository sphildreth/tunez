package queue

import (
	"fmt"
	"testing"

	"github.com/tunez/tunez/internal/provider"
)

func sampleTracks(n int) []provider.Track {
	var out []provider.Track
	for i := 0; i < n; i++ {
		out = append(out, provider.Track{ID: fmt.Sprintf("t%d", i), Title: fmt.Sprintf("Track %d", i)})
	}
	return out
}

func TestQueueAddAndCurrent(t *testing.T) {
	q := New()
	tracks := []provider.Track{{ID: "1"}, {ID: "2"}}
	q.Add(tracks...)
	if q.Len() != 2 {
		t.Fatalf("expected len 2 got %d", q.Len())
	}
	cur, err := q.Current()
	if err != nil || cur.ID != "1" {
		t.Fatalf("expected first track, got %v err %v", cur, err)
	}
}

func TestQueueNextPrev(t *testing.T) {
	q := New()
	q.Add(sampleTracks(3)...)
	if _, err := q.Next(); err != nil {
		t.Fatalf("next err: %v", err)
	}
	cur, _ := q.Current()
	if cur.ID != "t1" {
		t.Fatalf("expected t1 got %s", cur.ID)
	}
	if _, err := q.Prev(); err != nil {
		t.Fatalf("prev err: %v", err)
	}
	cur, _ = q.Current()
	if cur.ID != "t0" {
		t.Fatalf("expected t0 got %s", cur.ID)
	}
}

func TestQueueRemove(t *testing.T) {
	q := New()
	q.Add(sampleTracks(3)...)
	if err := q.Remove(1); err != nil {
		t.Fatalf("remove err: %v", err)
	}
	if q.Len() != 2 {
		t.Fatalf("len after remove: %d", q.Len())
	}
	cur, _ := q.Current()
	if cur.ID != "t0" {
		t.Fatalf("expected current stay t0 got %s", cur.ID)
	}
	if err := q.Remove(0); err != nil {
		t.Fatalf("remove err: %v", err)
	}
	cur, _ = q.Current()
	if cur.ID != "t2" {
		t.Fatalf("expected current t2 got %s", cur.ID)
	}
}

func TestQueueMove(t *testing.T) {
	q := New()
	q.Add(sampleTracks(3)...)
	if err := q.Move(0, 2); err != nil {
		t.Fatalf("move err: %v", err)
	}
	cur, _ := q.Current()
	if cur.ID != "t0" {
		t.Fatalf("expected current t0 got %s", cur.ID)
	}
}
