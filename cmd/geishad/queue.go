package main

import "sort"
import "math/rand"

type queue struct {
	loop   bool
	repeat bool
	curr   int
	q      []Song
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
		[]Song{},
	}
}

func (q *queue) len() int {
	return len(q.q)
}

func (q *queue) current() Song {
	q.next(0, false)
	if 0 <= q.curr && q.curr < len(q.q) {
		return q.q[q.curr]
	}
	return Song("")
}

func (q *queue) next(i int, force bool) {
	n := len(q.q)
	if n == 0 {
		return
	}
	if force {
		q.curr = mod(q.curr+i, n)
		return
	}
	if !q.repeat {
		c := q.curr + i
		if (0 <= c && c < n) || q.loop {
			q.curr = mod(c, n)
		} else {
			q.curr = c
		}
	}
}

func (q *queue) remove(i int) {
	if q.curr == len(q.q) {
		q.curr--
	}
	q.q = append(q.q[:i], q.q[i+1:]...)
}

func (q *queue) insert(i int, x Song) {
	if len(q.q) == 0 {
		q.q = append(q.q, x)
		return
	}
	q.q = append(q.q, Song(""))
	copy(q.q[i+1:], q.q[i:])
	q.q[i] = x
	if i < q.curr {
		q.curr++
	}
}

func (q *queue) append(x Song) {
	q.q = append(q.q, x)
}

func (q *queue) sort() {
	if len(q.q) == 0 {
		return
	}
	sort.Slice(q.q, func(i, j int) bool {
		swap := q.q[i] < q.q[j]
		if swap {
			if q.curr == i {
				q.curr = j
			} else if q.curr == j {
				q.curr = i
			}
		}
		return swap
	})
}

func (q *queue) shuffle() {
	if len(q.q) == 0 {
		return
	}
	rand.Shuffle(len(q.q), func(i, j int) {
		q.q[i], q.q[j] = q.q[j], q.q[i]
		if i == q.curr {
			q.curr = j
		} else if j == q.curr {
			q.curr = i
		}
	})
}
