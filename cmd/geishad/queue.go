package main

import "sort"
import "math/rand"

type queueEntry struct {
	Id   int  `json:"id"`
	Song Song `json:"song"`
}

type queue struct {
	loop   bool
	repeat bool
	curr   int
	max_id int
	q      []*queueEntry
}

func mod(d, r int) int {
	m := d % r
	if m < 0 {
		m += r
	}
	return m
}

func newQueue(loop, repeat bool) *queue {
	return &queue{
		loop,
		repeat,
		0,
		0,
		[]*queueEntry{},
	}
}

func (q *queue) len() int {
	return len(q.q)
}

func (q *queue) current() *queueEntry {
	// need this line in case loop-mode changes in between calls to current() and next()
	q.next(0, false)
	if 0 <= q.curr && q.curr < len(q.q) {
		return q.q[q.curr]
	}
	return nil
}

func (q *queue) next(i int, force bool) {
	n := len(q.q)
	if n > 0 && (!q.repeat || force) {
		q.curr += i
		// only mod if we are in loop mode or are forced to.
		// so if q.curr goes out of bounds then we know we
		// have exhaused the queue.
		if q.loop || force {
			q.curr = mod(q.curr, n)
		}
	}
}

func (q *queue) find(x int) int {
	for i, c := range q.q {
		if c.Id == x {
			return i
		}
	}
	return -1
}

func (q *queue) remove(i int) {
	// so that future calls to next(1) wraps around properly
	if q.curr == len(q.q) || i < q.curr {
		q.curr--
	}
	q.q = append(q.q[:i], q.q[i+1:]...)
}

func (q *queue) insert(i int, x Song) {
	if len(q.q) == 0 {
		q.append(x)
		return
	}
	q.max_id++
	entry := &queueEntry{
		Song: x,
		Id:   q.max_id,
	}
	q.q = append(q.q, nil)
	copy(q.q[i+1:], q.q[i:])
	q.q[i] = entry
	if i < q.curr {
		q.curr++
	}
}

func (q *queue) append(x Song) {
	q.max_id++
	entry := &queueEntry{
		Song: x,
		Id:   q.max_id,
	}
	q.q = append(q.q, entry)
}

func (q *queue) sort() {
	curr := q.current()
	sort.Slice(q.q, func(i, j int) bool {
		swap := q.q[i].Id < q.q[j].Id
		if swap {
			if curr == q.q[i] {
				q.curr = j
			} else if curr == q.q[j] {
				q.curr = i
			}
		}
		return swap
	})
}

func (q *queue) shuffle() {
	rand.Shuffle(len(q.q), func(i, j int) {
		q.q[i], q.q[j] = q.q[j], q.q[i]
		if i == q.curr {
			q.curr = j
		} else if j == q.curr {
			q.curr = i
		}
	})
}
