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

func TestQueueShuffle(t *testing.T) {
	q := New()
	q.Add(sampleTracks(5)...)
	if q.IsShuffled() {
		t.Fatal("queue should not be shuffled initially")
	}
	q.ToggleShuffle()
	if !q.IsShuffled() {
		t.Fatal("queue should be shuffled after toggle")
	}
	// Current track should still be accessible
	cur, err := q.Current()
	if err != nil {
		t.Fatalf("current err after shuffle: %v", err)
	}
	if cur.ID == "" {
		t.Fatal("current track should not be empty")
	}
	// Toggle back
	q.ToggleShuffle()
	if q.IsShuffled() {
		t.Fatal("queue should not be shuffled after second toggle")
	}
}

func TestQueueRepeat(t *testing.T) {
	q := New()
	q.Add(sampleTracks(2)...)
	if q.RepeatMode() != RepeatOff {
		t.Fatalf("expected RepeatOff, got %d", q.RepeatMode())
	}
	q.CycleRepeat()
	if q.RepeatMode() != RepeatAll {
		t.Fatalf("expected RepeatAll, got %d", q.RepeatMode())
	}
	q.CycleRepeat()
	if q.RepeatMode() != RepeatOne {
		t.Fatalf("expected RepeatOne, got %d", q.RepeatMode())
	}
	q.CycleRepeat()
	if q.RepeatMode() != RepeatOff {
		t.Fatalf("expected RepeatOff after cycling, got %d", q.RepeatMode())
	}
}

func TestQueueRepeatAll(t *testing.T) {
	q := New()
	q.Add(sampleTracks(2)...)
	q.CycleRepeat() // RepeatAll
	// Go to end
	q.Next()
	// Next should wrap to beginning
	track, err := q.Next()
	if err != nil {
		t.Fatalf("next with repeat all err: %v", err)
	}
	if track.ID != "t0" {
		t.Fatalf("expected wrap to t0, got %s", track.ID)
	}
}

func TestQueueRepeatOne(t *testing.T) {
	q := New()
	q.Add(sampleTracks(2)...)
	q.CycleRepeat() // RepeatAll
	q.CycleRepeat() // RepeatOne
	// Next should return same track
	track1, _ := q.Current()
	track2, err := q.Next()
	if err != nil {
		t.Fatalf("next with repeat one err: %v", err)
	}
	if track1.ID != track2.ID {
		t.Fatalf("expected same track in repeat one, got %s vs %s", track1.ID, track2.ID)
	}
}
