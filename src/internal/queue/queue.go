package queue

import (
	"errors"
	"math/rand"

	"github.com/tunez/tunez/internal/provider"
)

type RepeatMode int

const (
	RepeatOff RepeatMode = iota
	RepeatAll
	RepeatOne
)

// Queue maintains an ordered list of tracks and the current position.
type Queue struct {
	items      []provider.Track
	current    int
	repeatMode RepeatMode
	shuffled   bool
	original   []provider.Track
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

func (q *Queue) CurrentIndex() int {
	return q.current
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

func (q *Queue) ToggleShuffle() {
	q.shuffled = !q.shuffled
	if q.shuffled {
		// Save original order
		q.original = make([]provider.Track, len(q.items))
		copy(q.original, q.items)

		currentTrack := provider.Track{}
		if q.current >= 0 && q.current < len(q.items) {
			currentTrack = q.items[q.current]
		}

		// Go 1.20+ auto-seeds; no need for rand.Seed
		rand.Shuffle(len(q.items), func(i, j int) {
			q.items[i], q.items[j] = q.items[j], q.items[i]
		})

		// Find current track and update index
		if currentTrack.ID != "" {
			for i, t := range q.items {
				if t.ID == currentTrack.ID {
					q.current = i
					break
				}
			}
		}
	} else {
		// Restore original order
		if q.original != nil {
			// We need to find where current track is in original
			currentTrack := provider.Track{}
			if q.current >= 0 && q.current < len(q.items) {
				currentTrack = q.items[q.current]
			}

			q.items = q.original
			q.original = nil

			if currentTrack.ID != "" {
				for i, t := range q.items {
					if t.ID == currentTrack.ID {
						q.current = i
						break
					}
				}
			}
		}
	}
}

func (q *Queue) CycleRepeat() RepeatMode {
	q.repeatMode = (q.repeatMode + 1) % 3
	return q.repeatMode
}

func (q *Queue) RepeatMode() RepeatMode {
	return q.repeatMode
}

func (q *Queue) IsShuffled() bool {
	return q.shuffled
}

func (q *Queue) Next() (provider.Track, error) {
	if len(q.items) == 0 {
		return provider.Track{}, ErrEmpty
	}

	if q.repeatMode == RepeatOne {
		if q.current == -1 {
			q.current = 0
		}
		return q.items[q.current], nil
	}

	if q.current < len(q.items)-1 {
		q.current++
	} else if q.repeatMode == RepeatAll {
		q.current = 0
	} else {
		// End of queue
		return provider.Track{}, errors.New("end of queue")
	}
	return q.items[q.current], nil
}

func (q *Queue) PeekNext() (provider.Track, error) {
	if len(q.items) == 0 {
		return provider.Track{}, ErrEmpty
	}
	nextIdx := q.current + 1
	if nextIdx >= len(q.items) {
		return provider.Track{}, errors.New("no next track")
	}
	return q.items[nextIdx], nil
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
