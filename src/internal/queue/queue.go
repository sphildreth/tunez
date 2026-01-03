package queue

import (
	"errors"

	"github.com/tunez/tunez/internal/provider"
)

// Queue maintains an ordered list of tracks and the current position.
type Queue struct {
	items   []provider.Track
	current int
}

var ErrEmpty = errors.New("queue is empty")

func New() *Queue {
	return &Queue{items: []provider.Track{}, current: -1}
}

func (q *Queue) Items() []provider.Track {
	out := make([]provider.Track, len(q.items))
	copy(out, q.items)
	return out
}

func (q *Queue) Len() int { return len(q.items) }

func (q *Queue) Current() (provider.Track, error) {
	if q.current < 0 || q.current >= len(q.items) {
		return provider.Track{}, ErrEmpty
	}
	return q.items[q.current], nil
}

func (q *Queue) Add(tracks ...provider.Track) {
	q.items = append(q.items, tracks...)
	if q.current == -1 && len(q.items) > 0 {
		q.current = 0
	}
}

func (q *Queue) AddNext(track provider.Track) {
	if q.current == -1 {
		q.items = []provider.Track{track}
		q.current = 0
		return
	}
	idx := q.current + 1
	q.items = append(q.items[:idx], append([]provider.Track{track}, q.items[idx:]...)...)
}

func (q *Queue) Remove(idx int) error {
	if idx < 0 || idx >= len(q.items) {
		return errors.New("index out of range")
	}
	q.items = append(q.items[:idx], q.items[idx+1:]...)
	if len(q.items) == 0 {
		q.current = -1
		return nil
	}
	if idx < q.current {
		q.current--
	} else if idx == q.current && q.current >= len(q.items) {
		q.current = len(q.items) - 1
	}
	return nil
}

func (q *Queue) Move(from, to int) error {
	if from < 0 || from >= len(q.items) || to < 0 || to >= len(q.items) {
		return errors.New("index out of range")
	}
	if from == to {
		return nil
	}
	item := q.items[from]
	if from < to {
		copy(q.items[from:], q.items[from+1:to+1])
	} else {
		copy(q.items[to+1:], q.items[to:from])
	}
	q.items[to] = item
	if q.current == from {
		q.current = to
	} else if from < q.current && to >= q.current {
		q.current--
	} else if from > q.current && to <= q.current {
		q.current++
	}
	return nil
}

func (q *Queue) Next() (provider.Track, error) {
	if len(q.items) == 0 {
		return provider.Track{}, ErrEmpty
	}
	if q.current < len(q.items)-1 {
		q.current++
	}
	return q.items[q.current], nil
}

func (q *Queue) Prev() (provider.Track, error) {
	if len(q.items) == 0 {
		return provider.Track{}, ErrEmpty
	}
	if q.current > 0 {
		q.current--
	}
	return q.items[q.current], nil
}

func (q *Queue) SetCurrent(idx int) error {
	if idx < 0 || idx >= len(q.items) {
		return errors.New("index out of range")
	}
	q.current = idx
	return nil
}

func (q *Queue) Clear() {
	q.items = nil
	q.current = -1
}
