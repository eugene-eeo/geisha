package main

import "sort"
import "math/rand"

type queue struct {
	curr int
	q    []Song
}

func newQueue() *queue {
	return &queue{
		curr: -1,
		q:    []Song{},
	}
}

func (q *queue) len() int {
	return len(q.q)
}

func (q *queue) next(repeat, loop bool) Song {
	n := len(q.q)
	if n == 0 || (!loop && q.curr == n) {
		return Song("")
	}
	if !repeat || q.curr == -1 {
		q.curr++
	}
	if loop {
		q.curr %= n
	}
	if q.curr == n {
		return Song("")
	}
	return q.q[q.curr]
}

func (q *queue) remove() {
	c := q.curr
	q.q = append(q.q[:c], q.q[c+1:]...)
}

func insert(q []Song, i int, x Song) []Song {
	q = append(q, Song(""))
	copy(q[i+1:], q[i:])
	q[i] = x
	return q
}

func (q *queue) append(song Song) {
	if q.len() == 0 || q.curr == 0 {
		q.q = append(q.q, song)
		return
	}
	q.q = insert(q.q, q.curr-1, song)
	q.curr++
}

func (q *queue) prepend(song Song) {
	if q.len() == 0 {
		q.q = append(q.q, song)
		return
	}
	q.q = insert(q.q, q.curr, song)
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
